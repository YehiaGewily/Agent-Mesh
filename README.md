# Agent Mesh: High-Frequency Distributed Task Orchestration

![Go](https://img.shields.io/badge/Backend-Go_1.23-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![React](https://img.shields.io/badge/Frontend-React_18-61DAFB?style=for-the-badge&logo=react&logoColor=black)
![Redis](https://img.shields.io/badge/Broker-Redis-DC382D?style=for-the-badge&logo=redis&logoColor=white)
![Docker](https://img.shields.io/badge/Deploy-Docker_Compose-2496ED?style=for-the-badge&logo=docker&logoColor=white)

**Agent Mesh** is a high-performance, event-driven orchestration system designed to manage complex AI agent workflows at scale. Built with a focus on **low-latency state synchronization** and **horizontal scalability**, it simulates a complete Software Development Lifecycle (SDLC) pipeline using specialized worker nodes.

> **Performance Benchmark**: Capable of handling **750+ requests per second** with <200ms end-to-end latency on standard hardware.

---

## Engineering Highlights

### 1. Zero-Lag Responsive Dashboard

The frontend "Command Center" is engineered to remain responsive under extreme load:

* **Throttled State Batching**: Incoming WebSocket events are buffered and flushed at a stable **10 FPS (100ms interval)**, decoupling network throughput from render cycles.
* **Visual Virtualization**: Task lists are intelligently capped (Top 20 Pending / Last 15 History) to maintain a low DOM node count, ensuring a locked **60 FPS** even with thousands of queued tasks.
* **High Traffic Mode**: Automatic congestion detection triggers a visual alert when the event buffer exceeds 500 msgs/sec.

### 2. Process-Level Telemetry

Unlike generic monitoring, Agent Mesh tracks **Process Resident Memory (RSS)**.

* Workers report exact memory footprint (in MB) relative to a soft-limit of **512MB**.
* This provides precise, noise-free health metrics that reflect the actual application state, not system background noise.

### 3. Distributed Event Bus

* **Decoupled Architecture**: Producers and Workers never communicate directly. All coordination happens via **Redis Pub/Sub** and **Atomic Lists**.
* **Reliability**: Features **Dead Letter Queues (DLQ)** and **Exponential Backoff** for resilient error handling.

---

## System Architecture

The system follows a reactive, microservices-based pattern:

```mermaid
flowchart LR
    Client([React Command Center]) <-->|WS / HTTP| Producer[Producer Service]
    
    subgraph "Orchestration Layer"
        Producer -->|Enqueue Task| Redis[(Redis Queue)]
        Redis -->|Pub/Sub Events| Producer
    end
    
    subgraph "Execution Layer"
        Worker[Go Worker Nodes] <-->|Pop Task| Redis
        Worker -->|Persist State| DB[(PostgreSQL)]
        Worker -->|Broadcast Health| Redis
    end
```

### The "Software Squad" Simulation

The system models a realistic engineering workflow with specialized agent roles:

* 🟠 **ARCHITECT (The Strategist)**: High-priority system design & planning.
* 🔵 **DEVELOPER (The Builder)**: Code implementation & feature delivery.
* 🟢 **QA ENGINEER (The Auditor)**: Testing, verification & quality assurance.

---

## ⚡ Quick Start

### Prerequisites

* Docker & Docker Compose

### 1. Launch the Mesh

Start the entire infrastructure (Redis, Postgres, Producer, Worker, UI) with one command:

```powershell
./run.ps1
```

*Access the dashboard at `http://localhost:5173`*
![alt text](image.png)

### 2. Simulate Load

Flood the system with **500 concurrent tasks** to test the "High Traffic Mode" and throughput:

```powershell
./stress.ps1
```

*Watch the "Pending Queue" explode and drain in seconds while the UI stays buttery smooth.*

---

## Technology Stack

| Component | Technology | Role |
| :--- | :--- | :--- |
| **Backend** | **Go (Golang) 1.23** | High-concurrency workers & API |
| **Frontend** | **React + Vite** | Real-time visualization |
| **Styling** | **Tailwind CSS + Framer Motion** | GPU-accelerated animations |
| **Broker** | **Redis** | Pub/Sub event bus & Task Queue |
| **Database** | **PostgreSQL 16** | ACID-compliant state persistence |
| **Metrics** | **gopsutil** | Real-time CPU/RSS Memory tracking |

---

*Designed & Engineered by [YehiaGewily](https://github.com/YehiaGewily).*
