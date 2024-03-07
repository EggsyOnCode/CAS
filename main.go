package main

import (
	"bytes"
	"log"
	"time"

	"github.com/EggsyOnCode/CAS/p2p"
	"github.com/EggsyOnCode/CAS/storage"
)

func makeServer(listenAddr string, nodes ...string) *FileServer {
	tcpTransporterOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		Handshakefunc: p2p.NOPHandshake,
		Decoder:       p2p.DefaultDecoder{},
	}

	tcpTransporter := p2p.NewTCPTransporter(tcpTransporterOpts)

	fileServerOpts := FileServerOpts{
		StorageRoot:       listenAddr + "network",
		PathTransformFunc: storage.CASPathTransformFunc,
		Transporter:       tcpTransporter,
		BootstrapNodes:    nodes,
	}

	s := NewFileServer(fileServerOpts)

	tcpTransporter.OnPeer = s.OnPeer

	return s
}
func main() {
	s1 := makeServer(":3000")
	s2 := makeServer(":4000", ":3000")

	// could be called in a gorotuine
	go func() {
		log.Fatal(s1.Start())
	}()

	//we need to make this a goroutine so that we can do other stuff whiel the server is starting otherwise the thread would be blocked here
	go s2.Start()
	time.Sleep(1 * time.Second)

	data := bytes.NewReader([]byte("this is my data"))

	s2.StoreData("myprivatefiles", data)

}
