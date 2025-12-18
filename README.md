# AgentMesh

**AgentMesh** is a high-performance distributed task orchestrator designed to manage thousands of concurrent AI agents with extreme reliability. Built with Go, Redis, and PostgreSQL, it decouples task ingestion (Producer) from execution (Workers) to ensure scalability and fault tolerance.

Typical Use Case: Orchestrating a fleet of specialized AI agents (e.g., Magnus, Cedric, Lyra) to process complex workflows in real-time.

---

## 🚀 Architecture

AgentMesh employs a **Producer-Broker-Worker** architecture with a focus on data integrity and high availability.

```mermaid
flowchart LR
    Client([Client Request]) -->|POST /tasks| Producer[Producer API]
    
    subgraph Core Engine
        Producer -->|Store Pending| DB[(PostgreSQL)]
        Producer -->|Enqueue Job| Redis[(Redis Queue)]
        
        Redis -->|BRPOP Task| Worker[Worker Service]
        Worker -->|Update Status| DB
    end
    
    Worker -->|Execute AI Agent| AgentLogic{AI Processing}
    AgentLogic -->|Success| Complete[Mark Completed]
    AgentLogic -->|Fail| Retry[Retry / DLQ]
```

### High-Level Data Flow

1. **Producer (API)**: Receives task requests, validates them, persists them to PostgreSQL as `PENDING`, and enqueues the Task ID into Redis.
2. **Broker (Redis)**: Manages Priority Queues (`agent_high`, `agent_medium`, `agent_low`) to ensure critical tasks are processed first.
3. **Worker (Consumer)**:
    * **Fetch**: Blocks via `BRPOP` until a task is available.
    * **Claim Pattern**: *Immediately* updates the task status in PostgreSQL to `RUNNING` upon retrieval to prevent race conditions or zombie tasks.
    * **Execute**: Simulates AI task processing (e.g., LLM inference).
    * **Result**: Updates PostgreSQL with `COMPLETED` or handling failure.

### Reliability Mechanisms

* **Exponential Backoff**: Transient failures trigger a retry mechanism with a wait time of $2^n$ seconds.
* **Dead Letter Queue (DLQ)**: Tasks failing more than 5 times are moved to a `agent_dead_letter` Redis queue and marked as `PERMANENT_FAILURE` in the database for manual inspection.
* **Concurrency**: Workers run on multiple goroutines, allowing a single binary to process multiple tasks in parallel.
* **ACID Persistence**: All state transitions are logged in PostgreSQL (pgx/v5) for auditability.

---

## 🛠️ Technology Stack

* **Language**: Go (Golang) 1.22+
* **Broker**: Redis (using `go-redis/v9`)
* **Database**: PostgreSQL 16 (using `jackc/pgx/v5` for connection pooling)
* **Infrastructure**: Docker & Docker Compose

---

## Project Structure

```bash
├── cmd
│   ├── producer          # API Entrypoint (Task Ingestion)
│   └── worker            # Worker Entrypoint (Task Execution)
├── internal
│   ├── config            # Twelve-Factor App Configuration
│   ├── models            # Domain Models & Constants (Agent Types)
│   └── worker            # Core Worker Logic (Backoff, DLQ, Loop)
├── pkg
│   ├── broker            # Redis Broker Implementation (Priority Queues)
│   └── database          # PostgreSQL Layer (Queries & Connection Pool)
├── migrations            # SQL Schema Definitions
└── docker-compose.yml    # Infrastructure Setup
```

---

## Supported Agents (The "Rankup" Squad)

AgentMesh is pre-configured to orchestrate specific agent personas:

* **MAGNUS_STRATEGIST** (`agent_high`): Handles high-priority strategic planning tasks.
* **CEDRIC_WRITER** (`agent_medium`): Handles content generation and copywriting.
* **LYRA_AUDITOR** (`agent_low`): Handles background verification and data auditing.

---

## Getting Started

### Prerequisites

* Docker & Docker Compose
* Go 1.22+

### 1. Initialize Infrastructure

Start Redis and PostgreSQL containers:

```bash
docker-compose up -d
```

### 2. Apply Migrations

(Currently manual or via a future migration tool. For dev, you can run the SQL directly):

```bash
# Connect to Postgres container and run SQL
docker exec -i agentmesh_postgres psql -U user -d agentmesh < migrations/001_init_schema.sql
docker exec -i agentmesh_postgres psql -U user -d agentmesh < migrations/002_add_retry_count.sql
```

### 3. Start the Worker (Consumer)

The worker will start 5 concurrent goroutines and listen for tasks.

```bash
go run cmd/worker/main.go
# Output:
# Starting Worker Service...
# Connected to DB
# Connected to Redis at: localhost:6379
# Starting 5 concurrent workers...
```

### 4. Start the Producer (API)

(Note: API implementation is in progress. Currently initializes connection).

```bash
go run cmd/producer/main.go
```

---

## Configuration

AgentMesh follows the **Twelve-Factor App** methodology. Configure via Environment Variables:

| Variable | Description | Default |
| :--- | :--- | :--- |
| `REDIS_ADDR` | Redis connection string | `localhost:6379` |
| `DB_DSN` | PostgreSQL connection string | `postgres://user:password@localhost:5432/agentmesh...` |

---

## Internal Mechanics Highlight

### The "Claim" Pattern

To ensure robust distributed locking without complex mutexes, the Worker performs a robust fetch-and-claim:

```go
// Fetch from Redis (Blocking)
taskID := redis.BRPop(queues...)

// Atomic Status Update (PostgreSQL)
db.UpdateTaskStatus(taskID, "running") // This creates the audit trail
```

### Failure Handling & DLQ

If `simulateAIWork()` returns an error:

1. **Increment Retry**: `UPDATE tasks SET retry_count = retry_count + 1`
2. **Check Threshold**: If `retry_count > 5`:
    * Push to `agent_dead_letter` (Redis)
    * Set status to `PERMANENT_FAILURE` (DB)
3. **Backoff**: If `retry_count <= 5`:
    * Sleep `pow(2, retry_count) * second`
    * Re-enqueue to original priority queue.

---
*Built with ❤️ by [YehiaGewily](https://github.com/YehiaGewily) for the Agentic Future.*
