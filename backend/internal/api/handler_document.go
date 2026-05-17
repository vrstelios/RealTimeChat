package api

import (
	"RealTimeChat/backend/internal/database"
	"RealTimeChat/backend/internal/helpers"
	"RealTimeChat/backend/internal/metrics"
	"RealTimeChat/backend/internal/rag"
	"RealTimeChat/backend/internal/type/model"
	"encoding/json"
	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/genai"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type DocumentHandler struct {
	geminiClient *genai.Client
}

func NewDocumentHandler(client *genai.Client) *DocumentHandler {
	return &DocumentHandler{geminiClient: client}
}

// @Summary Upload a PDF document
// @Description Uploads a PDF file, processes it into chunks, generates embeddings and stores them per room
// @Tags documents
// @Accept multipart/form-data
// @Produce json
// @Param room query string true "Room name"
// @Param file formData file true "PDF file to upload"
// @Success 200 {object} model.DocumentResponse
// @Failure 400 {string} string
// @Failure 405 {string} string
// @Failure 500 {string} string
// @Router /api/documents/upload [post]
func (h *DocumentHandler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodPost {
		helpers.WriteJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	room := r.URL.Query().Get("room")
	if room == "" {
		helpers.WriteJSONError(w, http.StatusBadRequest, "room is required")
		return
	}

	// Limit request body size (32MB)
	r.Body = http.MaxBytesReader(w, r.Body, 32<<20)

	// Parse multipart form — max 32MB
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		helpers.WriteJSONError(w, http.StatusBadRequest, "Failed to parse multipart form: "+err.Error())
		return
	}

	// Get the uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		helpers.WriteJSONError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	// Validate file type (basic check)
	if filepath.Ext(header.Filename) != ".pdf" {
		helpers.WriteJSONError(w, http.StatusBadRequest, "only PDF files allowed")
		return
	}

	tmpFile, err := os.CreateTemp("", "upload-*.pdf")
	if err != nil {
		helpers.WriteJSONError(w, http.StatusInternalServerError, "failed to create temp file")
		return
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err = io.Copy(tmpFile, file); err != nil {
		helpers.WriteJSONError(w, http.StatusInternalServerError, "failed to save file")
		return
	}

	text, err := rag.ParsePDFToChunks(tmpFile.Name())
	if err != nil {
		helpers.WriteJSONError(w, http.StatusInternalServerError, "Failed to parse PDF")
		return
	}

	if text == "" {
		helpers.WriteJSONError(w, http.StatusBadRequest, "empty pdf")
		return
	}

	chunks := rag.ChunkText(text, room, header.Filename)
	if len(chunks) == 0 {
		helpers.WriteJSONError(w, http.StatusBadRequest, "No content extracted from PDF")
		return
	}

	log.Printf("file=%s chunks=%d room=%s", header.Filename, len(chunks), room)

	embeddings, err := rag.EmbedChunksWithRetry(ctx, h.geminiClient, chunks)
	if err != nil {
		//log.Printf("embedding error: %v", err)
		helpers.WriteJSONError(w, http.StatusInternalServerError, "Failed to generate embeddings")
		return
	}

	if err = rag.DeleteRoomChunks(ctx, room); err != nil {
		log.Printf("cleanup warning: %v", err)
	}

	if err = rag.StoreChunks(ctx, chunks, embeddings); err != nil {
		log.Printf("store error: %v", err)
		helpers.WriteJSONError(w, http.StatusInternalServerError, "Failed to store document chunks")
		return
	}

	if err = database.SaveDocument(room, header.Filename, len(chunks)); err != nil {
		log.Printf("mongo error: %v", err)
	}

	resp := model.DocumentResponse{
		Message: "document uploaded successfully",
		Data: model.Document{
			Id:          bson.NewObjectID(),
			Room:        room,
			File:        header.Filename,
			ChuckCount:  len(chunks),
			LastUpdated: time.Now(),
		},
	}

	metrics.DocumentsUploaded.WithLabelValues(room).Inc()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("response encode error: %v", err)
	}
}

// @Summary List documents by room
// @Description Retrieves all uploaded documents for a specific room
// @Tags documents
// @Produce json
// @Param room query string true "Room name"
// @Success 200 {array} model.Document
// @Failure 400 {string} string
// @Failure 405 {string} string
// @Failure 500 {string} string
// @Router /api/documents/ [get]
func (h *DocumentHandler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helpers.WriteJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	room := r.URL.Query().Get("room")
	if room == "" {
		helpers.WriteJSONError(w, http.StatusBadRequest, "room is required")
		return
	}

	docs, err := database.GetDocuments(room)
	if err != nil {
		//log.Println("MongoDB get documents error:", err)
		helpers.WriteJSONError(w, http.StatusInternalServerError, "Failed to retrieve documents")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(docs)

}
