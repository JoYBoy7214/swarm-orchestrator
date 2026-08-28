package jet_stream

import (
	"context"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type JetStreamHandler struct {
	Jstream jetstream.JetStream
	nats    *nats.Conn
}

func CreateJetStreamHandler(natsUrl string) (*JetStreamHandler, error) {
	//ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//defer cancel()
	nc, err := nats.Connect(natsUrl) //nats.DefaultURL
	if err != nil {
		log.Println("error in connecting nats", err)
		return nil, err
	}
	js, err := jetstream.New(nc)
	if err != nil {
		log.Println("error in creating a jetStream ", err)
		return nil, err
	}
	log.Println("Js server started")
	return &JetStreamHandler{
		Jstream: js,
		nats:    nc,
	}, err
}

func CreateStream(ctx context.Context, streamName string, subject string) error {
	_, err := p.Jstream.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     streamName,
		Subjects: []string{subject},
	})
	if err != nil {
		return fmt.Errorf("Error in creating stream %s, Error :%w", streamName, err)
	}
	return err
}

func (p *JetStreamHandler) Publish(message []byte, ctx context.Context, subject string) error {
	pAck, err := p.Jstream.Publish(ctx, subject, message)
	if err != nil {
		log.Println("error in pusblishing message ", err)
		return err
	}
	log.Printf("Message published to %s. Sequence: %d", pAck.Stream, pAck.Sequence)
	return nil

}
