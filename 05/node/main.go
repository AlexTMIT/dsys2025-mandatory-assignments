package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	pb "auction/grpc"

	"google.golang.org/grpc"
)

const auctionDuration = 100 * time.Second

type Node struct {
	pb.UnimplementedAuctionServer
	id            int
	peers         []string
	leader        int
	mu            sync.Mutex
	highest       int32
	highestBidder string
	start         time.Time
}

func NewNode(id int, peers []string) *Node {
	leader := 999999
	for i := range peers {
		if i < leader {
			leader = i
		}
	}

	n := &Node{
		id:     id,
		peers:  peers,
		leader: leader,
		start:  time.Now(),
	}

	log.SetPrefix(fmt.Sprintf("[N%d] ", id))

	go func() {
		time.Sleep(auctionDuration)
		log.Print("[AUCTION] auction time limit reached; no more bids should be accepted")
	}()

	return n
}

func (n *Node) reportLeaderFailure(where string, err error) {
	log.Printf("[LEADER/FAIL] at=%s leaderID=%d leaderAddr=%s err=%v", where, n.leader, n.peers[n.leader], err)
	go n.tryElectNewLeader()
}

func (n *Node) tryElectNewLeader() {
	n.mu.Lock()
	oldLeader := n.leader
	n.mu.Unlock()

	log.Printf("[LEADER/ELECT] starting election (oldLeader=%d)", oldLeader)

	for id, addr := range n.peers {
		if id == oldLeader {
			continue
		}

		cctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		conn, err := grpc.NewClient(addr, grpc.WithInsecure())
		if err != nil {
			cancel()
			log.Printf("[LEADER/ELECT] candidateID=%d addr=%s unreachable (connect err=%v)", id, addr, err)
			continue
		}

		cli := pb.NewAuctionClient(conn)
		_, err = cli.Result(cctx, &pb.ResultRequest{})
		cancel()
		conn.Close()

		if err != nil {
			log.Printf("[LEADER/ELECT] candidateID=%d addr=%s unreachable (Result err=%v)", id, addr, err)
			continue
		}

		n.mu.Lock()
		n.leader = id
		n.mu.Unlock()

		log.Printf("[LEADER/ELECT] newLeaderID=%d addr=%s (oldLeader=%d)", id, addr, oldLeader)
		return
	}

	log.Printf("[LEADER/ELECT] no reachable replacement; keeping leaderID=%d", oldLeader)
}

func (n *Node) Bid(ctx context.Context, req *pb.BidRequest) (*pb.BidResponse, error) {
	// foward to leader
	if n.id != n.leader {
		log.Printf("[BID/FWD] bidder=%s amount=%d -> leaderID=%d leaderAddr=%s",
			req.Bidder, req.Amount, n.leader, n.peers[n.leader])

		conn, err := grpc.NewClient(n.peers[n.leader], grpc.WithInsecure())
		if err != nil {
			n.reportLeaderFailure("Bid(connect)", err)
			return &pb.BidResponse{Outcome: "fail: cannot reach leader"}, nil
		}
		defer conn.Close()

		cli := pb.NewAuctionClient(conn)
		resp, err := cli.Bid(ctx, req)
		if err != nil {
			n.reportLeaderFailure("Bid(RPC)", err)
			return &pb.BidResponse{Outcome: "fail: leader error"}, nil
		}

		log.Printf("[BID/FWD] response from leader: outcome=%s", resp.Outcome)
		return resp, nil
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	now := time.Since(n.start)
	if now > auctionDuration {
		log.Printf("[BID/LEADER] REJECT bidder=%s amount=%d reason=auction-ended (t=%.1fs)",
			req.Bidder, req.Amount, now.Seconds())
		return &pb.BidResponse{Outcome: "fail: auction ended"}, nil
	}

	if req.Amount <= n.highest {
		log.Printf("[BID/LEADER] REJECT bidder=%s amount=%d reason=too-low current=%d currentBidder=%s",
			req.Bidder, req.Amount, n.highest, n.highestBidder)
		return &pb.BidResponse{Outcome: "fail: too low"}, nil
	}

	n.highest = req.Amount
	n.highestBidder = req.Bidder

	log.Printf("[BID/LEADER] ACCEPT bidder=%s amount=%d -> newHighest=%d",
		req.Bidder, req.Amount, n.highest)

	highest := n.highest
	highestBidder := n.highestBidder

	for peerID, addr := range n.peers {
		if peerID == n.id {
			continue
		}
		go func(peerID int, addr string, h int32, b string) {
			log.Printf("[REPL/SEND] toNodeID=%d addr=%s highest=%d bidder=%s",
				peerID, addr, h, b)

			conn, err := grpc.NewClient(addr, grpc.WithInsecure())
			if err != nil {
				log.Printf("[REPL/FAIL] toNodeID=%d addr=%s stage=connect err=%v",
					peerID, addr, err)
				return
			}
			defer conn.Close()

			cli := pb.NewAuctionClient(conn)
			_, err = cli.Replicate(context.Background(),
				&pb.ReplicaMsg{
					Highest:       h,
					HighestBidder: b,
				})
			if err != nil {
				log.Printf("[REPL/FAIL] toNodeID=%d addr=%s stage=rpc err=%v",
					peerID, addr, err)
				return
			}

			log.Printf("[REPL/ACK] fromNodeID=%d", peerID)
		}(peerID, addr, highest, highestBidder)
	}

	return &pb.BidResponse{Outcome: "success"}, nil
}

func (n *Node) Replicate(ctx context.Context, msg *pb.ReplicaMsg) (*pb.Ack, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	oldHighest := n.highest
	oldBidder := n.highestBidder

	if msg.Highest > n.highest {
		n.highest = msg.Highest
		n.highestBidder = msg.HighestBidder

		log.Printf("[REPL/APPLY] %d/%s -> %d/%s",
			oldHighest, oldBidder, n.highest, n.highestBidder)
	} else {
		log.Printf("[REPL/IGNORE] stale incoming=%d/%s current=%d/%s",
			msg.Highest, msg.HighestBidder,
			n.highest, n.highestBidder)
	}

	return &pb.Ack{}, nil
}

func (n *Node) localCopy() *pb.ResultResponse {
	n.mu.Lock()
	ended := time.Since(n.start) > auctionDuration
	highest := n.highest
	bidder := n.highestBidder
	n.mu.Unlock()
	return &pb.ResultResponse{
		Ended:         ended,
		Highest:       highest,
		HighestBidder: bidder,
	}
}

func (n *Node) Result(ctx context.Context, _ *pb.ResultRequest) (*pb.ResultResponse, error) {
	if n.id != n.leader {
		log.Printf("[RESULT/FWD] -> leaderID=%d leaderAddr=%s",
			n.leader, n.peers[n.leader])

		cctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
		defer cancel()

		conn, err := grpc.NewClient(n.peers[n.leader], grpc.WithInsecure())
		if err == nil {
			cli := pb.NewAuctionClient(conn)
			res, err := cli.Result(cctx, &pb.ResultRequest{})
			conn.Close()
			if err == nil {
				log.Printf("[RESULT/FWD] from leader ended=%v highest=%d bidder=%s",
					res.Ended, res.Highest, res.HighestBidder)
				return res, nil
			}
			n.reportLeaderFailure("Result(RPC)", err)
		} else {
			n.reportLeaderFailure("Result(connect)", err)
		}

		log.Print("[RESULT/FWD] using LOCAL COPY (leader unreachable)")
		return n.localCopy(), nil
	}

	ended := time.Since(n.start) > auctionDuration

	n.mu.Lock()
	highest := n.highest
	bidder := n.highestBidder
	n.mu.Unlock()

	log.Printf("[RESULT/LEADER] ended=%v highest=%d bidder=%s",
		ended, highest, bidder)

	return &pb.ResultResponse{
		Ended:         ended,
		Highest:       highest,
		HighestBidder: bidder,
	}, nil
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("usage: node <id> <peeraddr1> <peeraddr2> ...")
		return
	}

	id, _ := strconv.Atoi(os.Args[1])
	peers := os.Args[2:]

	node := NewNode(id, peers)

	lis, err := net.Listen("tcp", peers[id])
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	pb.RegisterAuctionServer(s, node)

	log.Printf("[START] listening=%s leaderID=%d leaderAddr=%s",
		peers[id], node.leader, peers[node.leader])

	s.Serve(lis)
}
