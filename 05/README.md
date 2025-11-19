Run the cluster
---------------
In three separate terminals, start the three nodes:

cd node

# Terminal 1 – node 0 (initial leader)
go run main.go 0 localhost:5000 localhost:5001 localhost:5002

# Terminal 2 – node 1
go run main.go 1 localhost:5000 localhost:5001 localhost:5002

# Terminal 3 – node 2
go run main.go 2 localhost:5000 localhost:5001 localhost:5002

Each node will print log lines about bids, replication, and leader election.
Stop a node with Ctrl+C to trigger leader failover.

Send bids
---------
Use the client to place bids against any node:

cd client

# Alice bids 900 via node 1
go run client.go localhost:5001 alice 900

# Bob bids 1000 via node 2
go run client.go localhost:5002 bob 1000

The client prints whether the bid succeeded and the current highest bid/result.
Stop the client with Ctrl+C or let it exit after the response.