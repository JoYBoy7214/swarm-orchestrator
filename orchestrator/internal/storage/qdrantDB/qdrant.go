package qdrantdb

import (
	"context"
	"fmt"
	"log"

	"github.com/qdrant/go-client/qdrant"
)

type QdrantStore struct {
	client *qdrant.Client
}

func CreateQdrantStore() (*QdrantStore, error) {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: "qdrant",
		Port: 6334,
	})

	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
		return nil, err
	}

	fmt.Println("Successfully connected to Qdrant!")

	return &QdrantStore{client: client}, nil
}

func (Q *QdrantStore) CreateNewCollection(ctx context.Context, collectionName string) error {
	err := Q.client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: collectionName,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     4,                      // Dimensionality of your vectors
			Distance: qdrant.Distance_Cosine, // Similarity metric
		}),
	})

	if err != nil {
		log.Fatalf("Failed to create collection: %v", err)
		return err
	}
	fmt.Println("Collection created.")
	return nil
}
