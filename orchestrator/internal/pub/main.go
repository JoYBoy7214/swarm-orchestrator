package pub

import (
	"context"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type Publisher struct {
	Jstream jetstream.JetStream
}

func CreateStream() (*Publisher, *nats.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Println("error in connecting nats", err)
		return nil, nil, err
	}
	js, err := jetstream.New(nc)
	if err != nil {
		log.Println("error in creating a jetStream ", err)
		return nil, nil, err
	}
	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     "myswarm",
		Subjects: []string{"task.>"},
	})
	log.Println("Js server started")
	return &Publisher{
		Jstream: js,
	}, nc, err
}

func (p *Publisher) Publish(message []byte, ctx context.Context) error {
	pAck, err := p.Jstream.Publish(ctx, "task.test", message)
	if err != nil {
		log.Println("error in pusblishing message ", err)
		return err
	}
	log.Printf("Message published to %s. Sequence: %d", pAck.Stream, pAck.Sequence)
	return nil

}
