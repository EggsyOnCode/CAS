package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/EggsyOnCode/CAS/p2p"
	"github.com/EggsyOnCode/CAS/storage"
)

// file server is the central node that will coordinate the delegation of jobs
// will allow one peer to fetch and query files from the network, store htem on hte network etc...

type FileServerOpts struct {
	StorageRoot       string
	PathTransformFunc storage.PathTransformFunc
	Transporter       p2p.Transporter
	BootstrapNodes    []string
}

type FileServer struct {
	FileServerOpts
	store *storage.Store
	// why not msgch for the file server daemon?
	quitch chan struct{}

	peerLock sync.Mutex
	peers    map[string]p2p.Peer
}

type Message struct {
	Payload any
}

// type DataMessage struct {
// 	Key  string
// 	Data []byte
// }

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := storage.StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}

	return &FileServer{
		store:          storage.NewStore(&storeOpts),
		FileServerOpts: opts,
		quitch:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}

}
func (f *FileServer) Stop() {
	close(f.quitch)
}

func (f *FileServer) OnPeer(p p2p.Peer) error {
	f.peerLock.Lock()
	defer f.peerLock.Unlock()

	f.peers[p.RemoteAddr().String()] = p

	log.Printf("connected with remote peer %s", p.RemoteAddr().String())

	return nil
}

func (f *FileServer) bootStrapNodes() error {
	for _, address := range f.BootstrapNodes {
		if len(address) == 0 {
			continue
		}
		go func(address string) {
			fmt.Println("attempting to connect to bootstrap node")
			if err := f.Transporter.Dial(address); err != nil {
				log.Fatalf("error attempting to connect to bootstrap nodes %s", err)
			}
		}(address)
	}

	return nil
}

func (f *FileServer) broadcast(msg *Message) error {
	//treat the peers as io.Writers and stream file to them
	peers := []io.Writer{}
	for _, peer := range f.peers {
		peers = append(peers, peer)
	}

	//multiwriters
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(msg)
}

func (f *FileServer) StoreData(key string, r io.Reader) error {
	//store the file to disk
	//broadcast the file to all known peers in hte network

	buf := new(bytes.Buffer)
	msg := Message{
		Payload: []byte("hello"),
	}
	//encode the msg
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range f.peers {
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}

	time.Sleep(time.Second * 3)
	payload := ([]byte("LARGE FILE"))

	for _, peer := range f.peers {
		if err := peer.Send(payload); err != nil {
			return err
		}
	}

	return nil
	// tee := io.TeeReader(r, buf)

	// if err := f.store.Write(key, tee); err != nil {
	// 	return err
	// }

	// p := &DataMessage{
	// 	Key:  key,
	// 	Data: buf.Bytes(),
	// }
	// fmt.Print(buf.Bytes())
	// return f.broadcast(&Message{
	// 	From:    "todo",
	// 	Payload: p,
	// })
}

// we are having loop for the server deaemon to recieve msgs in its channels and process them concurrently
// the for {select {}} is used to execure teh select {} indefinitely
// select {} is used to handle multiple channle operations concurrently in a non-blocking fashion
func (f *FileServer) loop() {
	defer func() {
		log.Println("the server stopped due to user quitting action")
		if err := f.Transporter.Close(); err != nil {
			log.Println(err)
		}
	}()
	for {
		select {
		case rpc := <-f.Transporter.Consume():
			var msg Message
			// decoding the msg
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Fatal(err)
			}

			fmt.Printf("rcvd msg is %+s\n", msg.Payload.([]byte))

			peer, ok := f.peers[rpc.From.String()]
			if !ok {
				log.Fatalf("peer not found %s", rpc.From.String())
			}

			b := make([]byte, 1024)
			if _, err := peer.Read(b); err != nil {
				panic(err)
			}
			fmt.Printf("the data file received is %v\n", string(b))
			peer.Wg().Done()

		case <-f.quitch: //when channel quits
			return
		}
	}
}

// func (f *FileServer) handleMsg(msg *Message) error {
// 	switch v := msg.Payload.(type) {
// 	case *DataMessage:
// 		fmt.Printf("rcvv data is %+v\n", v.Data)
// 	}

//		return nil
//	}

func (f *FileServer) Start() error {
	if err := f.Transporter.ListenAndAccept(); err != nil {
		return err
	}

	// when the file server starts connect to bootstrap nodes
	f.bootStrapNodes()
	// start the loop i.e the server daemon
	f.loop()
	return nil
}
