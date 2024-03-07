package main

import (
	"fmt"
	"log"

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
	}

}
func (f *FileServer) Stop() {
	close(f.quitch)
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
