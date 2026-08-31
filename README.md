[English](README_en.md) | [Русский](README.md)

# Kaban SSO Service

Сервис аутентификации и авторизации (Single Sign-On) для экосистемы Kaban.

Этот микросервис отвечает за управление пользователями, безопасное хранение паролей, генерацию и валидацию JWT-токенов. Он предоставляет gRPC API для внутренних сервисов (например, для проверки токенов на лету) и REST API для клиентских приложений.

## Особенности

- **Алгоритмы шифрования**: Пароли хэшируются с использованием `Argon2id` (с поддержкой обратной совместимости с `bcrypt`), что защищает от брутфорса и атак по сторонним каналам.
- **JWT & RSA**: Токены подписываются асимметричными ключами (RSA512). Публичные ключи раздаются в формате JWK.
- **Ротация Refresh-токенов**: Безопасный механизм обновления сессий с инвалидацией старых токенов.
- **Микросервисная архитектура**: Предоставляет gRPC API для других сервисов (напр. Kaban Syncer Service).
- **Миграции БД**: Встроенный механизм миграций с использованием `golang-migrate` для PostgreSQL.
- **Observability**: Prometheus метрики (попытки входа, валидации токенов) и структурированное логирование (`log/slog`).

## Стек технологий

- **Go 1.25**
- **gRPC:** [google.golang.org/grpc](https://pkg.go.dev/google.golang.org/grpc)
- **База данных:** PostgreSQL (`pgx/v5`)
- **Миграции:** [golang-migrate](https://github.com/golang-migrate/migrate)
- **Токены:** JWT (`golang-jwt/jwt/v5`)
- **Криптография:** `golang.org/x/crypto/argon2`
- **Конфигурация:** `cleanenv`

## Структура проекта

```text
├── cmd
│   ├── sso             # Основное приложение (gRPC + Gateway)
│   └── migrator        # Утилита для наката миграций
├── config              # YAML конфигурации
├── internal
│   ├── app             # Сборка gRPC и HTTP Gateway
│   ├── domain          # Сущности (User, App, Token)
│   ├── grpc            # gRPC хэндлеры
│   ├── lib             # Хэширование (Argon2id), JWT, RSA ключи
│   ├── services        # Бизнес-логика аутентификации
│   └── storage         # PostgreSQL репозитории
├── migrations          # SQL скрипты миграций
└── Dockerfile          # Сборка сервиса
```

## Запуск проекта

### Требования
- Go 1.25+
- PostgreSQL
- [Task](https://taskfile.dev/) (опционально)

### Конфигурация (`config/local.yaml`)
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

### Запуск локально
Накатить миграции:
```bash
go run ./cmd/migrator --config=./config/local.yaml
```

Запустить сервер:
```bash
go run ./cmd/sso --config=./config/local.yaml
```

## Тестирование
Проект покрыт unit-тестами для алгоритмов хэширования, парсинга конфигурации и обработки ошибок.
```bash
go test ./... -v
```
