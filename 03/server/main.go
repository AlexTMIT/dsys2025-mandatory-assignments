package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	pb "chitchat/grpc"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedEchoServer

	mu      sync.Mutex
	clients map[uint64]chan *pb.Event
	clock   uint64
	nextID  uint64
}

func newServer() *server {
	return &server{
		clients: make(map[uint64]chan *pb.Event),
		clock:   0,
		nextID:  1,
	}
}

func (s *server) tick(c uint64) uint64 {
	s.mu.Lock()
	if c > s.clock {
		s.clock = c
	}
	s.clock++
	lamport := s.clock
	s.mu.Unlock()
	return lamport
}

func (s *server) broadcast(ev *pb.Event) {
	s.mu.Lock()
	for id, ch := range s.clients {
		select {
		case ch <- ev:
			log.Printf("[SERVER] [L=%d] [DELIVER to client=%d]", ev.GetLamport(), id)
		default: // for non blocking
			log.Printf("[SERVER] [L=%d] [DELIVER_SKIPPED to client=%d]", ev.GetLamport(), id)
		}
	}
	s.mu.Unlock()
}

func (s *server) Join(ctx context.Context, req *pb.JoinRequest) (*pb.JoinReply, error) {
	s.mu.Lock()
	id := s.nextID
	s.nextID = s.nextID + 1
	s.mu.Unlock()

	clock := s.tick(req.GetClock())
	log.Printf("[CLIENT] [L=%d] [JOIN_RPC client=%d] [in=%d]", clock, id, req.GetClock())
	return &pb.JoinReply{ClientId: id, ServerClock: clock}, nil
}

func (s *server) Leave(ctx context.Context, req *pb.LeaveRequest) (*pb.LeaveReply, error) {
	id := req.ClientId

	ch, ok := s.clients[id]
	if ok {
		removeClient(s, id, ch)
	}

	clock := s.tick(req.GetClock())
	log.Printf("[SERVER] [L=%d] [LEAVE_RPC client=%d] [in=%d]", clock, id, req.GetClock())
	s.broadcast(&pb.Event{
		Type:     pb.EventType_EVENT_LEAVE,
		ClientId: id,
		Text:     fmt.Sprintf("Participant %d left Chit Chat at logical time %d", id, clock),
		Lamport:  clock,
	})

	return &pb.LeaveReply{ServerClock: clock}, nil
}

func (s *server) Subscribe(req *pb.SubscribeRequest, stream pb.Echo_SubscribeServer) error {
	id := req.GetClientId()
	ch := make(chan *pb.Event, 128)

	s.mu.Lock()
	s.clients[id] = ch
	s.mu.Unlock()

	clock := s.tick(req.GetClock())

	log.Printf("[SERVER] [L=%d] [BROADCAST_JOIN client=%d]", clock, id)
	s.broadcast(&pb.Event{
		Type:     pb.EventType_EVENT_JOIN,
		ClientId: id,
		Text:     fmt.Sprintf("Participant %d joined to Chit Chat at logical time %d", id, clock),
		Lamport:  clock,
	})

	for {
		select {
		case ev, ok := <-ch:
			if !ok {
				log.Printf("[SERVER] [DISCONNECT client=%d] [reason=channel_closed]", id)
				return nil
			}
			err := stream.Send(ev)
			if err != nil {
				log.Printf("[SERVER] [DISCONNECT client=%d] [reason=send_error] [%v]", id, err)
				removeClient(s, id, ch)
				return err
			}
		case <-stream.Context().Done():
			log.Printf("[SERVER] [DISCONNECT client=%d] [reason=context_done]", id)
			s.broadcast(&pb.Event{
				Type:     pb.EventType_EVENT_LEAVE,
				ClientId: id,
				Text:     fmt.Sprintf("Participant %d left unexpectedly at logical time %d", id, clock),
				Lamport:  clock,
			})
			removeClient(s, id, ch)
			return nil
		}
	}
}

func removeClient(s *server, id uint64, ch chan *pb.Event) {
	s.mu.Lock()
	delete(s.clients, id)
	close(ch)
	s.mu.Unlock()
}

func (s *server) Publish(ctx context.Context, req *pb.PublishRequest) (*pb.PublishReply, error) {
	if len([]rune(req.GetText())) > 128 {
		return nil, fmt.Errorf("message too long (>128 UTF-8 chars)")
	}
	clock := s.tick(req.GetClock())
	log.Printf("[CLIENT] [L=%d] [PUBLISH_RPC from=%d] [in=%d]", req.GetClock(), req.GetClientId(), clock)
	s.broadcast(&pb.Event{
		Type:     pb.EventType_EVENT_USER_MSG,
		ClientId: req.GetClientId(),
		Text:     fmt.Sprintf("Participant %d at logical time %d: %s", req.ClientId, req.Clock, req.GetText()),
		Lamport:  clock,
	})
	log.Printf("[SERVER] [L=%d] [BROADCAST_MSG from=%d]", clock, req.GetClientId())
	return &pb.PublishReply{ServerClock: clock}, nil
}

func main() {
	addr := "127.0.0.1:50051"

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("[SERVER] [STARTUP_FAILED] [%v]", err)
	}

	gs := grpc.NewServer()
	pb.RegisterEchoServer(gs, newServer())

	log.Printf("[SERVER] [STARTUP] [addr=%s]", addr)
	defer log.Printf("[SERVER] [SHUTDOWN]")

	err = gs.Serve(lis)
	if err != nil {
		log.Fatalf("[SERVER] [SERVE_ERROR] [%v]", err)
	}
}
