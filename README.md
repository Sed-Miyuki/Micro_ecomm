# Ecomm Microservice

A Golang-based e-commerce application built using a microservice architecture. The project consists of three microservices:

* **ecomm-api** — HTTP REST API for interacting with the application.
* **ecomm-grpc** — gRPC service for internal communication and business operations.
* **ecomm-notification** — Handles order-related email notifications through a stateful database-backed queue.

The project uses **PostgreSQL** as its database, with **pgx** for PostgreSQL connectivity and database operations.

## Tech Stack

* **Language:** Go
* **Database:** PostgreSQL
* **Database Driver:** pgx
* **API:** HTTP REST
* **Internal Communication:** gRPC
* **Authentication:** JWT
* **Routing:** go-chi
* **Migrations:** golang-migrate
* **Containerization:** Docker

## Project Structure

```text
.
├── cmd/
│   ├── ecomm-api/
│   ├── ecomm-grpc/
│   └── ecomm-notification/
├── db/
│   └── migrations/
├── internal/
├── .env.example
├── .gitignore
├── go.mod
└── README.md
```

## Getting Started

### Prerequisites

Make sure the following are installed:

* [Go](https://go.dev/dl/)
* [Docker](https://docs.docker.com/get-docker/)
* PostgreSQL or Docker
* [golang-migrate](https://github.com/golang-migrate/migrate)

### Environment Variables

The application uses a `.env` file to configure the PostgreSQL database, authentication, service address, and admin account.

Create a `.env` file in the project root:

```env
POSTGRES_PASSWORD=your_secure_password
POSTGRES_DB=ecomm

DB_URL=postgres://postgres:your_secure_password@postgres:5432/ecomm?sslmode=disable

SECRET_KEY=your_32_character_secret_key

SCV_ADDR=0.0.0.0:9091

ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=your_secure_admin_password
```

#### Environment Variable Description

| Variable            | Description                                                         |
| ------------------- | ------------------------------------------------------------------- |
| `POSTGRES_PASSWORD` | Password for the PostgreSQL database user.                          |
| `POSTGRES_DB`       | Name of the PostgreSQL database.                                    |
| `DB_URL`            | PostgreSQL connection string used by the application through `pgx`. |
| `SECRET_KEY`        | Secret key used for authentication and JWT signing.                 |
| `SCV_ADDR`          | Address and port on which the service listens.                      |
| `ADMIN_EMAIL`       | Email address of the initial admin account.                         |
| `ADMIN_PASSWORD`    | Password of the initial admin account.                              |

**Important:**

* Never commit `.env` to version control.
* Add `.env` to `.gitignore`.
* Replace all placeholder values with secure credentials.
* The hostname `postgres` in `DB_URL` should resolve to the PostgreSQL container or service when running with Docker Compose.
* If you run the Go application directly on your host machine, you may need to replace `postgres` with `localhost` and use the exposed PostgreSQL port.
* Ensure `SECRET_KEY` meets the length requirements enforced by your authentication implementation.

### Database Setup

Start a PostgreSQL container using Docker:

```bash
docker pull postgres:16

docker run --name ecomm-postgres \
  -p 5432:5432 \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=your_secure_password \
  -e POSTGRES_DB=ecomm \
  -d postgres:16
```

Apply the database migrations:

```bash
migrate \
  -path db/migrations \
  -database "postgres://postgres:your_secure_password@localhost:5432/ecomm?sslmode=disable" \
  up
```

Make sure the database connection string matches the configuration used by the Go microservices.

### Run the Go Applications

Start the gRPC microservice:

```bash
go run cmd/ecomm-grpc/main.go
```

Start the API microservice in a separate terminal:

```bash
go run cmd/ecomm-api/main.go
```

Start the notification microservice according to its configured entry point:

```bash
go run cmd/ecomm-notification/main.go
```

## Notification Queue

When an admin creates an order, a notification event for the `pending` status is enqueued into the `notification_events_queue` table.

A new notification event is also enqueued whenever the order status is updated, allowing the system to send an email notification for each update.

The `notification_states` table maintains the state of notification events even after they have been deleted or dequeued from the `notification_events_queue` table.

A notification event can have one of the following states:

* `not sent`
* `sent`
* `failed`

The state of a notification event represents the status of the email delivery process. It is independent of the order's status.

### Notification Processing

The `ecomm-notification` microservice retrieves notification events from the database queue, ordered by their `created_at` timestamps, and attempts to send an email notification for each event.

There are three possible outcomes:

1. **Successful delivery:** The event is deleted from the notification queue, and its state in `notification_states` is updated to `sent`.
2. **Failed delivery with remaining attempts:** If the number of attempts is below the configured maximum, the attempt count is updated. The event remains in the queue and is retried during the next processing cycle.
3. **Failed delivery after maximum attempts:** If the maximum number of attempts has been reached, the event is deleted from the queue, and its state is updated to `failed`.

## Security

* Keep sensitive credentials in `.env` and never commit them to version control.
* Use strong, unique passwords for the database and admin account.
* Use a secure secret key for JWT signing.
* Do not use development credentials in production.