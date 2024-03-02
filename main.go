package main

import (
	"fmt"
	"log"

	"github.com/EggsyOnCode/CAS/p2p"
)

func OnPeer(p p2p.Peer) error { fmt.Printf("entertain the peer \n"); return nil }
func main() {
	tcpOpts := p2p.TCPTransportOpts{
		ListenAddr:    ":3000",
		Handshakefunc: p2p.NOPHandshake,
		Decoder:       p2p.DefaultDecoder{},
		OnPeer:        OnPeer,
	}
	tr := p2p.NewTCPTransporter(tcpOpts)
	go func() {
		for {
			msg := <-tr.Consume()
			fmt.Printf("%+v\n", msg)
		}
	}()
	// this syntax means that the err recevied from the listen func ; test the con on it
	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Server is listening!")

	//we are blocking the thread here ; why?
	// we are blocking to keep the server alive  and keep listening
	select {}
}
