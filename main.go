package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/libp2p/go-libp2p"
	peerstore "github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	multiaddr "github.com/multiformats/go-multiaddr"
)

func main() {
	// Start a libp2p node that listens on a random local TCP port (since 0 at end),
	// but without running the built-in ping protocol
	node, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/192.168.15.38/tcp/0"),
		libp2p.Ping(false),
	)
	if err != nil {
		panic(err)
	}

	// Configure our own ping protocol
	pingService := &ping.PingService{Host: node}
	node.SetStreamHandler(ping.ID, pingService.PingHandler)

	// To print the node's PeerInfo in multiaddr format
	peerInfo := peerstore.AddrInfo{
		ID:    node.ID(),
		Addrs: node.Addrs(),
	}
	addrs, err := peerstore.AddrInfoToP2pAddrs(&peerInfo)
	if err != nil {
		panic(err)
	}

	// Print the libp2p node address
	fmt.Println("libp2p node address:", addrs[0])

	// If a remote peer has been passed on the command line, connect to it
	// and send it 5 ping messages, otherwise wait for a signal to stop
	if len(os.Args) > 1 {
		addr, err := multiaddr.NewMultiaddr(os.Args[1])
		if err != nil {
			panic(err)
		}

		peer, err := peerstore.AddrInfoFromP2pAddr(addr)
		if err != nil {
			panic(err)
		}

		if err := node.Connect(context.Background(), *peer); err != nil {
			panic(err)
		}
		fmt.Println("\nSending 5 ping messages to", addr)

		ch := pingService.Ping(context.Background(), peer.ID)
		for i := 0; i < 5; i++ {
			res := <-ch
			fmt.Printf("\n%v. Pinged %v in %v", i+1, addr, res.RTT)
		}
	} else {
		// Wait for a SIGINT or SIGTERM signal
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		signalReceiver := <-ch
		fmt.Printf("\n--Received signal '%v', shutting down...\n", signalReceiver.String())
	}

	// Shut the node down
	if err := node.Close(); err != nil {
		panic(err)
	}
}
