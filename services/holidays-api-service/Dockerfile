FROM golang:1.25 as builder

WORKDIR /app/services/holidays-api-service

# Module files
COPY services/holidays-api-service/go.mod ./go.mod
COPY services/holidays-api-service/go.sum ./go.sum

# Shared observability module
COPY pkg /app/pkg

RUN go mod download

# Service source
COPY services/holidays-api-service ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/holidays-api ./cmd

FROM gcr.io/distroless/base-debian12

WORKDIR /app
COPY --from=builder /app/holidays-api /app/holidays-api
COPY services/holidays-api-service/config/config.docker.yaml /app/config/config.docker.yaml

EXPOSE 8081
ENTRYPOINT ["/app/holidays-api"]
