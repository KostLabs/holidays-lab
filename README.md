# Holiday lab application

Holiday lab application built in a microservices architecture using Go, Gin, MongoDB, k6 for load testing, and OpenTelemetry + New Relic for observability. The goal of the project is to demonstrate best practices in building scalable and observable microservices.

## Services

- holidays-api-service
	- Stores holidays in MongoDB.
	- Fetches Estonian holidays from the public API `https://xn--riigiphad-v9a.ee/?output=json&year=YYYY`.
	- Endpoint: `GET /api/v1/holidays-api/fetch?year=YYYY`.

- holidays-calculator-service
	- Uses holidays-api-service to look up holidays.
	- Calculates days until a named holiday from a given date.
	- Endpoint: `GET /api/v1/holidays-calculator/calculate?date=YYYY-MM-DD&name=holiday_name`.

- holidays-bff-service
	- Public entrypoint / API gateway.
	- Forwards requests to holidays-api-service and holidays-calculator-service based on configuration.
	- Endpoints:
		- `GET /api/v1/holidays-bff/holidays?year=YYYY`
		- `GET /api/v1/holidays-bff/holidays/calculate?date=YYYY-MM-DD&name=holiday_name`

## Prerequisites

- Go 1.25+ (for local builds).
- Docker and Docker Compose.
- A New Relic account and license key (for traces/metrics).

## Configuration

Each service reads its configuration from a YAML file and environment variables.

For Docker usage, configuration is provided via `config.docker.yaml` in each service directory and environment variables declared in `docker-compose.yml`.

Observability is configured with the following environment variables (used by the shared OpenTelemetry module):

- `OTEL_EXPORTER_OTLP_ENDPOINT` – New Relic OTLP endpoint, for EU region for example: `https://otlp.eu01.nr-data.net`.
- `NEW_RELIC_LICENSE_KEY` – your New Relic ingest license key.
- `OTEL_SERVICE_NAME` – set by each service; normally this is already configured in code.

Create a `.env` file in the repository root (or export variables in your shell) with at least:

```bash
NEW_RELIC_LICENSE_KEY=your_license_key_here
OTEL_EXPORTER_OTLP_ENDPOINT=https://otlp.eu01.nr-data.net
```

## Running with Docker

From the repository root:

```bash
docker compose up --build
```

This will start:

- MongoDB
- holidays-api-service
- holidays-calculator-service
- holidays-bff-service

Once all containers are healthy, the BFF will be available on the port configured in `docker-compose.yml` (by default `http://localhost:8080`).

## Example requests

Assuming the BFF is running on `http://localhost:8080`:

```bash
# Fetch holidays for a year via BFF
curl "http://localhost:8080/api/v1/holidays-bff/holidays?year=2025"

# Calculate days until a named holiday via BFF
curl "http://localhost:8080/api/v1/holidays-bff/holidays/calculate?date=2025-01-01&name=Jaanipäev"
```

You can also call the downstream services directly if they are exposed on separate ports (see `docker-compose.yml`):

```bash
# Direct holidays-api-service
curl "http://localhost:<api_port>/api/v1/holidays-api/fetch?year=2025"

# Direct holidays-calculator-service
curl "http://localhost:<calculator_port>/api/v1/holidays-calculator/calculate?date=2025-01-01&name=Jaanipäev"
```

## Load testing

The project is designed to work with k6 for load testing. You can add or adapt k6 scripts to hit the BFF endpoints above and observe performance and traces in New Relic.

## Observability

All services are instrumented with OpenTelemetry and export traces (and related spans for HTTP and MongoDB) to New Relic using the OTLP exporter.

In New Relic, you should see at least these services:

- `holidays-bff-service`
- `holidays-api-service`
- `holidays-calculator-service`

Service maps should show calls from the BFF to the calculator and API services, and database operations from holidays-api-service to MongoDB.
