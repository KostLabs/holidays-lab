FROM golang:1.25 as builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o holidays-calculator ./cmd

FROM gcr.io/distroless/base-debian12

WORKDIR /app
COPY --from=builder /app/holidays-calculator /app/holidays-calculator
COPY config/config.yaml /app/config/config.yaml
COPY config/config.docker.yaml /app/config/config.docker.yaml

EXPOSE 8082
ENTRYPOINT ["/app/holidays-calculator"]
