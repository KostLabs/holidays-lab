FROM golang:1.25 as builder

WORKDIR /app/services/holidays-calculator-service

# Module files
COPY services/holidays-calculator-service/go.mod ./go.mod
COPY services/holidays-calculator-service/go.sum ./go.sum

# Shared observability module
COPY pkg /app/pkg

RUN go mod download

# Service source
COPY services/holidays-calculator-service ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/holidays-calculator ./cmd

FROM gcr.io/distroless/base-debian12

WORKDIR /app
COPY --from=builder /app/holidays-calculator /app/holidays-calculator
COPY services/holidays-calculator-service/config/config.yaml /app/config/config.yaml
COPY services/holidays-calculator-service/config/config.docker.yaml /app/config/config.docker.yaml

EXPOSE 8082
ENTRYPOINT ["/app/holidays-calculator"]
