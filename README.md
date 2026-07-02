# SOFRESERVE

SOFRESERVE is a backend-focused event reservation and ticketing platform built with Go.

The project was created as a portfolio application to demonstrate backend engineering skills through a realistic event management workflow rather than isolated CRUD operations.

It simulates the complete lifecycle of an event, including event creation, reservations, individual ticket generation, public ticket access, participant check-in, and organizer dashboards.

The main goal is to demonstrate software architecture, business workflow implementation, and clean backend design.

---

# 🎯 What This Project Demonstrates

SOFRESERVE showcases practical backend development concepts, including:

- Clean Architecture
- Layered application design
- Repository Pattern
- Dependency Injection
- HTTP server development with Go
- PostgreSQL integration
- Token-based ticket validation
- QR-code ticket generation
- Server-side rendering
- Business workflow implementation
- Benchmarking and stress testing
- Separation of concerns

Rather than exposing isolated endpoints, the project models an end-to-end business process similar to real-world ticketing platforms.

---

# 🔄 End-to-End Workflow

```txt
Organizer creates an event
        │
        ▼
Participant submits a reservation
        │
        ▼
Reservation is confirmed
        │
        ▼
Individual tickets are generated
(one token per participant)
        │
        ▼
Organizer shares tickets via WhatsApp
        │
        ▼
Participant opens:

/ticket/{token}

        │
        ▼
Public ticket page displays:
- Event information
- Full ticket token
- QR Code
- Current status
        │
        ▼
Organizer validates ticket
during event entrance
        │
        ▼
Ticket status becomes:

CHECKED-IN
```

---

# 🚀 Features

- Event creation
- Public reservation page
- Reservation confirmation
- Reservation cancellation
- Individual ticket generation
- Public ticket page
- QR-code rendering
- Token-based ticket validation
- Event check-in
- Organizer dashboard
- Reservation management
- Recent check-in history
- WhatsApp ticket sharing
- Benchmark / stress test tool

---

# 🛠 Tech Stack

- Go
- PostgreSQL
- HTML Templates
- CSS
- Vanilla JavaScript
- Docker Compose

---

# 🏛 Architecture

The project follows Clean Architecture principles with clear separation between business rules, HTTP layer and persistence.

```txt
cmd/

internal/
    adapter/
        http/
        repository/

    core/
        entity/
        usecase/
        port/

    shared/

    view/

migrations/
```

Main architectural concepts:

- HTTP Handlers
- Use Cases
- Repository Pattern
- Dependency Injection
- Layered Architecture
- Separation of Concerns

---

# ⚙ Configuration

```bash
cp .env.example .env
```

---

# ▶ Running locally

Start PostgreSQL:

```bash
docker-compose up -d
```

Run the application:

```bash
go run cmd/api/main.go
```

Application:

```txt
http://localhost:8080
```

---

# 📊 Benchmark

SOFRESERVE includes a lightweight benchmark tool capable of generating concurrent HTTP requests against application endpoints.

Current benchmark capabilities include:

- Configurable concurrency
- Configurable request volume
- Average latency
- Min / Max latency
- Requests per second (RPS)
- HTTP success/error statistics

Example:

```bash
go run cmd/bench/main.go \
  -endpoint=http://localhost:8080/events/{event}/checkin \
  -token={ticket_token} \
  -requests=100 \
  -concurrency=10
```

---

# ✅ Current Functionality

- Event management
- Reservation workflow
- Confirmation flow
- Individual ticket generation
- Public ticket visualization
- QR-code rendering
- WhatsApp ticket sharing
- Organizer dashboard
- Participant check-in
- Benchmarking

## Screenshots

### Event Owner Dashboard

![Event Dashboard](docs/screenshots/sofreserve_dashboard.png)

### Reservation Confirmation

![Reservation Confirmation](docs/screenshots/sofreserve_ticket_confirmation.png)

### Public Access Ticket

![Public Ticket](docs/screenshots/sofreserve_public_access_ticket.png)

### Event Check-in

![Event Check-in](docs/screenshots/sofreserve_event_checkin.png)

---

## Goals

The project is being continuously improved with focus on:

- Clean architecture concepts
- Realistic backend flows
- Scalability
- Event management experience
- Go backend development practices

```

```
