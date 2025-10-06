# tcp-sim

## 5 questions

a. What are packages in your implementation? What data structure do you use to transmit data and meta-data?
- the packages are the syn/syn+ack/ack messages exchanged between client and server.
- each package is represented as a fixed 20 byte array (a tcp header without options).  
- meta-data is encoded inside the header: src port, dst port, seq number, ack number, and flags.  
- this array is the data structure i send over udp.

b. Does your implementation use threads or processes? Why is it not realistic to use threads?
- my implementation runs one server process and one client process.  
- they communicate over udp sockets, hopefully somewhat like real distributed programs.  
- using threads would not be realistic because threads share memory inside one process and could bypass the network.  
- in real tcp, the client and server would probalby run on different machines (or at least in different processes), 
and must communicate only through message passing over a network. if i used threads, then there is no network,
and it is therefore not a distributed system, but a multithreaded program.

c. In case the network changes the order in which messages are delivered, how would you handle message re-ordering?
- i would sequence numbers as logical time to reorder messages correctly.  
- my simulation includes sequence numbers in the header, but i do not implement the reordering logic.  
- if messages came in the wrong order, my handshake would currently fail.
- however, messages are not sent without receiving one (except intial client message); so unordered messages here could not happen.

d. In case messages can be delayed or lost, how does your implementation handle message loss?
- really, i should detect loss with timeouts and possibly resend missing headers.  
- Conn has a function SetReadDeadline which could be used something like this in a for loop:

<pre>
c.Conn.WriteToUDP(out[:], dst)
c.Conn.SetReadDeadline(time.Now().Add(time.Second))
n, addr, err := c.Conn.ReadFromUDP(Buf)
if err != nil {
    fmt.Println("timeout, resending...")
    continue
}
// got response
copy(in[:], Buf[:n])
c.Conn.SetReadDeadline(time.Time{})
</pre>


- my simulation does not handle delays or loss. if a message is dropped, the handshake never completes.
- as an additional note, i should do a checksum for message tampering check for extra security.

e. Why is the 3-way handshake important?
- because it proves that both sides can send and receive packets. 
- and because it sets up a reliable connection before 'real' data is exchanged.

## how to run

1. cd into the 02 folder

   example:
<pre>
cd 02
</pre>

2. in one terminal start the server:

    go run ./server -port- -seq-

   example:
<pre>
go run ./server 9000 1234
</pre>

3. in another terminal start the client:

go run ./client -port- -seq-

   example:
<pre>
go run ./client 9000 2345
</pre>

## example server
<pre>
% go run ./server 9000 1234
Listening on port 9000

[SERVER: RECEIVED SYN]
 SrcPort: 64441  DstPort: 9000 
 Seq: 2345        Ack: 0         
 Flags: SYN
 Raw:  FB B9 23 28 00 00 09 29 
        00 00 00 00 00 02 00 00 
        00 00 00 00 

[SERVER: SENT SYNACK]
 SrcPort: 9000   DstPort: 64441
 Seq: 1234        Ack: 2346      
 Flags: SYN | ACK
 Raw:  23 28 FB B9 00 00 04 D2 
        00 00 09 2A 00 12 00 00 
        00 00 00 00 

[SERVER: RECEIVED ACK]
 SrcPort: 64441  DstPort: 9000 
 Seq: 2346        Ack: 1235      
 Flags: ACK
 Raw:  FB B9 23 28 00 00 09 2A 
        00 00 04 D3 00 10 00 00 
        00 00 00 00 

Handshake complete.%                                                                                                                                  
</pre>

## example client
<pre>
% go run ./client 9000 2345

[CLIENT: SENT SYN]
 SrcPort: 64441  DstPort: 9000 
 Seq: 2345        Ack: 0         
 Flags: SYN
 Raw:  FB B9 23 28 00 00 09 29 
        00 00 00 00 00 02 00 00 
        00 00 00 00 

[CLIENT: RECEIVED SYNACK]
 SrcPort: 9000   DstPort: 64441
 Seq: 1234        Ack: 2346      
 Flags: SYN | ACK
 Raw:  23 28 FB B9 00 00 04 D2 
        00 00 09 2A 00 12 00 00 
        00 00 00 00 

[CLIENT: SENT ACK]
 SrcPort: 64441  DstPort: 9000 
 Seq: 2346        Ack: 1235      
 Flags: ACK
 Raw:  FB B9 23 28 00 00 09 2A 
        00 00 04 D3 00 10 00 00 
        00 00 00 00 

Handshake complete.%                                                                                                                                  
</pre>