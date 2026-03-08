package main

import (
	"context"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal("error in connecting nats", err)
	}
	js, err := jetstream.New(nc)
	if err != nil {
		log.Fatal("error in creating a jetStream ", err)
	}

	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     "task",
		Subjects: []string{"task.test"},
	})
	pAck, err := js.Publish(ctx, "task.test", []byte("hello from test"))
	if err != nil {
		log.Fatal("error in pusblishing message ", err)
	}
	log.Printf("Message published to %s. Sequence: %d", pAck.Stream, pAck.Sequence)

}
