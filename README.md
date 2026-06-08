# E-Commerce REST API

A robust, production-ready E-commerce RESTful API built from scratch using Go. This project follows a clean, modular architecture and integrates modern DevOps tools for local development, database migrations, and cloud-native file storage.

## Features

* **RESTful Routing & Architecture:** Clean separation of concerns with a dedicated `internal` layer and `cmd/api` entry point.
* **Database Migrations:** Structured DB versioning managed easily via automated migration files.
* **AWS S3 Integration:** Built-in S3 file upload service for handling product images and digital assets.
* **LocalStack & Nginx Environment:** Full local replication of AWS S3 via LocalStack, coupled with an Nginx reverse proxy to flawlessly serve uploaded files locally.
* **Linting & Quality:** Pre-configured with `golangci-lint` to maintain strict code quality standards.

---

## Tech Stack

* **Language:** Go (Golang)
* **Infrastructure & Local Cloud:** Docker, LocalStack (S3)
* **Web Server / Reverse Proxy:** Nginx
* **Tooling:** Makefile, Golangci-lint

---

## Repository Structure

```text
├── .env-example         # Template for environment variables
├── .golangci.yml        # Configuration for Go linting
├── Dockerfile           # Docker configuration for the API application
├── Makefile             # Commands for building, running, and migrating
├── cmd/
│   └── api/             # Application entry point
├── db/
│   └── migrations/      # SQL database migration files
├── docker/              # LocalStack and Nginx infrastructure setups
└── internal/            # Core business logic, handlers, and services (including S3 Uploader)
```

---

## Getting Started

### Prerequisites

Ensure you have the following installed on your local machine:
* **Go** (latest stable version)
* **Docker & Docker Desktop**
* `make` utility

### Installation & Setup

1. **Clone the Repository:**
   ```bash
   git clone [https://github.com/zellis-rameesn/E-commerce-rest-api.git](https://github.com/zellis-rameesn/E-commerce-rest-api.git)
   cd E-commerce-rest-api
   ```

2. **Configure Environment Variables:**
   Copy the example environment file and update it with your configurations:
   ```bash
   cp .env-example .env
   ```

3. **Run Database Migrations:**
   Use the Makefile helper to spin up your database schema:
   ```bash
   make migrate-up
   ```

4. **Spin up Infrastructure (Docker, LocalStack, Nginx):**
   ```bash
   docker-compose up -d
   ```

---

## 🛠️ Development & Commands

The project includes a `Makefile` to simplify everyday development tasks:

| Command | Description |
| :--- | :--- |
| `make build` | Compiles the Go application binaries. |
| `make run` | Launches the REST API server locally. |
| `make migrate-up` | Applies all pending database migrations. |
| `make lint` | Runs the code linter using the `.golangci.yml` rules. |

---

## 🔒 Security & Quality

* Code quality is enforced using **`golangci-lint`**. Run it before opening pull requests to ensure your code adheres to standard Go best practices.
* AWS S3 configurations point securely to a localized environment via Mock endpoints in development, keeping credentials entirely safe.
