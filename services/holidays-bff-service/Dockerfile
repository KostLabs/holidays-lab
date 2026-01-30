FROM golang:1.25 as builder

WORKDIR /app/services/holidays-bff-service

# Module files
COPY services/holidays-bff-service/go.mod ./go.mod
COPY services/holidays-bff-service/go.sum ./go.sum

# Shared observability module
COPY pkg /app/pkg

RUN go mod download

# Service source
COPY services/holidays-bff-service ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bff ./cmd

FROM gcr.io/distroless/base-debian12

WORKDIR /app
COPY --from=builder /app/bff /app/bff
COPY services/holidays-bff-service/config/config.docker.yaml /app/config/config.docker.yaml

EXPOSE 8080
ENTRYPOINT ["/app/bff"]
