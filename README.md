# napkin

napkin is a small real-time collaborative notepad written in Go, using Redis for storage.

## what it does

- saves the latest content of a napkin in Redis
- reads napkins through a small HTTP API
- broadcasts text messages to everyone connected to the same room
- persists WebSocket messages automatically
- removes inactive rooms from process memory while keeping the document in Redis until it expires
- exposes a Redis-backed health endpoint

## project structure

```text
├── .github/workflows/   # CI pipeline configuration
├── cmd/api/             # application entrypoint
└── internal/
    ├── client/          # WebSocket client read/write pumps
    ├── hub/             # per-room broadcast hub
    ├── infra/           # Redis client initialization
    ├── ip/              # client IP extraction
    ├── logger/          # logger setup using slog
    ├── manager/         # room lifecycle manager
    ├── middleware/      # recovery, request ID, logs, headers and rate limits
    └── napkin/          # napkin domain logic
        ├── handler/     # HTTP and WebSocket handlers
        ├── repository/  # Redis data layer
        ├── service/     # application logic
        └── napkin.go    # validation and domain types
```

## running locally

copy the environment template:

```bash
cp .env.example .env
```

start the API and Redis with docker compose:

```bash
docker compose up --build
```

by default:

- the API listens on `SERVER_PORT` inside the container (`8080` in the example environment);
- Docker exposes it on `localhost:HOST_PORT` (`8082` in the example environment);
- Redis is available to the API through `redis-local:6379`.

to run the API directly on your machine while keeping Redis in Docker:

```bash
docker compose up -d redis-local
REDIS_ADDR=localhost:6379 go run ./cmd/api
```

## HTTP API

### save a napkin

```http
POST /save
Content-Type: application/json
```

Request:

```json
{
  "code": "myroom",
  "content": "hello world"
}
```

Successful response:

```http
201 Created
Content-Type: application/json
```

```json
{
  "content": "hello world"
}
```

there is no separate create endpoint: saving a new code creates the napkin, while saving an existing code overwrites it.

the code must contain only letters and numbers and must have exactly `CODE_LENGTH` characters. saving an existing code overwrites its content and resets its expiration timer.

possible errors:

| cause | status | response body |
| --- | ---: | --- |
| malformed JSON or unknown field | `400` | `invalid request` |
| invalid code | `400` | `invalid code` |
| unsupported content type | `415` | `unsupported content type` |
| content exceeds `MAX_CONTENT_LENGTH` | `422` | `content too long` |
| request body exceeds 4 KiB | `413` | `request body too large` |
| POST rate limit exceeded | `429` | `too many requests` |
| unexpected server or Redis error | `500` | `internal server error` |

error responses are currently plain text. successful save responses are JSON.

### get a napkin

```http
GET /{code}
```

Example:

```bash
curl -i http://localhost:8082/myroom
```

Successful response:

```http
200 OK
Content-Type: application/json
```

```json
{
  "code": "myroom",
  "content": "hello world"
}
```

possible errors:

| cause | status | response body |
| --- | ---: | --- |
| invalid code | `400` | `invalid code` |
| napkin does not exist or has expired | `404` | `napkin not found` |
| GET rate limit exceeded | `429` | `too many requests` |
| unexpected server or Redis error | `500` | `internal server error` |

### health check

```http
GET /health
```

the endpoint performs a real Redis ping, so it checks the dependency required by the API.

when the API and Redis are healthy:

```http
200 OK
Content-Type: application/json
```

```json
{
  "message": "ok"
}
```

when Redis is unavailable:

```http
503 Service Unavailable
Content-Type: text/plain; charset=utf-8
```

```text
error
```

## real-time collaboration

```http
GET /{code}/ws
```

this endpoint upgrades the connection to WebSocket. a browser client could connect like this:

```js
const socket = new WebSocket("ws://localhost:8082/myroom/ws");

socket.onmessage = (event) => {
  console.log("napkin updated:", event.data);
};

socket.send("hello from the room");
```

the code is validated before the upgrade. missing or invalid codes result in `400 Bad Request`, and a browser origin that is not allowed by `ALLOWED_ORIGINS` cannot complete the handshake.

### current behavior

- each code represents a room
- every connected client in the room receives every accepted text message, including the sender
- messages are plain text WebSocket frames containing the complete current content
- messages over `MAX_CONTENT_LENGTH` are dropped
- accepted messages are persisted to Redis asynchronously
- a new WebSocket connection does not receive an initial snapshot automatically, use `GET /{code}` to load the current content first
- each connection is limited to 10 messages per second with a burst of 20
- when the last client disconnects, the room is removed from process memory
- the content remains in Redis until its TTL expires.

the WebSocket handshake also uses the configured `ALLOWED_ORIGINS` policy

opening a WebSocket connection is also a `GET` request, so the HTTP GET rate limit applies to the handshake. messages sent after the connection is established use the separate per connection WebSocket limiter.

## limits and configuration

The example `.env` contains the following defaults:

| variable | default value | purpose |
| --- | --- | --- |
| `APP_ENV` | `development` | application environment label |
| `SERVER_PORT` | `8080` | port used by the API inside the container |
| `HOST_PORT` | `8082` | host-side port used by Docker Compose |
| `REDIS_ADDR` | `redis-local:6379` | Redis address used by the API |
| `REDIS_PASSWORD` | `pwd` in the local example | Redis password |
| `ALLOWED_ORIGINS` | `http://localhost:3000` | allowed HTTP and WebSocket origins |
| `CODE_LENGTH` | `6` | required length of napkin codes |
| `MAX_CONTENT_LENGTH` | `400` | maximum content size in bytes |
| `DEFAULT_TTL` | `24h` | how long a napkin remains in Redis after saving |
| `RATE_LIMIT_POST_MAX` | `5` | maximum `POST /save` requests per window and IP |
| `RATE_LIMIT_GET_MAX` | `100` | maximum GET requests per window and IP |
| `RATE_LIMIT_WINDOW_SEC` | `60` | HTTP rate-limit window in seconds |

the HTTP rate limit is implemented using Redis

the IP is obtained from `CF-Connecting-IP`, then the first value of `X-Forwarded-For`, and finally `RemoteAddr`. these headers should only be trusted when the API is behind a trusted proxy.

the WebSocket message limiter is local to each connection and is not shared through Redis.

## known limitations

this is a deliberately small first version. some important limitations are currently intentional:

- real time collaboration does not scale beyond one API instance since rooms and connected clients live in process memory.
- there is no authentication or ownership system
- anyone who knows a valid code can read and overwrite the napkin
- there is no revision history, undo, cursor or selection state
- messages replace the whole document instead of describing deltas or operations
- there are no user identities, online-user events or presence information
- there is no conflict-resolution strategy or version checking
- the server does not send room history when a client connects
- Redis currently stores only the latest content, not an edit history
- the last write to Redis wins when multiple messages are saved close together
- napkins expire automatically and cannot currently be recovered after expiration.

supporting multiple instances would require a shared real-time coordination mechanism, such as Redis Pub/Sub, streams or another message broker. presence, deltas and conflict resolution would also require a richer message protocol than the current plain-text frames.

