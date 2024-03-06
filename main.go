package main

import (
	"fmt"
	"log"

	"github.com/EggsyOnCode/CAS/p2p"
	"github.com/EggsyOnCode/CAS/storage"
)

func OnPeer(p p2p.Peer) error { fmt.Printf("entertain the peer \n"); return nil }

func main() {
	tcpTransporterOpts := p2p.TCPTransportOpts{
		ListenAddr:    ":3000",
		Handshakefunc: p2p.NOPHandshake,
		Decoder:       p2p.DefaultDecoder{},
		OnPeer:        OnPeer,
	}

	tcpTransporter := p2p.NewTCPTransporter(tcpTransporterOpts)

	fileServerOpts := FileServerOpts{
		StorageRoot:       "xen_Net",
		PathTransformFunc: storage.CASPathTransformFunc,
		Transporter:       tcpTransporter,
	}

	fileServer := NewFileServer(fileServerOpts)

	if err := fileServer.Start(); err != nil {
		log.Fatal(err)
	}

	select {}
}
