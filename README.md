# PortiGo
<img width="1000" height="1000" alt="image" src="https://github.com/user-attachments/assets/77a7c044-5603-4c85-a292-836a88bbd242" />


LLM gateway written in Go that routes OpenAI-compatible requests to provider endpoints based on the request body model id.

## Quick start (Docker Compose)

```bash
docker compose up --build
```

- API: http://localhost:8080
- Postgres: localhost:5432 (user: app, pass: app, db: go_gateway)

Register a model:
```bash
curl -X POST http://localhost:8080/v1/models \
  -H 'Content-Type: application/json' \
  -d '{
    "id": "example/model-1",
    "name": "Example Model 1",
    "description": "",
    "endpoint": "https://api.example-llm.com",
    "api_key": ""
  }'
```

Send a chat completion:
```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -d '{
    "model": "example/model-1",
    "messages": [{"role":"user","content":"Say hello in one sentence."}],
    "stream": false
  }'
```

Or completions:
```bash
curl -X POST http://localhost:8080/v1/completions \
  -H 'Content-Type: application/json' \
  -d '{
    "model": "example/model-1",
    "prompt": "Write a haiku about containers"
  }'
```

## Local development (without Docker)

1) Start Postgres (example using Docker):
```bash
docker run --name go_gateway_pg -e POSTGRES_USER=app -e POSTGRES_PASSWORD=app -e POSTGRES_DB=go_gateway -p 5432:5432 -d postgres:16-alpine
```

2) Run the API:
```bash
export DATABASE_URL="postgres://app:app@localhost:5432/go_gateway?sslmode=disable"
export PORT=8080
go run ./cmd/gateway
```

## Configuration

- `DATABASE_URL` (default: `postgres://localhost:5432/go_gateway?sslmode=disable`)
- `PORT` (default: `8080`)

## API

- `GET /healthz`: health check
- `GET /v1/models`: list registered models
- `POST /v1/models`: create/update model
  - body: `{ id, name, description, endpoint, api_key }`
- `POST /v1/chat/completions`: proxy to provider based on body.model
- `POST /v1/completions`: proxy to provider based on body.model

Responses are passed through from the upstream provider.

## Notes

- The gateway reads `model` from the JSON body and routes to that model's `endpoint`.
- Preserve the upstream versioned paths; set `endpoint` to the base URL (no trailing `/v1`).
- If the client omits `Authorization`, the gateway sets `Authorization: Bearer <api_key>` when the stored `api_key` is non-empty.
- Streaming responses are supported and forwarded.
