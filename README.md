# Task Management API (Go)

A production-ready Task Management API built with Go, focusing on security, performance, and modern development practices. This service features a robust authentication and authorization system, per-device session management, and Multi-Factor Authentication (MFA).

## 🚀 Features

- **Robust Authentication**: JWT-based auth with refresh token rotation.
- **Session Management**: Per-device session tracking with server-side revocation.
- **Multi-Factor Authentication (MFA)**: TOTP support with encrypted secrets at rest.
- **Account Security**: Account activation and password reset via email (mailpit/SMTP support).
- **Profile Management**: Secure user profile viewing and updates.
- **Modern Tech Stack**: Built with Fiber (web framework), GORM (ORM), and supports Postgres/Redis.

## 🛠 Tech Stack

- **Lanuage**: Go 1.25+
- **Framework**: [Fiber v2](https://gofiber.io/)
- **Database**: Postgres (primary), GORM (ORM)
- **Cache/Session**: Redis
- **Security**: JWT, Argon2id (password hashing), AES-GCM (encryption)
- **Validation**: [Go Playground Validator](https://github.com/go-playground/validator)
- **Logging**: [Logrus](https://github.com/sirupsen/logrus) & [oarkflow/log](https://github.com/oarkflow/log)

## 📦 Getting Started

### Prerequisites

- Go 1.25+
- Postgres & Redis
- [Air](https://github.com/cosmtrek/air) (optional, for hot-reload)
- [Mailpit](https://github.com/axllent/mailpit) (for local email testing)

### Installation

1.  **Clone the repository**:
    ```bash
    git clone https://github.com/Prince-Letsyo/task-management-api-go.git
    cd task-management-api-go
    ```

2.  **Configuration**:
    The service uses `.app.config.dev.yaml` for development. Copy the example and adjust your settings:
    ```bash
    cp .app.config.dev.yaml.example .app.config.dev.yaml # If example exists, or just edit existing
    ```

3.  **Run the application**:
    ```bash
    make run
    # OR using Air
    make run.air
    ```

## 📜 Makefile Commands

| Command | Description |
| :--- | :--- |
| `make run` | Build and run the application |
| `make test` | Run tests with security and lint checks |
| `make migrate` | Run database migrations |
| `make swag` | Generate/Format Swagger documentation |
| `make lint` | Run `golangci-lint` |
| `make clean` | Remove build artifacts |

## 🔑 Authentication Endpoints

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `POST` | `/api/auth/sign-up` | Register a new user |
| `POST` | `/api/auth/sign-in` | Authenticate and receive tokens |
| `POST` | `/api/auth/sign-in-mfa` | Complete MFA authentication |
| `POST` | `/api/auth/access` | Refresh access token |
| `GET` | `/api/auth/activate-account` | Activate account via email token |
| `POST` | `/api/auth/logout` | Revoke current session |
| `GET` | `/api/auth/sessions` | List all active sessions (Auth required) |
| `POST` | `/api/auth/enable-2fa` | Enable TOTP MFA (Auth required) |

## 🛡 MFA Flow

1.  **Sign In**: Receives `requires_2fa: true` and a `temp_2fa_token`.
2.  **Generate TOTP**: Use an app like Google Authenticator.
3.  **Verify**: Send the `temp_2fa_token` and `totp_token` to `/sign-in-mfa`.

## 📮 Postman

A comprehensive Postman collection is provided for easy testing:
`resources/postman/task-management-api-go.postman_collection.json`

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.
