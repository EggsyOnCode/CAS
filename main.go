package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
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

	pic, err := os.Open("./assets/testpic.jpg")
	if err != nil {
		log.Fatal(err)

	}
	defer pic.Close()
	size, err := pic.Stat()
	if err != nil {
		log.Fatal(err)
	}
	picBuf := make([]byte, size.Size())
	if _, err := bufio.NewReader(pic).Read(picBuf); err != nil {
		log.Fatal(err)
	}

	for i := 1; i < 2; i++ {
		data := bytes.NewReader(picBuf)
		s2.StoreData(fmt.Sprintf("pic_%d", i), data)
		time.Sleep(5 * time.Millisecond)
	}

	// for i := 1; i < 4; i++ {

	// 	n, err := s2.Get(fmt.Sprintf("pic_%d", i))
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	b, err := ioutil.ReadAll(n)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	fmt.Printf("{%s} is the data of ur file\n", string(b))
	// }
	select {}
}
