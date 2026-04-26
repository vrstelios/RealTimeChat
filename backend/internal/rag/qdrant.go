package rag

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/qdrant/go-client/qdrant"
	"log"
)

const (
	collectionName = "documents"
	vectorSize     = 3072
)

var qdrantClient *qdrant.Client

func InitQdrant(host string, port int) error {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: host,
		Port: port,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to Qdrant: %w", err)
	}

	qdrantClient = client
	log.Println("Qdrant connected!")

	// Creat collection if not exist
	return ensureCollection(context.Background())
}

func ensureCollection(ctx context.Context) error {
	exists, err := qdrantClient.CollectionExists(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to check collection: %w", err)
	}

	if exists {
		log.Printf("Qdrant collection '%s' already exists\n", collectionName)
		return nil
	}

	// Creat new collection
	err = qdrantClient.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: collectionName,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     vectorSize,
			Distance: qdrant.Distance_Cosine,
		}),
	})
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	log.Printf("Qdrant collection '%s' created!\n", collectionName)
	return nil
}

func StoreChunks(ctx context.Context, chunks []Chunk, embeddings [][]float32) error {
	if len(chunks) != len(embeddings) {
		return fmt.Errorf("chunks and embeddings count mismatch: %d vs %d", len(chunks), len(embeddings))
	}

	points := make([]*qdrant.PointStruct, len(chunks))
	for i, chunk := range chunks {
		points[i] = &qdrant.PointStruct{
			Id:      qdrant.NewIDUUID(uuid.NewString()),
			Vectors: qdrant.NewVectorsDense(embeddings[i]),
			Payload: qdrant.NewValueMap(map[string]any{
				"text":     chunk.Text,
				"room":     chunk.Room,
				"filename": chunk.Filename,
				"index":    chunk.Index,
			}),
		}
	}

	_, err := qdrantClient.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: collectionName,
		Points:         points,
	})
	if err != nil {
		return fmt.Errorf("failed to store chunks: %w", err)
	}

	log.Printf("Stored %d chunks in Qdrant for room: %s\n", len(chunks), chunks[0].Room)
	return nil
}

// SearchChunks αναζητά τα πιο σχετικά chunks για μια ερώτηση
func SearchChunks(ctx context.Context, queryEmbedding []float32, room string, topK uint64) ([]string, error) {
	result, err := qdrantClient.Query(ctx, &qdrant.QueryPoints{
		CollectionName: collectionName,
		Query:          qdrant.NewQuery(queryEmbedding...),
		Limit:          qdrant.PtrOf(topK),
		WithPayload:    qdrant.NewWithPayload(true),
		// Filter only for specific room
		Filter: &qdrant.Filter{
			Must: []*qdrant.Condition{
				qdrant.NewMatch("room", room),
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	var texts []string
	for _, point := range result {
		if text, ok := point.Payload["text"]; ok {
			texts = append(texts, text.GetStringValue())
		}
	}

	return texts, nil
}

func DeleteRoomChunks(ctx context.Context, room string) error {
	_, err := qdrantClient.Delete(ctx, &qdrant.DeletePoints{
		CollectionName: collectionName,
		Points: qdrant.NewPointsSelectorFilter(&qdrant.Filter{
			Must: []*qdrant.Condition{
				qdrant.NewMatch("room", room),
			},
		}),
	})
	if err != nil {
		return fmt.Errorf("failed to delete chunks for room %s: %w", room, err)
	}

	log.Printf("Deleted all chunks for room: %s\n", room)
	return nil
}
