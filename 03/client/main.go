package main

import (
	"bufio"
	"context"
	"log"
	"os"
	"time"

	pb "chitchat/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var id uint64
var clock uint64
var joined bool

var evCh = make(chan *pb.Event, 32)
var inCh = make(chan string, 1)

func main() {
	addr := "127.0.0.1:50051"

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("[CLIENT] [DIAL_ERROR] [%v]", err)
	}
	defer conn.Close()

	client := pb.NewEchoClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	joinResp, err := client.Join(ctx, &pb.JoinRequest{Clock: 0})
	if err != nil {
		log.Fatalf("[CLIENT] [JOIN_ERROR] [SERVER NOT STARTED]")
	}

	id = joinResp.ClientId
	clock = joinResp.ServerClock
	joined = true

	sub, err := client.Subscribe(context.Background(), &pb.SubscribeRequest{ClientId: id, Clock: clock})
	if err != nil {
		log.Fatalf("[CLIENT] [SUBSCRIBE_ERROR] [%v]", err)
	}

	go listenForStream(sub)
	go listenForInput()

	for {
		select {
		case line, ok := <-inCh:
			if !ok {
				log.Printf("[CLIENT] [STDIN_CLOSED]")
				return
			}
			handleIn(line, client)

		case ev, ok := <-evCh:
			if !ok {
				log.Printf("[CLIENT] [STREAM_CLOSED]")
				return
			}
			newClock := clock
			if ev.Lamport > newClock {
				newClock = ev.GetLamport()
			}
			newClock++
			clock = newClock
			log.Println(ev.Text)
		}
	}
}

func listenForStream(sub grpc.ServerStreamingClient[pb.Event]) {
	for {
		ev, err := sub.Recv()
		if err != nil {
			return
		}
		evCh <- ev
	}
}

func listenForInput() {
	func() {
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			inCh <- sc.Text()
		}
	}()
	log.Printf("\"leave\" to leave; anything else to post")
}

func handleIn(line string, client pb.EchoClient) {
	switch {
	case line == "leave":
		if !joined {
			log.Println("[CLIENT] not joined")
			return
		}
		clock++
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_, err := client.Leave(ctx, &pb.LeaveRequest{ClientId: id, Clock: clock})
		cancel()
		if err != nil {
			log.Printf("[CLIENT] [LEAVE_ERROR] [%v]", err)
		} else {
			log.Printf("[CLIENT] [LEAVE ok]")
		}
		joined = false
		return

	case line == "join":
		if joined {
			log.Println("[CLIENT] already joined")
			return
		}
		jctx, jcancel := context.WithTimeout(context.Background(), 5*time.Second)
		resp, err := client.Join(jctx, &pb.JoinRequest{Clock: clock})
		jcancel()
		if err != nil {
			log.Printf("[CLIENT] [JOIN_ERROR] [%v]", err)
			return
		}
		id = resp.GetClientId()
		clock = resp.GetServerClock()

		sub, err := client.Subscribe(context.Background(), &pb.SubscribeRequest{ClientId: id, Clock: clock})
		if err != nil {
			log.Printf("[CLIENT] [SUBSCRIBE_ERROR] [%v]", err)
			return
		}
		go func() {
			for {
				ev, err := sub.Recv()
				if err != nil {
					return
				}
				evCh <- ev
			}
		}()
		joined = true
		log.Printf("[CLIENT] [JOIN ok id=%d] [server_clock=%d]", id, clock)
		return
	}

	if len([]rune(line)) > 128 {
		log.Printf("[CLIENT] [PUBLISH skipped] [reason=too_long len=%d]", len([]rune(line)))
		return
	}
	if !joined {
		log.Println("[CLIENT] not joined; type 'join' first")
		return
	}

	clock++
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	_, err := client.Publish(ctx, &pb.PublishRequest{
		ClientId: id,
		Text:     line,
		Clock:    clock,
	})
	cancel()
	if err != nil {
		log.Printf("[CLIENT] [PUBLISH_ERROR] [%v]", err)
	}
}
