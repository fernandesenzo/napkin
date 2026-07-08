a real-time collaborative notepad written in Go using redis and websockets.

## project structure

```
├── .github/workflows/   # CI pipeline configuration
├── cmd/api/             # app entrypoint (main.go)
└── internal/
    ├── client/          # websocket client (read/write pumps)
    ├── hub/             # per-room broadcast hub
    ├── infra/           # redis client initialization
    ├── ip/              # client ip extraction
    ├── logger/          # logger setup (slog)
    ├── manager/         # room lifecycle manager
    ├── middleware/       # global HTTP middlewares (recovery, request ID, logs, rate-limit)
    └── napkin/          # napkin domain logic
        ├── handler/     # HTTP + websocket handlers
        ├── repository/  # redis data layer
        ├── service/     # orchestrator / service layer
        └── napkin.go    # domain constants and validation
```

## running
1. copy the environment template:
   ```bash
   cp .env.example .env
   ```

2. start the services using docker compose:
   ```bash
   docker compose up --build
   ```
   the API will run on port `8082`, but you can change it to your favorite port on docker-compose.yml.

you can also run the api locally:
```bash
docker compose up -d redis-local
go run ./cmd/api
```

### save a napkin
`POST /api/save`
```json
{
  "code": "myroom",
  "content": "hello world"
}
```

response:
```json
{
  "content": "hello world"
}
```

### get a napkin
`GET /{code}`

returns the napkin contents as JSON.

### real-time collaboration
`GET /{code}/ws`

upgrades to a websocket connection. every client in the same room receives messages broadcast by other clients. messages are persisted to redis automatically.

## business logic
by default, codes have length 6, content is limited to 200 characters and napkins last for 24 hours. rate limiting is per-ip: 5 POST/min and 100 GET/min. you can change these settings on `internal/napkin/napkin.go` and the `.env` file.

## tests and linter

run tests:
```bash
go test ./...
```

run linter:
```bash
golangci-lint run
```

there is also a github actions pipeline to run tests and linter on every push to main
