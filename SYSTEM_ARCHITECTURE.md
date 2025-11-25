# SwarmOS: Autonomous Enterprise Workforce Platform
## Architectural Blueprint v1.0

### 1. High-Level Design
**Pattern:** Event-Driven Microservices (Hexagonal Architecture)
**Communication:** gRPC (Inter-service), WebSocket (Client-Push), REST (Client-Pull)

### 2. Service Responsibilities

#### A. Swarm Core (Golang)
* **Role:** The Gateway & State Manager.
* **Responsibilities:**
    * AuthN/AuthZ (JWT + RBAC).
    * WebSocket Hub (Real-time updates to React Frontend).
    * Marketplace Transaction Management.
    * Proxy for "Swarm Brain".
* **Key Libraries:** `gin` (Router), `gorm` (ORM), `gorilla/websocket`.

#### B. Swarm Brain (Python)
* **Role:** The Intelligence Engine.
* **Responsibilities:**
    * LangGraph Orchestration (State Machines).
    * LLM Context Management (Ollama/Groq Interface).
    * Code Analysis & AST Parsing.
    * Docker Sandbox Management (for executing agent code).
* **Key Libraries:** `fastapi`, `langgraph`, `pydantic`, `docker`.

### 3. Data Flow (The "Neural Loop")
1.  **User Request:** Client sends prompt via WebSocket to `Swarm Core`.
2.  **Dispatch:** `Core` validates and pushes event `JOB_CREATED` to Redis.
3.  **Consumption:** `Swarm Brain` subscribes to `JOB_CREATED`.
4.  **Execution:** `Brain` spins up LangGraph agents.
5.  **Feedback:** Agents emit `THOUGHT` events.
6.  **Streaming:** `Brain` pushes `THOUGHT` to Redis Pub/Sub.
7.  **Delivery:** `Core` consumes Redis stream and pipes to specific User WebSocket.

### 4. Infrastructure
* **Persistence:** PostgreSQL (ACID transactions for billing/projects).
* **Cache/Bus:** Redis (Hot state & Event Bus).
* **Object Store:** MinIO (Artifact storage: generated code, logs).