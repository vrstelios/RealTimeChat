# Real-Time Chat Platform

A production-grade, AI-powered real-time chat platform engineered with **Go**, **WebSockets**, and **Google Gemini**. Designed from scratch with a focus on horizontal scalability, distributed state management, vector-based Retrieval-Augmented Generation (RAG), high concurrency safety, and deep production observability.

> **Project Inspiration:** Based on the Architectural [Broadcast Server](https://roadmap.sh/projects/broadcast-server) specification from roadmap.sh, heavily extended into an enterprise-level distributed AI orchestration platform.

---

[![CI Execution](https://github.com/vrstelios/RealTimeChat/actions/workflows/ci.yml/badge.svg)](https://github.com/vrstelios/RealTimeChat/actions/workflows/ci.yml)
[![Go Architecture](https://img.shields.io/badge/Go-1.24-blue?logo=go)](https://golang.org)
[![Infrastructure](https://img.shields.io/badge/Infrastructure-Docker--Compose-red?logo=docker)](https://www.docker.com/)
[![License: MIT](https://img.shields.io/badge/license-MIT-green)](LICENSE)

---

## Architectural Overview

![WebSocketDiagram.png](images/WebSocketDiagram.png)

### Key Architectural Pillars

* **Distributed State & Real-Time Broadcast:** Leverages low-latency WebSockets utilizing `gorilla/websocket`. Multi-node horizontal scaling is achieved via a **Redis Pub/Sub** mesh backbone, decoupling individual cluster node termination from room broadcast states.
* **Production Concurrency Safety:** Employs advanced custom synchronization safeguards (`safeSend`) preventing write-on-closed-channel panics during volatile, high-throughput user connection drops.
* **Asynchronous AI Tool Extraction (Gemini & MCP):** Implements Model Context Protocol (MCP) pattern utilizing functional primitives for dynamic `search_web` (via DuckDuckGo) and semantic `search_documents` processing.
* **Retrieval-Augmented Generation (RAG):** Implements isolated multi-tenant semantic document vector injections. PDFs are computationally split into overlapping chunks, vectorized via the Gemini Embeddings engine (`RETRIEVAL_DOCUMENT`), and retained within **Qdrant Vector Database** with automated hardware room-isolation filtering.
* **Structured Resilience & Graceful Disconnection:** Captures `SIGINT`/`SIGTERM` POSIX signals to drain the HTTP connection pool, push WebSocket closure control frames (`1001 CloseGoingAway`) to active clients, and sequentially flush context keys out of Redis memory state stores to eliminate phantom data pollution.

---

## Tech Stack

| Layer | Component | Technology |
| :--- | :--- | :--- |
| **Runtime / Engineering** | Core Language | Go 1.24 (Highly optimized concurrency models) |
| **Real-Time Network** | WebSocket Protocol | `gorilla/websocket` |
| **AI LLM Engine** | Inference Framework | Google Gemini API Engine (`google.golang.org/genai`) |
| **Distributed Mesh** | Message Broker / PubSub | Redis Memory Cluster |
| **Persistent Ledger** | Relational Document DB | MongoDB Server |
| **Vector Indexing** | Semantic Embeddings Engine | Qdrant Vector Search Engine |
| **Telemetry System** | Time-Series Metrics | Prometheus Engine |
| **Analytical Board** | Visualization Suite | Grafana Labs Dashboard Systems |
| **Distributed Tracing** | Application APM Spans | OpenTelemetry Core + Jaeger Engine backend |
| **Pipeline Automation** | CI/CD Infrastructure | GitHub Actions Workflow engine |

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
│        │    ├── handler_document.go # API handlers for document upload and search
│        │    ├── handler_me.go       # API handlers for user authentication (signup/login)
│        │    └── handler_user.go     # API handlers for user profile management
│        ├── database/         
│        │    ├── connMogno.go        # MongoDB connection and operations
│        │    └── crud.go             # CRUD operations for messages and rooms
│        ├── helpers/         
│        │    ├── errors.go           # Custom error types and handling logic
│        │    └── token.go            # Include tools for password/token
│        ├── mcp/         
│        │    ├── search.go           # Web search tool for Gemini
│        │    └── tool.go             # MCP tool definitions
│        ├── metrics/   
│        │    └── metrics.go          # Prometheus metrics setup
│        ├── middleware/         
│        │    ├── auth.go             # Custom error types and handling logic
│        │    ├── metrics.go          # Prometheus metrics middleware for HTTP handlers
│        │    └── token.go            # Token provider
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
│   ├── login.html                    # User login page
│   ├── signup.html                   # User signup page
│   └── chat.html                     # Chat UI 
├── images/ 
├── test/ 
│   ├── package.json                  
│   ├── package-lock.json    
│   ├── TestRealTimeChat.pdf          # Sample PDF for RAG testing            
│   └── performance-script.js         # Load testing script for performance validation
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

* Go 1.24+ Compiler Toolchain
* Docker Engine Daemon & Compose V2 CLI Suite
* Valid Google Gemini API Key Client Identifier ([ai.google.dev](https://ai.google.dev))

### Installation & Deployment Execution

1. **Clone Repo Pipeline**
   ```bash
   git clone [https://github.com/vrstelios/RealTimeChat.git](https://github.com/vrstelios/RealTimeChat.git)
   cd RealTimeChat
 
### 2. Configure environment
 
```bash
cp .env.example .env
```
 
Configure `.env` using your required keys:
 
```env
APP_ADDR=:8080
GEMINI_API_KEY=your_gemini_api_key
 
REDIS_ADDR=localhost:6380
MONGO_URI=mongodb://admin:password@localhost:27017/?authSource=admin
QDRANT_HOST=localhost
QDRANT_PORT=6334
 
JAEGER_ENDPOINT=localhost:4318
```
 
### 3. Orchestrate Infrastructure Matrix
 
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
 
### 4. Launch RealTimeChat Engine
 
```bash
go run backend/cmd/api/main.go
```

The server starts at `http://localhost:8080`.

---

### Empirical Verification: Senior-Level Load Test Performance

To evaluate the operational stability, structural concurrency, and thread safety of the core architecture, a automated load testing framework was executed. The test validates 4 multi-layered scenarios under aggressive traffic spikes.

### Execution Metrics Output Log
```
═══════════════════════════════════════════════════════
  ⬡  RealTimeChat Load Test
  Senior-level system demonstration
═══════════════════════════════════════════════════════

[0.0s]    SCENARIO 1           Multi-user chat load test
[0.7s]    Users Created        10/10 ready
[0.8s]    WS Connected         user-onif7b → room-load-4
[0.8s]    WS Connected         user-nno6fh → room-load-3
...
[6.9s]    Message Sent         user-jmtjaf → [room-load-5]: "Load test message 10 — bipe5qc5xg"
[9.5s]    SCENARIO 1           Complete

[11.5s]   SCENARIO 2           AI interaction test
[11.5s]   WS Connected         user-srr42t → room-ai-load
[12.0s]   AI Question          What is a WebSocket?
[19.7s]   AI Response          [room-ai-load] A **WebSocket** is a computer communications protocol...
[20.0s]   AI Question          How does Redis Pub/Sub work?
[28.0s]   AI Question          Explain horizontal scaling
[28.5s]   AI Response          [room-ai-load] These two technologies are often used together to build...
[36.0s]   SCENARIO 2           Complete

[38.0s]   SCENARIO 3           PDF Upload + RAG query test
[38.1s]   PDF Path             C:\Users\User\GolandProjects\RealTimeChat\test\TestRealTimeChat.pdf
[38.1s]   PDF Exists           true
[38.5s]   PDF Upload           200: {"message":"document uploaded successfully","data":{"Id":"6a104609ea92f2d5dd0e1750","Room":"room-rag-load","File":"TestRealTimeChat.pdf","ChunkCount":14}}
[40.5s]   WS Connected         user-jddiew → room-rag-load
[41.1s]   RAG Query            Asking about document content...
[41.2s]   AI Response          [room-rag-load] ...
[51.1s]   SCENARIO 3           Complete

[53.1s]   SCENARIO 4           Room isolation test
[53.1s]   WS Connected         user-rm7jt5 → isolated-room-B
[53.1s]   WS Connected         user-jddiew → isolated-room-A
[55.6s]   Room Isolation       PASS — Rooms are properly isolated
[55.6s]   SCENARIO 4           Complete

═══════════════════════════════════════════════════════
LOAD TEST RESULTS
═══════════════════════════════════════════════════════
Duration: 57.6s
───────────────────────────────────────────────────────
AUTH
Signups: 10 | Failed: 0
Logins:  10 | Failed: 0
───────────────────────────────────────────────────────
WEBSOCKET
Connected: 14 | Errors: 0
Messages Sent: 100
Messages Recv: 2649
───────────────────────────────────────────────────────
AI (GEMINI)
Requests: 4 | Success: 3 | Failed: 0
───────────────────────────────────────────────────────
DOCUMENTS (RAG)
Uploads: 1 | Failed: 0
───────────────────────────────────────────────────────
Throughput: 1.7 msg/s
═══════════════════════════════════════════════════════
```

### Production Observability & Analytical Telemetry

## Real-Time Prometheus Insights (Grafana Dashboards)

The platform exposes deep internal runtime states to Prometheus scraped at a /metrics interface. The analytical graphs demonstrate optimal system responses under concurrent pressure loops:

* **Scenario 1 Metrics — Traffic Throughput & Active Room Density:** Tracks client broadcast waves alongside concurrent active chat room distributions.
![MessagesPerMinute.png](images/MessagesPerMinute.png)
* **Scenario 1 Metrics — Authentication Request Velocity & Ledger Mutations:** Validates secure JWT creation intervals alongside successful database ingestion.
![UserSignups&Logins.png](images/UserSignups&Logins.png)
* **Scenario 2 Metrics — Inference Pipeline Latency & Success Ratios:** Monitors processing time window distributions ($P_{50}$ / $P_{95}$) of the Gemini Engine layer and prompt success rates.
![GeminiP95Latency.png](images/GeminiP95Latency.png)
![AIRequests.png](images/AIRequests.png)

---

### Distributed APM Tracing (Jaeger Infrastructure)

Distributed transaction contexts are mapped across system bounds via OpenTelemetry collectors. This isolates bottlenecks across network boundaries (WebSocket $\rightarrow$ MongoDB $\rightarrow$ Vector Search $\rightarrow$ LLM Inference Engine):

## Scenario 3 APM Breakdown: Real-time semantic analysis during the vector search fetch flow.

Open `http://localhost:16686`, select service `realtimechat`.
 
Each AI request produces a trace:
![streamGeminiTrace.png](images/streamGeminiTrace.png)
![Metrics.png](images/Metrics_2.png)
```
streamGemini (chat)                          
├── load_history_mongodb              1.69ms
├── save_to_mongodb                   0ms
└── redis_broadcast                   2.22ms
streamGemini (UploadDocument)
├── load_history_mongodb              28.53ms
├── tool_call_search_documents        358.65m
├── save_to_mongodb                   0ms
└── redis_broadcast                   1.4ms
```

---

### Swagger Documentation

The API is fully documented using Swagger/OpenAPI 3.0. Once the server is running, you can access:
- Swagger UI: `http://localhost:8080/swagger/index.html`
- OpenAPI JSON: http://localhost:8080/swagger/doc.json
- OpenAPI YAML: Available in docs/swagger.yaml

## API Reference

![swagger.png](images/swagger.png)

---
### Author
[DoctorVerRossi](https://github.com/vrstelios)

---

If you find this project helpful, please give it a star on GitHub!

