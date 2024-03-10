package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/EggsyOnCode/CAS/encrypt"
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
		EncKey:            encrypt.NewEncryptionKey(),
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

	// testing by storing a pic in the network
	// pic, err := os.Open("./assets/testpic.jpg")
	// if err != nil {
	// 	log.Fatal(err)

	// }
	// defer pic.Close()
	// size, err := pic.Stat()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// picBuf := make([]byte, size.Size())
	// if _, err := bufio.NewReader(pic).Read(picBuf); err != nil {
	// 	log.Fatal(err)
	// }

	// for i := 1; i < 2; i++ {
	var testData []byte = []byte("this is new jpg file")
	data := bytes.NewReader(testData)
	s2.StoreData("pic_1", data)
	time.Sleep(20 * time.Millisecond)
	// }

	if err := s2.store.Delete("pic_1"); err != nil {
		log.Fatal(err)
	}

	// for i := 1; i < 2; i++ {

	n, err := s2.Get("pic_1")
	if err != nil {
		log.Fatal(err)
	}

	b, err := ioutil.ReadAll(n)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("{%s} is the data of ur file\n", string(b))
	// }
	select {}
}
