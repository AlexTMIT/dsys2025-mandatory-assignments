# Ricart–Agrawala

tiny Ricart–Agrawala mutual exclusion implementation over gRPC.  
each node runs as a separate process, exchanges `Request`/`Reply` messages, and arbitrates entry to a demo cs which i emulate via prints.

---

## Prerequisites

- **Go** 1.21+ (tested with 1.22)
- **protoc** + Go plugins (only needed if you edit `grpc/ra.proto`):

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
export PATH="$(go env GOPATH)/bin:$PATH"
```

---

## Project layout

```
.
├── go.mod
├── go.sum
├── grpc
│   ├── ra.pb.go
│   ├── ra.proto
│   └── ra_grpc.pb.go
└── node
    └── node.go
```

---

## Generate stubs (only if you change the proto)

```bash
protoc -I grpc \
  --go_out=grpc --go_opt=paths=source_relative \
  --go-grpc_out=grpc --go-grpc_opt=paths=source_relative \
  grpc/ra.proto
```

then ensure dependencies are in place:

```bash
go mod tidy
```

---

## How to start the system

you can run from the project root using `go run ./node ...`  
(or `cd node && go run . ...` if you prefer).

### Start at least three nodes (the example below uses four)

**Terminal 1**
```bash
go run ./node --port=:50051 --peers=:50052,:50053,:50054
```

**Terminal 2**
```bash
go run ./node --port=:50052 --peers=:50051,:50053,:50054
```

**Terminal 3**
```bash
go run ./node --port=:50053 --peers=:50051,:50052,:50054
```

**Terminal 4**
```bash
go run ./node --port=:50054 --peers=:50051,:50052,:50053
```

> nodes block until **all peers are reachable** before beginning CS requests, so you can start them in any order.

---

## Demonstration logs

### Node on `:50051`
```text
(base) atm@MacBookPro node % go run . --port=:50051 --peers=:50052,:50053,:50054
2025/11/09 20:59:34 [localhost:50051] waiting for peers: 0/3 reachable
2025/11/09 20:59:34 [node localhost:50051] listening
2025/11/09 20:59:35 [localhost:50051] connected -> localhost:50052
2025/11/09 20:59:36 [localhost:50051] connected -> localhost:50053
2025/11/09 20:59:37 [localhost:50051] connected -> localhost:50054
2025/11/09 20:59:39 [localhost:50051] all peers reachable (3/3); starting
2025/11/09 20:59:42 [localhost:50051] REQUEST CS
2025/11/09 20:59:42 [localhost:50051] requesting CS (ts=0)
2025/11/09 20:59:42 [localhost:50051] recv Reply #1/3 | ts 1->4
2025/11/09 20:59:42 [localhost:50051] recv Reply #2/3 | ts 4->5
2025/11/09 20:59:42 [localhost:50051] recv Reply #3/3 | ts 5->6
2025/11/09 20:59:42 [localhost:50051] ENTER  CS
2025/11/09 20:59:43 [localhost:50051] recv Request from localhost:50052 -> DEFER (my reqTs=1 < their ts=4) | ts 6->7
2025/11/09 20:59:43 [localhost:50051] EXIT   CS
2025/11/09 20:59:43 [localhost:50051] releasing deferred reply to localhost:50052
2025/11/09 20:59:43 [localhost:50051] send Reply -> localhost:50052 (ts 7->8)
2025/11/09 20:59:44 [localhost:50051] recv Request from localhost:50053 -> REPLY immediately (my reqTs=1, their ts=7) | ts 8->9
2025/11/09 20:59:44 [localhost:50051] send Reply -> localhost:50053 (ts 9->10)
2025/11/09 20:59:45 [localhost:50051] recv Request from localhost:50054 -> REPLY immediately (my reqTs=1, their ts=10) | ts 10->11
2025/11/09 20:59:45 [localhost:50051] send Reply -> localhost:50054 (ts 11->12)
2025/11/09 20:59:46 [localhost:50051] REQUEST CS
2025/11/09 20:59:46 [localhost:50051] requesting CS (ts=12)
2025/11/09 20:59:46 [localhost:50051] recv Reply #1/3 | ts 13->16
2025/11/09 20:59:46 [localhost:50051] recv Reply #2/3 | ts 16->17
2025/11/09 20:59:47 [localhost:50051] recv Reply #3/3 | ts 17->18
2025/11/09 20:59:47 [localhost:50051] ENTER  CS
2025/11/09 20:59:47 [localhost:50051] recv Request from localhost:50052 -> DEFER (my reqTs=13 < their ts=16) | ts 18->19
2025/11/09 20:59:48 [localhost:50051] EXIT   CS
2025/11/09 20:59:48 [localhost:50051] releasing deferred reply to localhost:50052
2025/11/09 20:59:48 [localhost:50051] send Reply -> localhost:50052 (ts 19->20)
2025/11/09 20:59:48 [localhost:50051] recv Request from localhost:50053 -> REPLY immediately (my reqTs=13, their ts=19) | ts 20->21
2025/11/09 20:59:48 [localhost:50051] send Reply -> localhost:50053 (ts 21->22)
^Csignal: interrupt
(base) atm@MacBookPro node % 
```

### Node on `:50052`
```text
(base) atm@MacBookPro node % go run . --port=:50052 --peers=:50051,:50053,:50054
2025/11/09 20:59:35 [localhost:50052] waiting for peers: 0/3 reachable
2025/11/09 20:59:35 [node localhost:50052] listening
2025/11/09 20:59:35 [localhost:50052] connected -> localhost:50051
2025/11/09 20:59:36 [localhost:50052] connected -> localhost:50053
2025/11/09 20:59:37 [localhost:50052] connected -> localhost:50054
2025/11/09 20:59:40 [localhost:50052] all peers reachable (3/3); starting
2025/11/09 20:59:42 [localhost:50052] recv Request from localhost:50051 -> REPLY immediately (my reqTs=0, their ts=1) | ts 0->2
2025/11/09 20:59:42 [localhost:50052] send Reply -> localhost:50051 (ts 2->3)
2025/11/09 20:59:43 [localhost:50052] REQUEST CS
2025/11/09 20:59:43 [localhost:50052] requesting CS (ts=3)
2025/11/09 20:59:43 [localhost:50052] recv Reply #1/3 | ts 4->7
2025/11/09 20:59:43 [localhost:50052] recv Reply #2/3 | ts 7->8
2025/11/09 20:59:43 [localhost:50052] recv Reply #3/3 | ts 8->9
2025/11/09 20:59:43 [localhost:50052] ENTER  CS
2025/11/09 20:59:44 [localhost:50052] recv Request from localhost:50053 -> DEFER (my reqTs=4 < their ts=7) | ts 9->10
2025/11/09 20:59:44 [localhost:50052] EXIT   CS
2025/11/09 20:59:44 [localhost:50052] releasing deferred reply to localhost:50053
2025/11/09 20:59:44 [localhost:50052] send Reply -> localhost:50053 (ts 10->11)
2025/11/09 20:59:45 [localhost:50052] recv Request from localhost:50054 -> REPLY immediately (my reqTs=4, their ts=10) | ts 11->12
2025/11/09 20:59:45 [localhost:50052] send Reply -> localhost:50054 (ts 12->13)
2025/11/09 20:59:46 [localhost:50052] recv Request from localhost:50051 -> REPLY immediately (my reqTs=4, their ts=13) | ts 13->14
2025/11/09 20:59:46 [localhost:50052] send Reply -> localhost:50051 (ts 14->15)
2025/11/09 20:59:47 [localhost:50052] REQUEST CS
2025/11/09 20:59:47 [localhost:50052] requesting CS (ts=15)
2025/11/09 20:59:47 [localhost:50052] recv Reply #1/3 | ts 16->19
2025/11/09 20:59:47 [localhost:50052] recv Reply #2/3 | ts 19->20
2025/11/09 20:59:48 [localhost:50052] recv Reply #3/3 | ts 20->21
2025/11/09 20:59:48 [localhost:50052] ENTER  CS
2025/11/09 20:59:48 [localhost:50052] recv Request from localhost:50053 -> DEFER (my reqTs=16 < their ts=19) | ts 21->22
2025/11/09 20:59:49 [localhost:50052] EXIT   CS
2025/11/09 20:59:49 [localhost:50052] releasing deferred reply to localhost:50053
2025/11/09 20:59:49 [localhost:50052] send Reply -> localhost:50053 (ts 22->23)
2025/11/09 20:59:50 [localhost:50052] recv Request from localhost:50054 -> REPLY immediately (my reqTs=16, their ts=22) | ts 23->24
2025/11/09 20:59:50 [localhost:50052] send Reply -> localhost:50054 (ts 24->25)
^Csignal: interrupt
(base) atm@MacBookPro node % 
```

### Node on `:50053`
```text
(base) atm@MacBookPro node % go run . --port=:50053 --peers=:50051,:50052,:50054
2025/11/09 20:59:36 [localhost:50053] waiting for peers: 0/3 reachable
2025/11/09 20:59:36 [node localhost:50053] listening
2025/11/09 20:59:36 [localhost:50053] connected -> localhost:50052
2025/11/09 20:59:36 [localhost:50053] connected -> localhost:50051
2025/11/09 20:59:37 [localhost:50053] connected -> localhost:50054
2025/11/09 20:59:41 [localhost:50053] all peers reachable (3/3); starting
2025/11/09 20:59:42 [localhost:50053] recv Request from localhost:50051 -> REPLY immediately (my reqTs=0, their ts=1) | ts 0->2
2025/11/09 20:59:42 [localhost:50053] send Reply -> localhost:50051 (ts 2->3)
2025/11/09 20:59:43 [localhost:50053] recv Request from localhost:50052 -> REPLY immediately (my reqTs=0, their ts=4) | ts 3->5
2025/11/09 20:59:43 [localhost:50053] send Reply -> localhost:50052 (ts 5->6)
2025/11/09 20:59:44 [localhost:50053] REQUEST CS
2025/11/09 20:59:44 [localhost:50053] requesting CS (ts=6)
2025/11/09 20:59:44 [localhost:50053] recv Reply #1/3 | ts 7->10
2025/11/09 20:59:44 [localhost:50053] recv Reply #2/3 | ts 10->11
2025/11/09 20:59:44 [localhost:50053] recv Reply #3/3 | ts 11->12
2025/11/09 20:59:44 [localhost:50053] ENTER  CS
2025/11/09 20:59:45 [localhost:50053] recv Request from localhost:50054 -> DEFER (my reqTs=7 < their ts=10) | ts 12->13
2025/11/09 20:59:45 [localhost:50053] EXIT   CS
2025/11/09 20:59:45 [localhost:50053] releasing deferred reply to localhost:50054
2025/11/09 20:59:45 [localhost:50053] send Reply -> localhost:50054 (ts 13->14)
2025/11/09 20:59:46 [localhost:50053] recv Request from localhost:50051 -> REPLY immediately (my reqTs=7, their ts=13) | ts 14->15
2025/11/09 20:59:46 [localhost:50053] send Reply -> localhost:50051 (ts 15->16)
2025/11/09 20:59:47 [localhost:50053] recv Request from localhost:50052 -> REPLY immediately (my reqTs=7, their ts=16) | ts 16->17
2025/11/09 20:59:47 [localhost:50053] send Reply -> localhost:50052 (ts 17->18)
2025/11/09 20:59:48 [localhost:50053] REQUEST CS
2025/11/09 20:59:48 [localhost:50053] requesting CS (ts=18)
2025/11/09 20:59:48 [localhost:50053] recv Reply #1/3 | ts 19->22
2025/11/09 20:59:48 [localhost:50053] recv Reply #2/3 | ts 22->23
2025/11/09 20:59:49 [localhost:50053] recv Reply #3/3 | ts 23->24
2025/11/09 20:59:49 [localhost:50053] ENTER  CS
2025/11/09 20:59:50 [localhost:50053] recv Request from localhost:50054 -> DEFER (my reqTs=19 < their ts=22) | ts 24->25
2025/11/09 20:59:50 [localhost:50053] EXIT   CS
2025/11/09 20:59:50 [localhost:50053] releasing deferred reply to localhost:50054
2025/11/09 20:59:50 [localhost:50053] send Reply -> localhost:50054 (ts 25->26)
^Csignal: interrupt
(base) atm@MacBookPro node % 
```

### Node on `:50054`
```text
(base) atm@MacBookPro node % go run . --port=:50054 --peers=:50051,:50052,:50053
2025/11/09 20:59:36 [localhost:50054] waiting for peers: 0/3 reachable
2025/11/09 20:59:36 [node localhost:50054] listening
2025/11/09 20:59:37 [localhost:50054] connected -> localhost:50052
2025/11/09 20:59:37 [localhost:50054] connected -> localhost:50051
2025/11/09 20:59:37 [localhost:50054] connected -> localhost:50053
2025/11/09 20:59:42 [localhost:50054] all peers reachable (3/3); starting
2025/11/09 20:59:42 [localhost:50054] recv Request from localhost:50051 -> REPLY immediately (my reqTs=0, their ts=1) | ts 0->2
2025/11/09 20:59:42 [localhost:50054] send Reply -> localhost:50051 (ts 2->3)
2025/11/09 20:59:43 [localhost:50054] recv Request from localhost:50052 -> REPLY immediately (my reqTs=0, their ts=4) | ts 3->5
2025/11/09 20:59:43 [localhost:50054] send Reply -> localhost:50052 (ts 5->6)
2025/11/09 20:59:44 [localhost:50054] recv Request from localhost:50053 -> REPLY immediately (my reqTs=0, their ts=7) | ts 6->8
2025/11/09 20:59:44 [localhost:50054] send Reply -> localhost:50053 (ts 8->9)
2025/11/09 20:59:45 [localhost:50054] REQUEST CS
2025/11/09 20:59:45 [localhost:50054] requesting CS (ts=9)
2025/11/09 20:59:45 [localhost:50054] recv Reply #1/3 | ts 10->13
2025/11/09 20:59:45 [localhost:50054] recv Reply #2/3 | ts 13->14
2025/11/09 20:59:45 [localhost:50054] recv Reply #3/3 | ts 14->15
2025/11/09 20:59:46 [localhost:50054] ENTER  CS
2025/11/09 20:59:46 [localhost:50054] recv Request from localhost:50051 -> DEFER (my reqTs=10 < their ts=13) | ts 15->16
2025/11/09 20:59:47 [localhost:50054] EXIT   CS
2025/11/09 20:59:47 [localhost:50054] releasing deferred reply to localhost:50051
2025/11/09 20:59:47 [localhost:50054] send Reply -> localhost:50051 (ts 16->17)
2025/11/09 20:59:47 [localhost:50054] recv Request from localhost:50052 -> REPLY immediately (my reqTs=10, their ts=16) | ts 17->18
2025/11/09 20:59:47 [localhost:50054] send Reply -> localhost:50052 (ts 18->19)
2025/11/09 20:59:48 [localhost:50054] recv Request from localhost:50053 -> REPLY immediately (my reqTs=10, their ts=19) | ts 19->20
2025/11/09 20:59:48 [localhost:50054] send Reply -> localhost:50053 (ts 20->21)
2025/11/09 20:59:50 [localhost:50054] REQUEST CS
2025/11/09 20:59:50 [localhost:50054] requesting CS (ts=21)
2025/11/09 20:59:50 [localhost:50054] recv Reply #1/3 | ts 22->26
2025/11/09 20:59:50 [localhost:50054] recv Reply #2/3 | ts 26->27
^Csignal: interrupt
(base) atm@MacBookPro node % ```