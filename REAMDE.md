# Distributed 2PC Transaction Coordinator

A high-integrity **Two-Phase Commit (2PC)** implementation in **Go**, designed to ensure atomic consistency across heterogeneous microservices (Postgres & Redis).



## 🚀 The Problem
In a distributed architecture, updating multiple databases (e.g., a SQL ledger and a NoSQL cache) creates the risk of **partial failures**. If the SQL update succeeds but the Cache update fails, the system enters an inconsistent state. This project implements the 2PC protocol to ensure a "Global Commit" or a "Global Rollback."

## 🛠 Features
* **Atomic Transactions:** Guarantees that all participants finalize changes or none do.
* **Microservices Architecture:** Coordinator and Participants communicate via **gRPC** and **Protobuf**.
* **Write-Ahead Logging (WAL):** Transaction states are persisted in a Redis-based Log Store to survive Coordinator crashes.
* **Self-Healing Recovery:** A background worker identifies orphaned "PREPARED" transactions and drives them to completion or rollback.
* **Honest Idempotency:** Distinguishes between successful retries and attempts to reuse IDs of aborted transactions.

## 🏗 System Architecture
The system is composed of three main microservices:
1.  **Coordinator:** The orchestrating "Boss" service that manages the state machine and participant lifecycle.
2.  **Postgres Participant:** Manages SQL transactions using native `PREPARE TRANSACTION` (Two-Phase Commit support).
3.  **Redis Participant:** Manages NoSQL updates using distributed locking to simulate the "Prepare" phase.



## 🚦 Transaction State Machine
To ensure durability, every transaction follows a strict state transition stored in the Redis WAL:

| State | Description |
| :--- | :--- |
| **START** | Transaction initialized; no participant data modified yet. |
| **PREPARED** | All participants have locked resources. This is the **Point of No Return**. |
| **COMMITTED** | Final state where all changes are permanent. |
| **ABORTED** | Failure state; all locks released and changes reverted. |



## 🚦 Getting Started

### Prerequisites
* Docker & Docker Compose
* Go 1.21+
* Protobuf Compiler (to modify `.proto` files)

### Running the System
```bash
# 1. Start all databases and microservices
docker-compose up --build

# 2. Monitor the logs for the coordinator logic
docker-compose logs -f coordinator