package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/EggsyOnCode/CAS/encrypt"
	"github.com/EggsyOnCode/CAS/p2p"
	"github.com/EggsyOnCode/CAS/storage"
)

// file server is the central node that will coordinate the delegation of jobs
// will allow one peer to fetch and query files from the network, store htem on hte network etc...

type FileServerOpts struct {
	EncKey            []byte
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

// defining the types of messages
type MessageStoreFile struct {
	Key  string
	Size int64
}

type MessageGetFile struct {
	Key string
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

func (f *FileServer) stream(msg *Message) error {
	//treat the peers as io.Writers and stream file to them
	peers := []io.Writer{}
	for _, peer := range f.peers {
		peers = append(peers, peer)
	}

	//multiwriters
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(msg)
}

func (f *FileServer) broadcast(msg *Message) error {
	msgBuf := new(bytes.Buffer)
	//encode the msg
	if err := gob.NewEncoder(msgBuf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range f.peers {
		//send the msg typ to the peer
		peer.Send([]byte{p2p.IncomingMessage})
		if err := peer.Send(msgBuf.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

// it will look for the file in the local store and if not found it will fetch it from the network
// and store it in its local store with the exact same pathName
func (f *FileServer) Get(key string) (io.Reader, error) {
	if f.store.Has(key) {
		return f.store.Read(key)
	}

	fmt.Printf("file not found in local store, fetching from network\n")

	msg := Message{
		Payload: MessageGetFile{
			Key: key,
		},
	}

	if err := f.broadcast(&msg); err != nil {
		return nil, err
	}
	time.Sleep(time.Microsecond * 10)
	for _, peer := range f.peers {
		// reading file size first to limit hte reading bytes sot htat hte reader doesnt' hang
		var fileSize int64
		binary.Read(peer, binary.LittleEndian, &fileSize)
		n, err := f.store.Write(key, io.LimitReader(peer, fileSize))
		if err != nil {
			return nil, err
		}
		fmt.Printf("[%s] received %d bytes from %s\n", f.Transporter.Addr(), n, peer.RemoteAddr().String())

		peer.CloseStream()
	}

	return f.store.Read(key)
}

func (f *FileServer) StoreData(key string, r io.Reader) error {
	//store the file to disk
	//broadcast the file to all known peers in hte network

	fileBuffer := new(bytes.Buffer)
	tee := io.TeeReader(r, fileBuffer)
	//returned is file size
	size, err := f.store.Write(key, tee)
	if err != nil {
		return err
	}
	msg := Message{
		Payload: MessageStoreFile{
			Key: key,
			//added 16 cuz we are prepending a nonce to the encrypted file of the data
			Size: size + 16,
		},
	}

	if err := f.broadcast(&msg); err != nil {
		return err
	}
	time.Sleep(time.Microsecond * 10)

	//add multiwriters here to write buf filebuff into the peers

	// data from buf i.e hte file data is being copied over to each peer (in the network)
	// because peer is also an io.Writer so we can use io.Copy to copy the data to it
	for _, peer := range f.peers {
		peer.Send([]byte{p2p.IncomingStream})
		n, err := encrypt.CopyEncrypt(f.EncKey, fileBuffer, peer)
		if err != nil {
			return err
		}

		fmt.Printf("[%s] sent %d bytes to %s\n", f.Transporter.Addr(), n, peer.RemoteAddr().String())
	}
	return nil
}

// we are having loop for the server deaemon to recieve msgs in its channels and process them concurrently
// the for {select {}} is used to execure teh select {} indefinitely
// select {} is used to handle multiple channle operations concurrently in a non-blocking fashion
func (f *FileServer) loop() {
	defer func() {
		log.Println("the server stopped due to error or user quitting action")
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
				log.Printf("error decoding msg %s", err)

			}
			if err := f.handleMsg(rpc.From.String(), &msg); err != nil {
				log.Printf("error handling msg %s", err)
			}

		case <-f.quitch: //when channel quits
			return
		}
	}
}

func (f *FileServer) handleMsg(from string, msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		return f.handleMsgStoreData(from, v)
	case MessageGetFile:
		return f.handleMsgGetData(from, v)
	}

	return nil
}

func (f *FileServer) handleMsgGetData(from string, msg MessageGetFile) error {
	//check if the file is in the local store
	if !f.store.Has(msg.Key) {
		return fmt.Errorf("need to serve file (%s) but file not found", msg.Key)
	}

	//fetch the file from the network
	//send the file to the requesting peer
	file, err := f.store.Read(msg.Key)
	if err != nil {
		return err
	}
	//closing the io.Reader that is retrieved above
	temp, ok := file.(io.ReadCloser)
	if ok {
		fmt.Println("closing the file")
		defer temp.Close()
	}

	peer, ok := f.peers[from]
	if !ok {
		return fmt.Errorf("peer not found %s", from)
	}

	//send an incoming message flag and then send the actual file size as an int64
	peer.Send([]byte{p2p.IncomingStream})
	fileSize, err := f.store.GetFileSize(msg.Key)
	log.Printf("file size is %d\n", fileSize)
	if err != nil {
		return err
	}
	binary.Write(peer, binary.LittleEndian, fileSize)

	n, err := io.Copy(peer, file)
	if err != nil {
		return err
	}

	log.Printf("%d bytes sent to %s\n", n, from)
	return nil

}
func (f *FileServer) handleMsgStoreData(from string, msg MessageStoreFile) error {

	peer, ok := f.peers[from]
	if !ok {
		log.Fatalf("peer not found %s", from)
	}
	fmt.Printf("rcv msg is %v\n", msg.Key)
	n, err := f.store.Write(msg.Key, io.LimitReader(peer, msg.Size))
	if err != nil {
		return err
	}
	log.Printf("[%s] wrote %d bytes to disk", f.Transporter.Addr(), n)
	peer.CloseStream()
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

func init() {
	gob.Register(MessageGetFile{})
	gob.Register(MessageStoreFile{})
}
