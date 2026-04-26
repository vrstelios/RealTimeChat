package rag

import (
	"fmt"
	"github.com/ledongthuc/pdf"
	"math"
	"strings"
)

type Chunk struct {
	Text     string
	Index    int
	Room     string
	Filename string
}

const (
	chunkSize    = 500
	chunkOverlap = 50
)

func ParsePDFToChunks(filepath string) (string, error) {
	f, r, err := pdf.Open(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF: %w", err)
	}
	defer f.Close()

	var sb strings.Builder
	totalPages := r.NumPage()

	for i := 1; i <= totalPages; i++ {
		page := r.Page(i)
		if page.V.IsNull() {
			continue
		}

		text, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		sb.WriteString(text)
		sb.WriteString(" ")
	}

	return strings.TrimSpace(sb.String()), nil
}

func ChunkText(text, room, filename string) []Chunk {
	text = strings.Join(strings.Fields(text), " ")
	if len(text) == 0 {
		return nil
	}

	totalChunks := int(math.Ceil(float64(len(text)) / float64(chunkSize-chunkOverlap)))
	chunks := make([]Chunk, 0, totalChunks)

	index := 0
	for start := 0; start < len(text); start += chunkSize - chunkOverlap {
		end := start + chunkSize
		if end > len(text) {
			end = len(text)
		}

		chunkText := strings.TrimSpace(text[start:end])
		if chunkText == "" {
			continue
		}

		chunks = append(chunks, Chunk{
			Text:     chunkText,
			Index:    index,
			Room:     room,
			Filename: filename,
		})
		index++

		if end == len(text) {
			break
		}
	}

	return chunks
}
