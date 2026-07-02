# SOFRESERVE

SOFRESERVE is a backend-focused event reservation and ticketing system built with Go.

The project was designed as a real-world backend simulation to demonstrate system design skills, clean architecture principles, and the ability to build scalable and realistic event-driven flows.

Instead of focusing only on CRUD operations, the system models a complete ticket lifecycle: from event creation to participant check-in at entry.

---

## 🎯 What This Project Demonstrates

SOFRESERVE was built to showcase backend engineering skills beyond basic APIs, including:

- System design for real-world flows
- Clean Architecture in Go
- Stateless ticket validation using tokens
- High-level separation of concerns
- Multi-step business workflows
- Server-side rendered applications
- Practical trade-offs between simplicity and realism

---

## 🔄 End-to-End System Flow

The system simulates a real event ticketing lifecycle:

````txt
1. Event creation by organizer
2. Participant reservation request
3. Ticket generation (unique token per participant)
4. Ticket distribution via WhatsApp share link
5. Participant accesses public ticket page (/ticket/{token})
6. Ticket is rendered with QR code and validation data
7. Check-in is performed at event entrance using token
8. Ticket status is updated to CHECKED-IN

## Features

- Event reservation flow
- Public ticket page
- QR-code ticket validation
- Event check-in flow
- Shareable ticket links
- WhatsApp ticket sharing
- Server-side rendering
- Token-based ticket access

---

## Stack

- Go
- PostgreSQL
- HTML Templates
- CSS / Vanilla JavaScript
- Docker Compose

---

## Architecture

Project organized using layered architecture concepts:

- HTTP Handlers
- UseCases
- Repositories
- Shared Services
- HTML Templates

Current structure:

```txt
cmd/
internal/
  adapter/
  core/
  shared/
  view/
migrations/
````

---

## Configuration (Edit the environment variables)

```bash
cp .env.example .env
```

---

## Running locally

### 1. Start database

```bash
docker-compose up -d
```

### 2. Run application

```bash
go run cmd/api/main.go
```

Application will run on:

```txt
http://localhost:8080
```

---

## Current Features Demonstrated

- Reservation confirmation flow
- Public participant ticket page
- Individual ticket tokens
- QR-code generation
- Event check-in validation
- WhatsApp ticket sharing
- Event owner dashboard

---

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
