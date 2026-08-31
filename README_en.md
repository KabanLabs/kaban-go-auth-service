[English](README_en.md) | [Русский](README.md)

# Kaban SSO Service

Authentication and Authorization Service (Single Sign-On) for the Kaban ecosystem.

This microservice is responsible for user management, secure password storage, and JWT token generation/validation. It provides a gRPC API for internal services (e.g., token validation on the fly) and a REST API for client applications.

## Features

- **Cryptography**: Passwords are hashed using `Argon2id` (with fallback support for `bcrypt`), protecting against brute-force and timing attacks.
- **JWT & RSA**: Tokens are signed using asymmetric keys (RSA512). Public keys are exposed in JWK format.
- **Refresh Token Rotation**: Secure session renewal mechanism with old token invalidation.
- **Microservices Ready**: Provides a gRPC API for seamless integration with other backend services.
- **Database Migrations**: Built-in migration tool using `golang-migrate` for PostgreSQL.
- **Observability**: Prometheus metrics (login attempts, validations) and structured logging (`log/slog`).

## Tech Stack

- **Go 1.25**
- **gRPC:** [google.golang.org/grpc](https://pkg.go.dev/google.golang.org/grpc)
- **Database:** PostgreSQL (`pgx/v5`)
- **Migrations:** [golang-migrate](https://github.com/golang-migrate/migrate)
- **Tokens:** JWT (`golang-jwt/jwt/v5`)
- **Crypto:** `golang.org/x/crypto/argon2`
- **Config Management:** `cleanenv`

## Project Structure

```text
├── cmd
│   ├── sso             # Main application (gRPC + Gateway)
│   └── migrator        # Migration runner utility
├── config              # YAML configurations
├── internal
│   ├── app             # gRPC and HTTP Gateway assembly
│   ├── domain          # Domain entities (User, App, Token)
│   ├── grpc            # gRPC handlers
│   ├── lib             # Hashing (Argon2id), JWT, RSA keys store
│   ├── services        # Authentication business logic
│   └── storage         # PostgreSQL repositories
├── migrations          # SQL migration scripts
└── Dockerfile          # Docker build instructions
```

## Setup & Running

### Requirements
- Go 1.25+
- PostgreSQL
- [Task](https://taskfile.dev/) (optional)

### Configuration (`config/local.yaml`)
```yaml
env: 'local'
grpc:
  port: 44044
  timeout: 10s
pg_config:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "password"
  db_name: "sso_db"
  ssl_mode: "disable"
```

### Running Locally
Run database migrations:
```bash
go run ./cmd/migrator --config=./config/local.yaml
```

Start the SSO server:
```bash
go run ./cmd/sso --config=./config/local.yaml
```

## Testing
The project includes unit tests covering hashing algorithms, configuration parsing, and error handling.
```bash
go test ./... -v
```
