package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
)

var incommingconnected bool = false

const service string = "rezon/fileshare"

const topicPubSub string = "fileshare/rezon"

const proto string = "/file/1.1.0"

var recipientPeerId peer.AddrInfo

var option int

func establishNode() (host.Host, context.Context) {
	ctx := context.Background()
	host, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"))
	if err != nil {
		fmt.Println("Error while creating the node")
	} else {
		fmt.Println("Successfully created the node")
	}
	kad_dht, err := dht.New(ctx, host)
	if err != nil {
		fmt.Println("Error while creating a new DHT")
	} else {
		fmt.Println("Successfully created a new DHT")
	}
	err = kad_dht.Bootstrap(ctx)
	if err != nil {
		fmt.Println("Error while setting the DHT to bootstrap mode")
	} else {
		fmt.Println("Successfull in setting the DHT to Bootstrap mode")
	}
	var bootConnWait sync.WaitGroup
	for _, bootPeers := range dht.GetDefaultBootstrapPeerAddrInfos() {
		bootConnWait.Add(1)
		go func(bootPeers peer.AddrInfo) {
			defer bootConnWait.Done()
			err := host.Connect(ctx, bootPeers)
			if err != nil {
				fmt.Println("Error while connecting to bootstrap peer")
			} else {
				fmt.Println("Successfull in connecting to bootstrap peer :", bootPeers.ID)
			}
		}(bootPeers)
	}
	bootConnWait.Wait()
	fmt.Println("Connection to all bootstrap peers concluded")
	return host, ctx
}

func initDHT(host host.Host, ctx context.Context) *dht.IpfsDHT {
	kad_dht, err := dht.New(ctx, host)
	if err != nil {
		fmt.Println("Error while creating the new DHT")
	} else {
		fmt.Println("Successfull in creating the new DHT")
	}
	return kad_dht
}

func discoverPeers(host host.Host, ctx context.Context) {
	kad_dht := initDHT(host, ctx)
	routingDiscovery := drouting.NewRoutingDiscovery(kad_dht)
	dutil.Advertise(ctx, routingDiscovery, service)
	fmt.Println("This may take some time")
	isConnected := false
	for !isConnected {
		if incommingconnected {
			fmt.Println("Incomming stream interut")
			break
		}
		fmt.Println("TOTAL CONNECTION :", len(host.Network().Peers()))
		peerChan, err := routingDiscovery.FindPeers(ctx, service)
		if err != nil {
			fmt.Println("Error while finding peers for service", service)
		} else {
			// fmt.Println("Successfully found some peers")
		}
		for peerAddr := range peerChan {
			// fmt.Println("Trying to connect to peers")
			if peerAddr.ID == host.ID() {
				continue
			}
			if incommingconnected {
				fmt.Println("Incomming stream interupted")
				break
			}
			err := host.Connect(ctx, peerAddr)
			if err != nil {
				// fmt.Println("Error while connecting to peer :", peerAddr.ID)
			} else {
				fmt.Println("Successfully connected to peer :", peerAddr.ID)
				recipientPeerId = peerAddr
				isConnected = true
				break
			}
		}
	}
	fmt.Println("Done with peer discovery")
	if !incommingconnected {
		openNewStream(ctx, host)
	}
}

func newPubSub(ctx context.Context, host host.Host) (*pubsub.Subscription, *pubsub.Topic) {
	pubSub, err := pubsub.NewGossipSub(ctx, host)
	if err != nil {
		fmt.Println("Error while creating a new pubsub service")
	} else {
		fmt.Println("Successfully created a new pubsub service")
	}
	topic, err := pubSub.Join(topicPubSub)
	if err != nil {
		fmt.Println("Error while joining the topic", topicPubSub)
	} else {
		fmt.Println("Successfully joined the topic", topicPubSub)
	}
	subscription, err := topic.Subscribe()
	if err != nil {
		fmt.Println("Error while subscribing to ", topic)
	} else {
		fmt.Println("Successfully subscribing to ", topic)
	}
	return subscription, topic
}

// there are no issues in send to node
func sendToNode(name string, rdwt *bufio.ReadWriter) {
	file, err := os.Open(name)
	if err != nil {
		fmt.Println("Error while opening the file ", name)
	} else {
		fmt.Println("Successfully opened the file")
	}
	buffer := make([]byte, 10)
	for {
		_, err = file.Read(buffer)
		if err == io.EOF {
			fmt.Println("End of file reached - end signal sent to the stream")
			rdwt.Write([]byte("<END>"))
			return
		}
		if err != nil {
			fmt.Println("Error while reading from the file")
		}
		// fmt.Print(string(buffer))
		_, err := rdwt.Write(buffer)
		if err != nil {
			fmt.Println("Error while writting to the stream")
		} else {
			// fmt.Println(n, "bytes written to stream")
		}
	}

}

func getFromNode(rdwt *bufio.ReadWriter) {
	file, err := os.Create("recievedfile.txt")
	if err != nil {
		fmt.Println("Error while opening the reciving file")
	} else {
		fmt.Println("Successfully opened the recieving file")
	}
	buffer := make([]byte, 10)
	defer func() {
		fmt.Println("Succesfully recieved the complete file")
		doneWriting = true
	}()
	for {
		_, err := rdwt.Read(buffer)
		if err == io.EOF {
			file.Close()
			fmt.Println("End of buffer reached")
			return
		}
		if err != nil {
			fmt.Println("Error while reading from the buffer")
		}
		file.Write(buffer)
		// fmt.Print(string(buffer))
	}
}

func handleStream(stream network.Stream) {
	incommingconnected = true
	fmt.Println("Using default stream handler")
	fmt.Println("Succeesfully started a new stream")
	redWrite := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	if option == 1 {
		go sendToNode("main.go", redWrite)
	} else {
		go getFromNode(redWrite)
	}
	select {}

}

func openNewStream(ctx context.Context, host host.Host) {
	fmt.Println("Using open stream")
	stream, err := host.NewStream(ctx, recipientPeerId.ID, protocol.ID(proto))
	if err != nil {
		fmt.Println("Error while establishing a new stream to", recipientPeerId.ID)
	} else {
		fmt.Println("Successfull in establishing a new stream to", recipientPeerId.ID)
	}

	redWrite := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	if option == 1 {
		go sendToNode("main.go", redWrite)
	} else {
		go getFromNode(redWrite)
	}

}

func main() {
	host, ctx := establishNode()
	opt := flag.Int("m", 0, "0 for read and 1 for write")
	flag.Parse()
	option = *opt
	if option == 1 {
		fmt.Println("Set as file sender")
	} else {
		fmt.Println("set as file reader")
	}
	host.SetStreamHandler(protocol.ID(proto), handleStream)
	_, _ = newPubSub(ctx, host)
	go discoverPeers(host, ctx)
	select {}
}
