package rag

import (
	"context"
	"fmt"
	"google.golang.org/genai"
	"time"
)

const embeddingModel = "gemini-embedding-001"

func EmbedChunks(ctx context.Context, client *genai.Client, chunks []Chunk) ([][]float32, error) {
	if len(chunks) == 0 {
		return nil, fmt.Errorf("no chunks to embed")
	}

	// Build contents for any chunk
	contents := make([]*genai.Content, len(chunks))
	for i, chunk := range chunks {
		contents[i] = genai.NewContentFromText(chunk.Text, genai.RoleUser)
	}

	result, err := client.Models.EmbedContent(
		ctx,
		embeddingModel,
		contents,
		&genai.EmbedContentConfig{
			TaskType: "RETRIEVAL_DOCUMENT",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("embedding failed: %w", err)
	}

	// Export float32 vectors
	embeddings := make([][]float32, len(result.Embeddings))
	for i, e := range result.Embeddings {
		embeddings[i] = e.Values
	}

	return embeddings, nil
}

// EmbedQuery create embedding for the question of user
func EmbedQuery(ctx context.Context, client *genai.Client, query string) ([]float32, error) {
	contents := []*genai.Content{
		genai.NewContentFromText(query, genai.RoleUser),
	}

	result, err := client.Models.EmbedContent(
		ctx,
		embeddingModel,
		contents,
		&genai.EmbedContentConfig{
			TaskType: "RETRIEVAL_QUERY",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("query embedding failed: %w", err)
	}

	if len(result.Embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	return result.Embeddings[0].Values, nil
}

// EmbedChunksWithRetry, retry if we have limit
func EmbedChunksWithRetry(ctx context.Context, client *genai.Client, chunks []Chunk) ([][]float32, error) {
	batchSize := 20
	var allEmbeddings [][]float32

	for i := 0; i < len(chunks); i += batchSize {
		end := i + batchSize
		if end > len(chunks) {
			end = len(chunks)
		}

		batch := chunks[i:end]
		embeddings, err := EmbedChunks(ctx, client, batch)
		if err != nil {
			return nil, fmt.Errorf("batch %d failed: %w", i/batchSize, err)
		}

		allEmbeddings = append(allEmbeddings, embeddings...)

		if end < len(chunks) {
			time.Sleep(500 * time.Millisecond)
		}
	}

	return allEmbeddings, nil
}
