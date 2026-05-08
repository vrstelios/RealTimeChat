# Real-Time Chat

A production-grade, AI-powered real-time chat platform built with **Go**, **WebSockets**, and **Google Gemini**. Designed with horizontal scalability, persistent storage, vector search (RAG), and full observability in mind.

---

[![CI](https://github.com/vrstelios/RealTimeChat/actions/workflows/ci.yml/badge.svg)](https://github.com/vrstelios/RealTimeChat/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.24-blue)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

---

## Architectural Overview

![WebSocketDiagram.png](images/WebSocketDiagram.png)

---

### Features

**Real-Time Communication**
- WebSocket-based chat with multiple rooms
- Redis Pub/Sub for cross-server message broadcasting
- Horizontal scaling ready — multiple server instances share state via Redis
**AI Integration (Google Gemini)**
- Streaming responses token-by-token (typing effect)
- Function Calling / MCP pattern with two tools:
  - `search_web` — DuckDuckGo web search
  - `search_documents` — semantic search on uploaded PDFs
- Full chat history context sent to Gemini on every request
**RAG (Retrieval-Augmented Generation)**
- Admin uploads PDF per room via `POST /api/documents/upload`
- PDF is chunked, embedded via Gemini Embeddings API, and stored in Qdrant
- On AI queries, relevant chunks are retrieved and injected as context
**Persistence**
- MongoDB stores all messages and document metadata
- History is loaded and displayed when a user rejoins a room
- Indexes on `room + timestamp` for fast queries
**Observability**
- Prometheus metrics exposed at `/metrics`
- Grafana dashboards: active connections, messages/min, Gemini latency, AI success rate
- OpenTelemetry distributed tracing → Jaeger UI
  - Traces full message lifecycle: WebSocket → MongoDB → Gemini → Tools → Redis
**CI/CD**
- GitHub Actions on every push: `go build`, `go test`, `go vet`
---

## Tech Stack
 
| Layer | Technology |
|---|---|
| Language | Go 1.24 |
| WebSockets | gorilla/websocket |
| AI | Google Gemini API (google.golang.org/genai) |
| Message Broker | Redis (Pub/Sub) |
| Primary Database | MongoDB |
| Vector Database | Qdrant |
| Metrics | Prometheus + Grafana |
| Tracing | OpenTelemetry + Jaeger |
| CI | GitHub Actions |
| Infrastructure | Docker Compose |
 
---

## Project Structure

```
.
├── .github/workflows/           
│   └── cl.yml                        # CI/CD Orchestration
├── backend/                          # Backend logic and API handlers
│   ├── cmd/                 
│   │    └── api/main.go              # HTTP handlers and WebSocket endpoints
│   ├── config/                       # Configuration management
│   │    └── config.go                # Read .env, return Config struct  
│   ├── docs/                         # Swagger documentation
│   │    ├── docs.go                
│   │    ├── swagger.json   
│   │    └── swagger.yaml 
│   └── internal/
│        ├── api/         
│        │    ├── handler_chat.go     # WebSocket handlers for chat rooms
│        │    └── handler_document.go # API handlers for document upload and search
│        ├── database/         
│        │    ├── connMogno.go        # MongoDB connection and operations
│        │    └── crud.go             # CRUD operations for messages and rooms
│        ├── mcp/         
│        │    ├── search.go           # Web search tool for Gemini
│        │    └── tool.go             # MCP tool definitions
│        ├── metrics/   
│        │    └── metrics.go          # Prometheus metrics setup
│        ├── rag/         
│        │    ├── embeddings.go       # Vector DB interactions for RAG
│        │    ├── pdf.go              # PDF processing and chunking
│        │    └── qdrant.go           # Quadrant-based search logic
│        ├── server/         
│        │    ├── client.go           # WebSocket client logic and  Gemini AI integration
│        │    └── room.go             # Chat room management 
│        ├── tracing/   
│        │    └── tarcer.go           # Jaeger tracing setup
│        └── type/model               # Data models (Message, document, search) 
├── frontend/                         # Frontend assets and templates
│   ├── assets/                       # Static files (CSS, JS)
│   ├── index.html                    # Home page
│   └── chat.html                     # Chat UI 
├── images/ 
├── go.mod                            # Go dependencies and module definition
├── go.sum
├── .env
├── docker-chat.monitoring            # Infrastructure orchestration (mogno, redis, jaeger, prometheus, grafana)
├── prometheus.yaml                   # Metrics scraping configuration for Prometheus
├── go.mod                   
└── README.md                         # Project documentation and architecture overview
```

---

## Getting Started
 
### Prerequisites
 
- Go 1.24+
- Docker & Docker Compose
- Google Gemini API key → [ai.google.dev](https://ai.google.dev)
### 1. Clone the repository
 
```bash
git clone https://github.com/vrstelios/RealTimeChat.git
cd RealTimeChat
```
 
### 2. Configure environment
 
```bash
cp .env.example .env
```
 
Edit `.env`:
 
```env
APP_ADDR=:8080
GEMINI_API_KEY=your_gemini_api_key
 
REDIS_ADDR=localhost:6380
MONGO_URI=mongodb://admin:password@localhost:27017/?authSource=admin
QDRANT_HOST=localhost
QDRANT_PORT=6334
 
JAEGER_ENDPOINT=localhost:4318
```
 
### 3. Start infrastructure
 
```bash
docker compose -f docker-chat.monitoring.yml up -d
```
 
This starts:
 
| Service | URL |
|---|---|
| Redis | localhost:6380 |
| MongoDB | localhost:27017 |
| Qdrant | localhost:6333 |
| Prometheus | http://localhost:9090 |
| Grafana | http://localhost:3000 |
| Jaeger | http://localhost:16686 |
 
### 4. Run the application
 
```bash
go run cmd/web/main.go
```

The server starts at `http://localhost:8080`.

---

## Usage
 
### Join a room
 
Navigate to `http://localhost:8080` and enter a room name, username, and optionally enable AI mode.
 
Or directly:
```
http://localhost:8080/chat?room=room1&username=stelios&useAI=true
```
 
### Upload a PDF (AI rooms only)
 
```bash
curl -X POST "http://localhost:8080/api/documents/upload?room=room1" \
  -F "file=@document.pdf"
```
 
The PDF is chunked, embedded, and stored in Qdrant. The AI will now answer questions based on its content.
 
### List documents for a room
 
```bash
curl "http://localhost:8080/api/documents?room=room1"
```

---
 
## API Reference
 
| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/` | Home page |
| `GET` | `/room` | WebSocket upgrade endpoint |
| `POST` | `/api/documents/upload?room=` | Upload & embed a PDF |
| `GET` | `/api/documents?room=` | List uploaded documents |
| `GET` | `/metrics` | Prometheus metrics |
 
### WebSocket Query Parameters
 
| Parameter | Type | Description |
|---|---|---|
| `room` | string | Room name |
| `name` | string | Username |
| `useAI` | bool | Enable Gemini AI |
 
---

## Observability
 
### Grafana Dashboards
 
Import `grafana-dashboard.json` into Grafana (`http://localhost:3000`).
 
Available panels:
- Active WebSocket connections
- Messages per minute (user vs Gemini)
- Gemini P95 / P50 latency
- AI requests — success vs error
- Documents uploaded per room
- Redis publish errors
- WebSocket errors
### Jaeger Tracing
 
Open `http://localhost:16686`, select service `realtimechat`.
 
Each AI request produces a trace:
![Metrics.png](images/Metrics_1.png)
![Metrics.png](images/Metrics_2.png)
```
streamGemini (chat)                          
├── load_history_mongodb              1.95ms
├── save_to_mongodb                   0ms
└── redis_broadcast                   1.38ms
streamGemini (UploadDocument)
├── load_history_mongodb              28.53ms
├── tool_call_search_documents        358.65m
├── save_to_mongodb                   0ms
└── redis_broadcast                   1.4ms
```

---

## How It Works
 
### Message Flow (no AI)
 
```
Client sends message
    → room.forward channel
    → Redis Publish (room:<name>)
    → subscribeRedis goroutine
    → broadcast to all local clients
    → saved to MongoDB
```
 
### Message Flow (with AI)
 
```
Client sends message
    → broadcast to room (above)
    → streamGemini() goroutine
        → load history from MongoDB
        → Gemini GenerateContent with tools
        → if tool call: execute search_web or search_documents
        → stream tokens to sender (receive channel)
        → save user + model messages to MongoDB
        → publish final answer to Redis for others
```
 
### RAG Flow
 
```
Admin uploads PDF
    → parse text → chunk (500 chars, 50 overlap)
    → Gemini Embeddings API (RETRIEVAL_DOCUMENT)
    → store vectors in Qdrant with room metadata
 
User asks question
    → Gemini decides to call search_documents
    → embed query (RETRIEVAL_QUERY)
    → Qdrant cosine similarity search
    → top-K chunks injected as context
    → Gemini answers based on document
```
 
---

### Author
[DoctorVerRossi](https://github.com/vrstelios)

---

If you find this project helpful, please give it a star on GitHub!

