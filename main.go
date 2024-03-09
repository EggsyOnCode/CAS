package main

import (
	"fmt"
	"io/ioutil"
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
	time.Sleep(1 * time.Second)
	//we need to make this a goroutine so that we can do other stuff whiel the server is starting otherwise the thread would be blocked here
	go s2.Start()
	time.Sleep(2 * time.Second)

	// for i := 0; i < 3; i++ {
	// 	data := bytes.NewReader([]byte(fmt.Sprintf("this is new jpg file %d", i)))
	// 	s2.StoreData(fmt.Sprintf("myprivateData_%d", i), data)
	// 	time.Sleep(5 * time.Millisecond)
	// }

	n, err := s2.Get("myprivateData_2")
	if err != nil {
		log.Fatal(err)
	}

	b, err := ioutil.ReadAll(n)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("{%s} is the data of ur file\n", string(b))
	select {}
}
