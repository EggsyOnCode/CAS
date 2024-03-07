package main

import (
	"fmt"
	"log"

	"github.com/EggsyOnCode/CAS/p2p"
	"github.com/EggsyOnCode/CAS/storage"
)

func OnPeer(p p2p.Peer) error { fmt.Printf("entertain the peer \n"); return nil }

func makeServer(listenAddr string, nodes ...string) *FileServer {
	tcpTransporterOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		Handshakefunc: p2p.NOPHandshake,
		Decoder:       p2p.DefaultDecoder{},
		OnPeer:        OnPeer,
	}

	tcpTransporter := p2p.NewTCPTransporter(tcpTransporterOpts)

	fileServerOpts := FileServerOpts{
		StorageRoot:       listenAddr + "network",
		PathTransformFunc: storage.CASPathTransformFunc,
		Transporter:       tcpTransporter,
		BootstrapNodes:    nodes,
	}

	return NewFileServer(fileServerOpts)

}
func main() {
	s1 := makeServer(":3000")
	s2 := makeServer(":4000", ":3000")

	// could be called in a gorotuine
	go func() {
		log.Fatal(s1.Start())
	}()

	s2.Start()

}
