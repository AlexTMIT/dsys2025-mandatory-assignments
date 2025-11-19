package main

import (
	"context"
	"log"
	"os"
	"strconv"

	pb "auction/grpc"

	"google.golang.org/grpc"
)

func main() {
	if len(os.Args) < 4 {
		log.Println("usage: client <node-adderss> <name> <amount|-1 for query>")
		return
	}

	log.SetFlags(0)

	addr := os.Args[1]
	bidder := os.Args[2]
	amount, _ := strconv.Atoi(os.Args[3])

	log.Printf("[CLIENT] target=%s bidder=%s amount=%d", addr, bidder, amount)

	conn, err := grpc.NewClient(addr, grpc.WithInsecure())
	if err != nil {
		log.Printf("[CLIENT] ERROR: cannot connect to %s: %v", addr, err)
		return
	}
	defer conn.Close()

	cli := pb.NewAuctionClient(conn)

	ctx := context.Background()

	if amount == -1 {
		log.Printf("[CLIENT] mode=RESULT-ONLY")
		res, err := cli.Result(ctx, &pb.ResultRequest{})
		if err != nil {
			log.Printf("[CLIENT] RESULT RPC error: %v", err)
			return
		}
		log.Printf("[CLIENT] RESULT ended=%v highest=%d highestBidder=%s",
			res.Ended, res.Highest, res.HighestBidder)
		return
	}

	log.Printf("[CLIENT] mode=BID+RESULT")
	resp, err := cli.Bid(ctx, &pb.BidRequest{
		Bidder: bidder,
		Amount: int32(amount),
	})
	if err != nil {
		log.Printf("[CLIENT] BID RPC error: %v", err)
		return
	}
	log.Printf("[CLIENT] BID outcome=%s", resp.Outcome)

	res, err := cli.Result(ctx, &pb.ResultRequest{})
	if err != nil {
		log.Printf("[CLIENT] RESULT RPC error: %v", err)
		return
	}
	log.Printf("[CLIENT] RESULT ended=%v highest=%d highestBidder=%s",
		res.Ended, res.Highest, res.HighestBidder)
}
