
# High-Performance WhatsApp Microservice (Go)

A high-throughput, non-blocking WhatsApp notification service built with **Golang**. This service utilizes the **Worker Pool pattern** and **Buffered Channels** to handle thousands of concurrent requests efficiently without relying on external message brokers like RabbitMQ.

Designed for scalability and reliability, it features **Graceful Shutdown** to ensure zero data loss during server restarts and implements **Round-Robin Session Rotation** for load balancing across multiple WhatsApp accounts.


## 🏗 Architecture

The system uses an **In-Memory Event Driven** architecture. HTTP requests are immediately acknowledged and pushed into a buffered channel. Background workers (Goroutines) pick up jobs from this channel and execute the heavy lifting (sending HTTP requests to WAHA).

```mermaid
graph LR
    User[Client / API] -- POST /send --> Gin[Gin HTTP Server]
    Gin -- Non-blocking --> Channel[Buffered Channel (RAM)]
    Channel --> Dispatcher[Dispatcher]
    Dispatcher --> Worker1[Worker 1]
    Dispatcher --> Worker2[Worker 2]
    Dispatcher --> WorkerN[Worker N...]
    Worker1 -- HTTP Request --> WAHA[WAHA Engine]
    Worker2 -- HTTP Request --> WAHA
    WAHA -- Send Message --> WhatsApp[WhatsApp Server]

```

## Key Features

*  Ultra Fast Response: The API returns `202 Accepted` immediately. Clients don't wait for the actual message delivery.
*  In-Memory Worker Pool: Replaces complex brokers (RabbitMQ/Redis) with Go's native Channels and Goroutines for lower latency and simpler infrastructure.
* Graceful Shutdown: Ensures all active workers finish their current jobs before the server stops (prevents data corruption).
* Session Load Balancer: Automatically rotates WhatsApp sessions (`session_1`, `session_2`, etc.) using a Round-Robin algorithm to prevent banning.
* Clean Architecture: Structured using standard Go project layout (`cmd`, `internal`, `pkg`).

## 📂 Folder Structure

```
whatsapp_microservices/
├── cmd/
│   └── server/
│       └── main.go       # Application Entry Point
├── internal/
│   ├── model/            # Data Structures (Payloads)
│   ├── waha/             # WAHA API Client Wrapper
│   └── worker/           # Worker Pool & Dispatcher Logic
├── .env                  # Configuration File
├── go.mod                # Go Modules
└── README.md             # Documentation

```

## 🛠 Prerequisites

* **Go** (version 1.18+)
* **WAHA (WhatsApp HTTP API)** running locally or on a server.
* *Note: This service acts as a manager/queue for WAHA.*



## 🚀 Installation & Setup

1. **Clone the repository**
```bash
git clone https://github.com/withrizky/whatsapp-microservice-go.git
cd whatsapp-microservice-go

```


2. **Install Dependencies**
```bash
go mod tidy

```


3. **Environment Configuration**
Create a `.env` file in the root directory:
```env
PORT=8080

# WAHA Configuration
WAHA_URL=http://localhost:3000
WAHA_API_KEY=your_secret_key

# WhatsApp Sessions (Comma separated for load balancing)
WAHA_SESSIONS=default,marketing,support

```


4. **Run the Server**
```bash
go run cmd/server/main.go

```



## 📡 API Documentation

### Send Message

Sends a message to the processing queue.

* **URL**: `/send`
* **Method**: `POST`
* **Content-Type**: `application/json`

**Request Body:**

```json
{
    "to": "6281234567890",
    "message": "Hello from Go Microservice!"
}

```

**Response (Success):**

```json
{
    "status": "queued",
    "message": "Pesan masuk antrean"
}

```

*Status Code: `202 Accepted*`

**Response (Error):**

```json
{
    "error": "Payload invalid"
}

```

*Status Code: `400 Bad Request*`

## 📈 Performance Strategy

This service is optimized for high concurrency:

1. **Buffered Channels**: Can hold up to **10,000** (configurable) pending messages in RAM.
2. **Concurrency**: Spawns **50** (configurable) concurrent workers. This means 50 messages are processed in parallel every millisecond.
3. **No I/O Blocking**: The HTTP handler does not wait for the 3rd party API (WAHA) to respond.

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.
