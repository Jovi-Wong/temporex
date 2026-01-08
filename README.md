# Kelutral

The nexus for players to sync in game - A real-time game frame synchronization server built with Go.

## Overview

Kelutral is a WebSocket-based game frame synchronization server that enables real-time multiplayer gaming by coordinating player actions across clients. It runs efficiently in Docker containers and provides ~60 FPS frame synchronization.

## Features

- **Real-time WebSocket communication** for low-latency game state sync
- **Frame synchronization** at ~60 FPS (16ms intervals)
- **Player connection management** with automatic cleanup
- **Health check endpoint** for monitoring
- **Docker support** for easy deployment
- **Lightweight** Alpine-based container image

## Quick Start

### Using Docker Compose (Recommended)

```bash
docker-compose up -d
```

The server will be available at `ws://localhost:8080/ws`

### Using Docker

```bash
# Build the image
docker build -t kelutral .

# Run the container
docker run -p 8080:8080 kelutral
```

### Running Locally

```bash
# Install dependencies
go mod download

# Run the server
go run main.go
```

## API Endpoints

### WebSocket Connection

**Endpoint:** `ws://localhost:8080/ws?player_id=<your_player_id>`

**Parameters:**
- `player_id` (optional): Unique identifier for the player. Auto-generated if not provided.

**Message Format:**
```json
{
  "player_id": "player1",
  "frame_num": 123,
  "timestamp": 1609459200000000000,
  "actions": {
    "move": "forward",
    "rotate": 45,
    "action": "jump"
  }
}
```

### Health Check

**Endpoint:** `GET http://localhost:8080/health`

**Response:**
```json
{
  "status": "ok",
  "players": 3
}
```

## Configuration

The server can be configured using environment variables:

- `PORT`: Server port (default: 8080)

Example:
```bash
PORT=9000 go run main.go
```

Or with Docker:
```bash
docker run -p 9000:9000 -e PORT=9000 kelutral
```

## Client Integration Example

```javascript
// Connect to the server
const ws = new WebSocket('ws://localhost:8080/ws?player_id=player1');

// Send frame data
ws.onopen = () => {
  setInterval(() => {
    ws.send(JSON.stringify({
      player_id: 'player1',
      frame_num: frameCounter++,
      actions: {
        x: playerX,
        y: playerY,
        action: currentAction
      }
    }));
  }, 16); // ~60 FPS
};

// Receive synchronized frames from other players
ws.onmessage = (event) => {
  const frameData = JSON.parse(event.data);
  updateGameState(frameData);
};
```

## Architecture

- **GameRoom**: Manages player connections and broadcasts frame updates
- **Player**: Represents individual WebSocket connections with send/receive pumps
- **FrameData**: Standard message format for game state synchronization

The server uses goroutines for concurrent player handling and a ticker for frame number progression.

## Development

### Build

```bash
go build -o kelutral
```

### Test

```bash
go test ./...
```

## License

See [LICENSE](LICENSE) file for details.
