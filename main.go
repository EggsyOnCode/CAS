package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
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
	s3 := makeServer(":5000", ":4000", ":3000")
	go func() {
		log.Fatal(s1.Start())
	}()
	time.Sleep(1 * time.Second)
	go func() {
		log.Fatal(s2.Start())
	}()
	//we need to make this a goroutine so that we can do other stuff whiel the server is starting otherwise the thread would be blocked here
	go s3.Start()
	time.Sleep(2 * time.Second)
	// List of file paths
	filePaths := []string{"./assets/resume.pdf"}

	go func() {
		for i, path := range filePaths {
			// Open the file
			file, err := os.Open(path)
			if err != nil {
				log.Fatalf("Failed to open file %s: %s", path, err)
			}
			defer file.Close()

			// Get file size
			fileInfo, err := file.Stat()
			if err != nil {
				log.Fatalf("Failed to get file info for %s: %s", path, err)
			}

			// Read file content into a buffer
			fileBuffer := make([]byte, fileInfo.Size())
			if _, err := file.Read(fileBuffer); err != nil {
				log.Fatalf("Failed to read file %s: %s", path, err)
			}

			// Create a reader for the file content
			fileReader := bytes.NewReader(fileBuffer)

			// Store file data
			key := fmt.Sprintf("file_%d", i+1)
			if err := s3.StoreData(key, fileReader); err != nil {
				log.Fatalf("Failed to store file %s data: %s", path, err)
			}
			fmt.Printf("Stored file %s data with key: %s\n", path, key)

			// Simulate fetching the file from the network
			time.Sleep(2 * time.Second)

			// Delete the local copy of the file
			if err := s3.store.Delete(key); err != nil {
				log.Fatalf("Failed to delete file %s: %s", key, err)
			}
			fmt.Printf("Deleted local copy of file %s\n", key)

			// Fetch the file from the network
			fetchedFile, err := s3.Get(key)
			if err != nil {
				log.Fatalf("Failed to fetch file %s: %s", key, err)
			}

			// Read fetched file content
			fetchedFileContent, err := ioutil.ReadAll(fetchedFile)
			if err != nil {
				log.Fatalf("Failed to read fetched file %s content: err is %s ; the content was %s", key, err, fetchedFileContent)
			}

			// Print fetched file content
			// fmt.Printf("Fetched file %s content: %s\n", key, fetchedFileContent)
		}
	}()

	select {}
}
