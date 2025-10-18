# Chit Chat - alext

## Prerequisites

- Go 1.21+ (tested with 1.22)
- `protoc` + plugins:
  ```bash
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  export PATH="$(go env GOPATH)/bin:$PATH"
  ```

## Project layout

```
project-root/
  client/        # client program
  server/        # server program
  grpc/          # .proto and generated code
  readme.md
  go.mod
```

## Build / generate

If you ever modify `grpc/echo.proto`, regenerate stubs:

```bash
protoc -I grpc \
  --go_out=grpc --go_opt=paths=source_relative \
  --go-grpc_out=grpc --go-grpc_opt=paths=source_relative \
  grpc/echo.proto
```

Fetch deps and tidy:

```bash
go get google.golang.org/grpc@latest
go get google.golang.org/protobuf@latest
go mod tidy
```

## Run

### 1) Start the server

```bash
go run ./server
```

You should see startup and subsequent logs, e.g.:

```
2025/10/18 17:31:18 [SERVER] [STARTUP] [addr=127.0.0.1:50051]
```

### 2) Start clients (in separate terminals)

```bash
go run ./client
```

The client accepts simple commands on stdin:

- `join` — join and subscribe to the stream (needed after a `leave`)
- `<anything else>` — publish a message (only when joined)
- `leave` — leave (client remains running; you may `join` again)

## What it looks like (proof of requirements)

### Server (one instance)

```
% go run ./server
2025/10/18 17:31:18 [SERVER] [STARTUP] [addr=127.0.0.1:50051]
2025/10/18 17:31:20 [CLIENT] [L=1] [JOIN_RPC client=1] [in=0]
2025/10/18 17:31:20 [SERVER] [L=2] [BROADCAST_JOIN client=1]
2025/10/18 17:31:20 [SERVER] [L=2] [DELIVER to client=1]
2025/10/18 17:31:22 [CLIENT] [L=3] [JOIN_RPC client=2] [in=0]
2025/10/18 17:31:22 [SERVER] [L=4] [BROADCAST_JOIN client=2]
2025/10/18 17:31:22 [SERVER] [L=4] [DELIVER to client=1]
2025/10/18 17:31:22 [SERVER] [L=4] [DELIVER to client=2]
2025/10/18 17:31:34 [CLIENT] [L=6] [PUBLISH_RPC from=1] [in=7]
2025/10/18 17:31:34 [SERVER] [L=7] [DELIVER to client=2]
2025/10/18 17:31:34 [SERVER] [L=7] [DELIVER to client=1]
2025/10/18 17:31:34 [SERVER] [L=7] [BROADCAST_MSG from=1]
2025/10/18 17:31:47 [CLIENT] [L=9] [PUBLISH_RPC from=2] [in=10]
2025/10/18 17:31:47 [SERVER] [L=10] [DELIVER to client=2]
2025/10/18 17:31:47 [SERVER] [L=10] [DELIVER to client=1]
2025/10/18 17:31:47 [SERVER] [L=10] [BROADCAST_MSG from=2]
2025/10/18 17:31:58 [CLIENT] [L=12] [PUBLISH_RPC from=1] [in=13]
2025/10/18 17:31:58 [SERVER] [L=13] [DELIVER to client=1]
2025/10/18 17:31:58 [SERVER] [L=13] [DELIVER to client=2]
2025/10/18 17:31:58 [SERVER] [L=13] [BROADCAST_MSG from=1]
2025/10/18 17:32:01 [SERVER] [L=16] [LEAVE_RPC client=1] [in=15]
2025/10/18 17:32:01 [SERVER] [DISCONNECT client=1] [reason=channel_closed]
2025/10/18 17:32:01 [SERVER] [L=16] [DELIVER to client=2]
2025/10/18 17:32:15 [CLIENT] [L=18] [PUBLISH_RPC from=2] [in=19]
2025/10/18 17:32:15 [SERVER] [L=19] [DELIVER to client=2]
2025/10/18 17:32:15 [SERVER] [L=19] [BROADCAST_MSG from=2]
2025/10/18 17:32:18 [CLIENT] [L=20] [JOIN_RPC client=3] [in=15]
2025/10/18 17:32:18 [SERVER] [L=21] [BROADCAST_JOIN client=3]
2025/10/18 17:32:18 [SERVER] [L=21] [DELIVER to client=3]
2025/10/18 17:32:18 [SERVER] [L=21] [DELIVER to client=2]
2025/10/18 17:32:22 [CLIENT] [L=23] [PUBLISH_RPC from=3] [in=24]
2025/10/18 17:32:22 [SERVER] [L=24] [DELIVER to client=3]
2025/10/18 17:32:22 [SERVER] [L=24] [DELIVER to client=2]
2025/10/18 17:32:22 [SERVER] [L=24] [BROADCAST_MSG from=3]
2025/10/18 17:32:30 [CLIENT] [L=26] [PUBLISH_RPC from=2] [in=27]
2025/10/18 17:32:30 [SERVER] [L=27] [DELIVER to client=3]
2025/10/18 17:32:30 [SERVER] [L=27] [DELIVER to client=2]
2025/10/18 17:32:30 [SERVER] [L=27] [BROADCAST_MSG from=2]
2025/10/18 17:32:31 [SERVER] [L=30] [LEAVE_RPC client=2] [in=29]
2025/10/18 17:32:31 [SERVER] [DISCONNECT client=2] [reason=channel_closed]
2025/10/18 17:32:31 [SERVER] [L=30] [DELIVER to client=3]
2025/10/18 17:32:34 [SERVER] [L=33] [LEAVE_RPC client=3] [in=32]
2025/10/18 17:32:34 [SERVER] [DISCONNECT client=3] [reason=channel_closed]
```

### Client A (id=1 then leaves, rejoins as id=3)

```
% go run ./client
2025/10/18 17:31:20 Participant 1 joined to Chit Chat at logical time 2
2025/10/18 17:31:22 Participant 2 joined to Chit Chat at logical time 4
Hey, I'm client number 1!
2025/10/18 17:31:34 Participant 1 at logical time 6: Hey, I'm client number 1!
2025/10/18 17:31:47 Participant 2 at logical time 9: Nice to meet you, client number 1. I'm clclient number two.
Cool. Dinner is ready, I must now go.
2025/10/18 17:31:58 Participant 1 at logical time 12: Cool. Dinner is ready, I must now go.
leave
2025/10/18 17:32:01 [CLIENT] [LEAVE ok]
join 
2025/10/18 17:32:18 [CLIENT] [JOIN ok id=3] [server_clock=20]
2025/10/18 17:32:18 Participant 3 joined to Chit Chat at logical time 21
I'm back!
2025/10/18 17:32:22 Participant 3 at logical time 23: I'm back!
2025/10/18 17:32:30 Participant 2 at logical time 26: It's late. Let's go to bed.
2025/10/18 17:32:31 Participant 2 left Chit Chat at logical time 30
leave
2025/10/18 17:32:34 [CLIENT] [LEAVE ok]

```

### Client B (id=2)

```
% go run ./client
2025/10/18 17:31:22 Participant 2 joined to Chit Chat at logical time 4
2025/10/18 17:31:34 Participant 1 at logical time 6: Hey, I'm client number 1!
Nice to meet you, client number 1. I'm clclient number two.
2025/10/18 17:31:47 Participant 2 at logical time 9: Nice to meet you, client number 1. I'm clclient number two.
2025/10/18 17:31:58 Participant 1 at logical time 12: Cool. Dinner is ready, I must now go.
2025/10/18 17:32:01 Participant 1 left Chit Chat at logical time 16
Oh ok. Enjoy dinner!
2025/10/18 17:32:15 Participant 2 at logical time 18: Oh ok. Enjoy dinner!
2025/10/18 17:32:18 Participant 3 joined to Chit Chat at logical time 21
2025/10/18 17:32:22 Participant 3 at logical time 23: I'm back!
It's late. Let's go to bed.
2025/10/18 17:32:30 Participant 2 at logical time 26: It's late. Let's go to bed.
leave
2025/10/18 17:32:31 [CLIENT] [LEAVE ok]

```