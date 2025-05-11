# LiveLite - Stream Management API

LiveLite is a lightweight API for managing LiveKit streaming rooms and connections, built with Go and Fiber.

## Features

- Room management (list/create rooms)
- Generate participant join tokens
- Get RTMP ingestion URLs
- Simple and lightweight

## Requirements

- Go 1.24.3+
- LiveKit server credentials
- Environment variables (see Configuration)

## Configuration

The following environment variables are required:

```bash
LIVEKIT_URL=your-livekit-server-url
LIVEKIT_API_KEY=your-api-key
LIVEKIT_API_SECRET=your-api-secret
```

## API Endpoints

### Room Management

- `GET /room` - List all active rooms
- `POST /room/:name` - Create a new room with the specified name

### Participant Access

- `GET /join/:room/:identity` - Generate a join token for a participant

### RTMP Ingestion

- `GET /rtmp/:room/:name/:identity` - Get RTMP ingestion URL for streaming

## Installation

1. Clone the repository:

```bash
git clone https://github.com/metalpoch/livelite.git
cd livelite
```

2. Set up environment variables:

```bash
export LIVEKIT_URL=your_url
export LIVEKIT_API_KEY=your_key
export LIVEKIT_API_SECRET=your_secret
```

3. Run the application:

```bash
go run main.go
```

The server will start on port 3000 by default.

## Development

To build and run:

```bash
go build -o livelite
./livelite
```

## License

[MIT](LICENSE)
