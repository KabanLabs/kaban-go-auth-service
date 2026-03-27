FROM golang:alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o sso-app ./cmd/sso
RUN go build -o migrator-app ./cmd/migrator

FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/sso-app .
COPY --from=builder /app/migrator-app .
COPY --from=builder /app/config ./config
COPY --from=builder /app/migrations ./migrations


ENV CONFIG_PATH=./config/local.yaml

EXPOSE 5055 8000

# По умолчанию применяем миграции (up) и запускаем основное приложение
CMD ["sh", "-c", "./migrator-app --migrations-path=./migrations && ./sso-app"]
