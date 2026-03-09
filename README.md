# Image Blur Service

A distributed image processing service written in Go that accepts image blur requests via gRPC, queues them in Redis, and processes them asynchronously with a worker that saves results to PostgreSQL.

## Architecture

```
Client
  |
  | gRPC (QueueImage RPC)
  v
gRPC Server (cmd/server)
  |
  | Redis RPush
  v
Redis Queue (image_blur_queue)
  |
  | Redis BLPop
  v
Worker (cmd/worker)
  |-- Applies Gaussian blur (disintegration/imaging)
  |-- Saves blurred image to ASSETS_DIRECTORY
  |-- Stores metadata in PostgreSQL (blurred_img table)
```

## Project Structure

```
image-blur-service/
├── api/
│   ├── img.proto           # gRPC service definition
│   ├── img.pb.go           # Generated protobuf types
│   └── img_grpc.pb.go      # Generated gRPC stubs
├── cmd/
│   ├── server/
│   │   └── main.go         # gRPC server (port :50051)
│   └── worker/
│       ├── main.go         # Async image processing worker
│       └── main_test.go    # Benchmarks and tests
├── internal/
│   ├── processor/
│   │   └── blur.go         # Core blur logic (imaging library)
│   ├── queue/
│   │   └── queue.go        # Redis queue (enqueue/dequeue)
│   └── storage/
│       └── db.go           # In-memory mock storage
├── platforms/
│   └── helper/
│       └── helper.go       # Random string generator
├── go.mod
├── go.sum
└── .env
```

## gRPC API

Defined in `api/img.proto`:

| RPC              | Request                        | Response                        | Description                      |
|------------------|--------------------------------|---------------------------------|----------------------------------|
| `QueueImage`     | `ImageRequest { image_id }`    | `ImageResponse { status }`      | Enqueues an image for blurring   |
| `DeleteImage`    | `DeleteImageRequest { source_img_id }` | `DeleteImageResponse { success }` | Deletes a blurred image record |

## Components

### gRPC Server (`cmd/server`)
- Listens on `:50051`
- Accepts `QueueImage` requests and pushes the `image_id` onto the Redis queue
- Validates that `image_id` is non-empty before enqueuing

### Worker (`cmd/worker`)
- Connects to Redis and PostgreSQL on startup
- Runs two concurrent `ProcessImages` goroutines
- Dequeues image IDs from Redis using a blocking pop (`BLPop`)
- Applies Gaussian blur (sigma=8.0) using the `disintegration/imaging` library
- Saves the blurred image file to `ASSETS_DIRECTORY` with a random name
- Inserts a record into the `blurred_img` PostgreSQL table with the source ID and blurred file path

**Alternative blur method:** A `BlurImage` function using FFmpeg (`gblur=sigma=20`) is also implemented but currently commented out. It enforces a 20KB output size limit.

### Queue (`internal/queue`)
- Redis-backed queue using a list (`image_blur_queue`)
- `Enqueue(imageID)` — appends to the right of the list (`RPush`)
- `Dequeue()` — blocks until an item is available (`BLPop`)

### Processor (`internal/processor`)
- Stateless `BlurImage([]byte)` function that decodes an image, applies blur (sigma=5), and returns the result as JPEG bytes

### Storage (`internal/storage`)
- Thread-safe in-memory map tracking which image IDs have been blurred
- `Exists(imageID)` and `MarkAsBlurred(imageID)`

## Prerequisites

- Go 1.24+
- Redis (default: `localhost:6379`)
- PostgreSQL with a database and the following table:

```sql
CREATE TABLE blurred_img (
    source_img_id  TEXT,
    blurred_img_url TEXT
);
```

- (Optional) FFmpeg — only needed if using the FFmpeg-based blur path

## Configuration

Copy `.env` and set the values:

```env
ASSETS_DIRECTORY=/path/to/images
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_NAME=your_db
DB_PASSWORD=your_password
DATABASE_URL=postgres://postgres:password@localhost:5432/your_db?sslmode=disable
```

## Running

### Start the gRPC Server

```bash
cd cmd/server
go run main.go
```

The server starts on `:50051`.

### Start the Worker

```bash
cd cmd/worker
go run main.go
```

The worker connects to Redis and PostgreSQL, then continuously polls the queue for image jobs.

## Testing

```bash
cd cmd/worker
go test ./...
```

Run benchmarks:

```bash
go test -bench=. ./cmd/worker/
```

## Dependencies

| Package | Purpose |
|---------|---------|
| `google.golang.org/grpc` | gRPC framework |
| `google.golang.org/protobuf` | Protocol Buffers |
| `github.com/redis/go-redis/v9` | Redis client |
| `github.com/lib/pq` | PostgreSQL driver |
| `github.com/disintegration/imaging` | Gaussian blur |
| `github.com/subosito/gotenv` | `.env` file loader |
