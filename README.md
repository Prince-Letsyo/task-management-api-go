# Task Management API (Go)

A production-ready Task Management API built with Go, focusing on security, performance, and modern architectural patterns. This service features a robust authentication system, decoupled service architecture, and asynchronous task processing.

## 🚀 Features

- **Decoupled Architecture**: Clean separation of concerns with dedicated `AccountService` (user lifecycle) and `AuthService` (authentication).
- **Asynchronous Processing**: Background email task processing using [Machinery](https://github.com/RichardKnop/machinery) with RabbitMQ/Redis backends.
- **Robust Authentication**: JWT-based auth with refresh token rotation and per-device session management.
- **Multi-Factor Authentication (MFA)**: TOTP support with encrypted secrets at rest (AES-GCM).
- **Standardized Observability**: Unified logging using [oarkflow/log](https://github.com/oarkflow/log) and structured error handling with custom `AppError` types.
- **Security First**: Argon2id password hashing, secure session revocation, and automated account activation.

## 🏗 Architecture

The project follows a modular design pattern to ensure scalability and testability:

- **Internal Modules**: Logic is encapsulated within `internal/` (e.g., `auth`, `user`, `profile`).
- **Dependency Injection**: Services and repositories utilize constructor injection, decoupled from global configuration.
- **Global Error Handling**: A centralized `ErrorHandler` maps internal errors to appropriate HTTP status codes using the `AppError` interface.

## 🛠 Tech Stack

- **Language**: Go 1.25+
- **Web Framework**: [Fiber v2](https://gofiber.io/)
- **ORM**: [GORM](https://gorm.io/) with Postgres support
- **Task Queue**: [Machinery v2](https://github.com/RichardKnop/machinery)
- **Security**: JWT, Argon2id, AES-GCM
- **In-Memory**: Redis (for sessions, caching, and queue state)

## 📦 Getting Started

### Prerequisites

- Go 1.25+
- Postgres & Redis
- RabbitMQ (for Machinery task queue)
- [Mailpit](https://github.com/axllent/mailpit) (for local email testing)

### Installation

1.  **Clone the repository**:
    ```bash
    git clone https://github.com/Prince-Letsyo/task-management-api-go.git
    cd task-management-api-go
    ```

2.  **Configuration**:
    The service uses YAML configuration. Copy the example and adjust your settings (database, redis, rabbitmq, jwt):
    ```bash
    cp .app.config.dev.yaml.example .app.config.dev.yaml
    ```

3.  **Run the application**:
    ```bash
    make run
    # OR for hot-reload development
    make run.air
    ```

## 📜 Makefile Commands

| Command | Description |
| :--- | :--- |
| `make run` | Build and run the API |
| `make test` | Run all unit and integration tests |
| `make migrate` | Execute database migrations |
| `make swag` | Generate Swagger documentation |
| `make lint` | Run golangci-lint |
| `make clean` | Cleanup build artifacts |

## 🔑 Key Endpoints

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `POST` | `/api/auth/sign-up` | Register a new user |
| `POST` | `/api/auth/sign-in` | Authenticate (supports MFA redirect) |
| `POST` | `/api/auth/sign-in-mfa` | Complete MFA verify |
| `POST` | `/api/auth/access` | Refresh access token |
| `GET` | `/api/auth/sessions` | List all active sessions |
| `POST` | `/api/profile/` | Update user profile |

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
