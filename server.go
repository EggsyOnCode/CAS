package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"

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
		case msg := <-f.Transporter.Consume():
			fmt.Println(msg)
		case <-f.quitch: //when channel quits
			return
		}
	}
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

type Payload struct {
	Key  string
	Data []byte
}

func (f *FileServer) broadcast(p Payload) error {
	//treat the peers as io.Writers and stream file to them
	peers := []io.Writer{}
	for _, peer := range f.peers {
		peers = append(peers, peer)
	}

	//multiwriters
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(p)
}

func (f *FileServer) StoreData(key string, r io.Reader) error {
	//store the file to disk

	if err := f.store.Write(key, r); err != nil {
		return err
	}

	//broadcast the file to all known peers in hte network
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, r)
	if err != nil {
		return err
	}

	fmt.Print(buf.Bytes())
	return nil
}

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
