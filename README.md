# Notification Service

Event-driven, microservice-based notification system built with **Go, PostgreSQL, Kafka, and Docker Compose**.

The system supports multi-channel notifications (SMS, Email, Push) with priority-based routing and an Outbox pattern for reliable event publishing.

---

# 🏗 Architecture Overview

## High-Level Architecture

```
Client → Notification API → PostgreSQL
                           ↓
                        Outbox Table
                           ↓
                     Outbox Publisher
                           ↓
                          Kafka
          ┌───────────────┼───────────────┐
       SMS Worker     Email Worker     Push Worker
```

---

## Components

### 1️⃣ notification-api

* REST API
* Writes notifications to `notifications` table
* Inserts event into `outbox` table (same transaction)
* Does NOT publish directly to Kafka (Outbox pattern)

### 2️⃣ outbox-publisher

* Polls `outbox` table
* Publishes events to Kafka topics
* Updates status (`pending → processing → published/failed`)
* Handles retry logic

### 3️⃣ Workers

Each channel has its own worker:

* `sms-worker`
* `email-worker`
* `push-worker`

Workers:

* Consume from Kafka
* Process notification
* Commit message

### 4️⃣ Kafka

Topics follow naming convention:

```
{channel}_{priority}
```

Examples:

* `sms_high`
* `email_medium`
* `push_low`

---

# 🗄 Database Schema

## notifications

| Column       | Type                 |
| ------------ | -------------------- |
| id           | UUID / string        |
| group_id     | string               |
| recipient    | text                 |
| channel      | text                 |
| content      | text                 |
| status       | text                 |
| priority     | text                 |
| scheduled_at | timestamp (nullable) |
| created_at   | timestamp            |

## outbox

| Column       | Type                                      |
| ------------ | ----------------------------------------- |
| id           | UUID                                      |
| aggregate_id | string                                    |
| group_id     | string                                    |
| event_type   | text                                      |
| topic        | text                                      |
| payload      | jsonb                                     |
| status       | pending / published / sended / failed     |
| retry_count  | int                                       |
| created_at   | timestamp                                 |
| published_at | timestamp                                 |

---

# 🚀 Setup Instructions

## 1️⃣ Prerequisites

* Docker
* Docker Compose
* Go 1.24+ (if running locally without Docker)

---

## 2️⃣ Start the System

From project root:

```
docker compose down -v
docker compose up --build
```

This will start:

* PostgreSQL
* Zookeeper
* Kafka
* Notification API
* Outbox Publisher
* SMS Worker
* Email Worker
* Push Worker

API will be available at:

```
http://localhost:8080
```

---

## 3️⃣ Verify Kafka Topics

Enter Kafka container:

```
docker exec -it notification-kafka bash
```

List topics:

```
kafka-topics --bootstrap-server kafka:9092 --list
```

---

## 4️⃣ Verify Database

Enter PostgreSQL container:

```
docker exec -it notification-postgres psql -U postgres -d notification
```

List tables:

```
\dt
```

---

# 📡 API Examples

## Create Notification

### Request

```
POST /notifications
Content-Type: application/json
```

```json
{
  "recipient": "+905555555555",
  "channel": "sms",
  "content": "Your verification code is 1234",
  "priority": "high"
}
```
```
GET /notifications
Content-Type: application/json
```

``` curl
curl --location 'http://localhost:8080/api/v1/notifications?status=sended&channel=email'
```

### Response

```json
{
  "id": "b1a2c3d4",
  "status": "pending",
  "createdAt": "2026-02-13T14:30:00Z"
}
```

---

## List Notifications

Supports filtering and pagination:

```
GET /notifications?status=published&channel=sms&page=1&page_size=20
```

Optional filters:

* `status`
* `channel`
* `priority`
* `start_date`
* `end_date`

---

# 🔁 Outbox Flow

1. API inserts notification + outbox record (same transaction)
2. Outbox Publisher polls `pending` events
3. Publishes to Kafka topic
4. On success → `published`
5. On failure → increments `retry_count`

---

# 📈 Scaling Strategy

* Increase Kafka partitions for higher throughput
* Scale workers horizontally
* Use consumer groups for parallel processing

---

# 🛡 Reliability Features

* Outbox Pattern (no message loss)
* Retry mechanism
* Idempotent worker design
* Transactional DB writes

---

# 🧪 Reset Environment

To fully reset Kafka and PostgreSQL:

```
docker compose down -v
docker compose up --build
```

---

# 📌 Future Improvements

* Dead Letter Queue (DLQ)
* Scheduled notification processor
* Rate limiting per channel
* Observability (Prometheus + Grafana)
* Distributed tracing

---

# 👨‍💻 Author

Notification Service – Microservice Event-Driven Architecture Example

Built with Go + Kafka + PostgreSQL
