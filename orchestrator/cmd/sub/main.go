package sub

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

func StatusListenerSub(js jetstream.JetStream, subject string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	InterruptCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	defer cancel()
	s, err := js.Stream(ctx, "myswarm")
	if err != nil {
		log.Fatal("error in finding the task stream", err)
	}
	c, err := s.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:       "CONS",
		FilterSubject: subject,
		AckPolicy:     jetstream.AckExplicitPolicy,
	})
	if err != nil {
		log.Fatal("error in creating consumer", err)
	}
	consumeCtx, err := c.Consume(func(msg jetstream.Msg) { // this will create a goroutine in the background so main thread is not blocked
		log.Println(string(msg.Data()))
		msg.Ack()
	})
	if err != nil {
		log.Fatal("error in consuming the messages")
	}
	<-InterruptCtx.Done()
	consumeCtx.Stop()
	consumeCtx.Closed()

}
