# 📦 Distributed Job Queue System (GoLang)

## 🧩 Problem Statement

Build a distributed, fault-tolerant, and extensible job queue system using Go. The system must handle asynchronous background tasks like sending emails, generating reports, etc., and must support retries, prioritization, scheduling, observability, and failure handling.

---

## 🎯 Goals

- RESTful API for job submission and status tracking
- Multiple worker support with concurrency
- Retry and dead-letter queue mechanism
- Job prioritization (high, medium, low)
- Delayed job scheduling
- Monitoring via Prometheus
- Optional: Admin UI dashboard

---

## 🔧 Tech Stack

| Component     | Tech                 |
| ------------- | -------------------- |
| Language      | Go                   |
| Queueing      | Redis or PostgreSQL  |
| API Framework | chi/gin              |
| Monitoring    | Prometheus + Grafana |
| Deployment    | Docker, optional K8s |

---

## ✅ Features

### 1. Job Producer API

```http
POST /submit-job
{
  "type": "email",
  "payload": {"email": "abc@xyz.com"},
  "priority": "high",
  "delay": 30
}
```

- Stores job in Redis (list/sorted set) or Postgres with `run_at` field

### 2. Worker Pool

- Goroutines fetch and execute jobs
- Configurable concurrency (e.g., 5 workers)

### 3. Retry + Dead Letter Queue

- Exponential backoff retry mechanism
- Max retries configurable (e.g., 3)
- Moves failed jobs to `dead-letter` queue with reason

### 4. Prioritization

- Use separate queues for priorities: `job:high`, `job:medium`, `job:low`
- Workers poll high priority queues first

### 5. Delayed Job Scheduling

- Scheduled jobs stored in Redis sorted set or Postgres `run_at`
- Mover service polls and enqueues ready jobs to live queues

### 6. Job Status Tracking

```http
GET /job-status/:id
```

Returns job metadata, status, retries, and result

### 7. Admin Dashboard (Optional)

- View active, failed, and dead-letter jobs
- Retry/Resubmit failed jobs

### 8. Observability

- Prometheus metrics:
  - `job_processing_duration_seconds`
  - `jobs_failed_total`
  - `jobs_processed_total`
- Health check endpoint: `/healthz`

### 9. Scalability

- Multiple worker nodes supported
- Redis BLPOP or Postgres row locking to avoid duplicate processing

---

## 📘 Example Use Case

- User signup triggers `/submit-job` with `type=email_welcome`
- Worker picks the job, sends welcome email
- On failure, retries 3 times, then moves to dead-letter

---

## 📦 Optional Improvements

- CLI for job management (enqueue/retry/delete)
- Graceful shutdown and job requeue
- Dashboard built with Go templates or React
- Integration tests for full pipeline

---

Generated docs ( may be inaccurate )
