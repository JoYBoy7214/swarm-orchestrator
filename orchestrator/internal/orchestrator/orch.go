package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JoYBoy7214/swarm-orchestrator/internal/jet_stream"
	"github.com/JoYBoy7214/swarm-orchestrator/internal/storage/postgresDb"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go/jetstream"
)

type Orchestrator struct {
	DbDriver      *postgresDb.PostgresDriver
	StreamHandler *jet_stream.JetStreamHandler
}

type tempschema struct {
	WorkFlow_id uuid.UUID `json:"Workflow_id"`
	Task_id     uuid.UUID `json:"Task_id"`
	Task_type   string    `json:"Task_type"`
}

func CreateOrchestrator(ctx context.Context, DbString string, natsUrl string) (*Orchestrator, error) {
	DbDriver, err := postgresDb.CreatePostgresDriver(ctx, DbString)
	if err != nil {
		return nil, fmt.Errorf("Error in creating Db Driver, Error: %w", err)
	}
	StreamHandler, err := jet_stream.CreateJetStreamHandler(ctx, natsUrl)
	if err != nil {
		return nil, fmt.Errorf("Error in creating Stream hnadler, Error: %w", err)
	}
	return &Orchestrator{
		DbDriver:      DbDriver,
		StreamHandler: StreamHandler,
	}, nil
}

func (orch *Orchestrator) StartOrchestrating() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	InterruptCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	consumer, err := orch.StreamHandler.Jstream.CreateOrUpdateConsumer(ctx, "TASKS", jetstream.ConsumerConfig{
		Durable:       "ORCH_CONS",
		FilterSubject: "task.COMPLETED",
		AckPolicy:     jetstream.AckExplicitPolicy,
	})
	if err != nil {
		log.Fatal("Error in creating consumers")
	}
	go orch.pullConsumer(ctx, consumer)
	<-InterruptCtx.Done()
	cancel()

}
func (orch *Orchestrator) pullConsumer(ctx context.Context, consumer jetstream.Consumer) {
	//we will evalute the DAG in batch so we can reduce the DB Calls why this works(according to me) let say we have
	// a graph of 500 nodes evaluating every time a task completed will inefficient and number of the ready task also will be less
	//so evaluting after 100 task completed or 1 sec could lead to more ready tasks
	for {
		bulk, err := consumer.Fetch(100, jetstream.FetchMaxWait(1*time.Second))
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) { // in case the queue is not full by one second
				log.Println("Deadline reached before filling the queue %w", err)
			} else {
				log.Fatal("Error in fetching message from stream %w", err)
				return
			}
		}
		for m := range bulk.Messages() {
			var msg tempschema
			err = json.Unmarshal(m.Data(), &msg)
			if err != nil {
				log.Println("Error in converting jetstream.Msg to tempschema %w", err)
				continue
			}

			updateContext, updatecontextCancel := context.WithTimeout(ctx, time.Second)
			defer updatecontextCancel()
			workflow_id := msg.WorkFlow_id //can be further optimized by evaluating one workflow only one time
			task_id := msg.Task_id
			err = orch.DbDriver.UpdateTask(updateContext, task_id, workflow_id, "COMPLETED")
			if err != nil {
				log.Println("Error in updating task_status %w:", err)
				continue
			}

			driverContext, driverContextCancel := context.WithTimeout(ctx, time.Second) //check how to close this gracefully
			defer driverContextCancel()
			ready_task, err := orch.DbDriver.Evaluate(driverContext, workflow_id)
			if err != nil {
				log.Println("Error in Evaluating DAG %w", err)
				continue
			}

			publisherContext, publisherContextCancel := context.WithTimeout(ctx, time.Second)
			defer publisherContextCancel()
			err = orch.bulkPublisher(publisherContext, ready_task, workflow_id, msg.Task_type)
			if err != nil {
				log.Println("Error in publishing the tasks %w", err)
				continue
			}

			err = m.Ack()
			if err != nil {
				log.Println("Error in ack the message %w", err)
				continue
			}
		}
	}
}

func (orch *Orchestrator) bulkPublisher(ctx context.Context, tasks []uuid.UUID, workflow_id uuid.UUID, task_type string) error {
	for _, task := range tasks {
		msg := tempschema{
			WorkFlow_id: workflow_id,
			Task_id:     task,
			Task_type:   task_type,
		}
		msgByte, err := json.Marshal(msg)
		err = orch.StreamHandler.Publish(ctx, msgByte, "tasks.EXCUTE")
		if err != nil {
			return err
		}
	}
	return nil
}

func (orch *Orchestrator) RootNodeUpdater(ctx context.Context, workflow_id uuid.UUID) []uuid.UUID {
	//This fucntion is to update the status of the root nodes of DAG since func StartOrchestrating() only relies on task.Completed it won't able to identify and
	//trigger the Root ndoes so every time when we create a DAG we have to run this func
	driverContext, driverContextCancel := context.WithTimeout(ctx, time.Second)
	defer driverContextCancel()
}
