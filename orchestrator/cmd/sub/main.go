package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	InterruptCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	defer cancel()
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal("error in connecting nats", err)
	}
	js, err := jetstream.New(nc)
	s, err := js.Stream(ctx, "myswarm")
	if err != nil {
		log.Fatal("error in finding the task stream", err)
	}
	c, err := s.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:       "CONS",
		FilterSubject: "task.>",
		AckPolicy:     jetstream.AckExplicitPolicy,
	})
	if err != nil {
		log.Fatal("error in creating consumer", err)
	}
	consumeCtx, err := c.Consume(func(msg jetstream.Msg) {
		log.Println(string(msg.Data()))
		msg.Ack()
	})
	c.Consume(func(msg jetstream.Msg) {
		fmt.Println("Worker 2 handled it")
		msg.Ack()
	})

	<-InterruptCtx.Done()
	consumeCtx.Stop()
	consumeCtx.Closed()

}
