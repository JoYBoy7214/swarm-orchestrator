package sub

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go/jetstream"
)

type Incoming_msg struct {
	Msg_id uuid.UUID `json:"event_id"`
	Status string    `json:"status"`
}

func StatusListenerSub(js jetstream.JetStream, subject string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	InterruptCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	defer cancel()
	s, err := js.Stream(ctx, "myswarm")
	if err != nil {
		log.Fatal("error in finding the task stream", err)
	}
	c, err := s.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{ // this creates consumers which can used multiple client nats take care of load balancing and only one client will recieve the message at a time
		Durable:       "CONS",
		FilterSubject: subject,
		AckPolicy:     jetstream.AckExplicitPolicy,
	})
	if err != nil {
		log.Fatal("error in creating consumer", err)
	}
	consumeCtx, err := c.Consume(func(msg jetstream.Msg) { // this will create a goroutine in the background so main thread is not blocked
		var m Incoming_msg
		err := json.Unmarshal(msg.Data(), &m)
		if err != nil {
			log.Printf("error in decoding the incoming data. error :%v", err)
		}
		msg.Ack()
	})
	if err != nil {
		log.Fatal("error in consuming the messages")
	}
	<-InterruptCtx.Done()
	consumeCtx.Stop()
	consumeCtx.Closed()

}
