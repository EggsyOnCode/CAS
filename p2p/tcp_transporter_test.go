package p2p

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCP(t *testing.T) {
	opts := TCPTransportOpts{
		ListenAddr:    ":3000",
		Handshakefunc: NOPHandshake,
		Decoder:       DefaultDecoder{},
	}
	tr := NewTCPTransporter(opts)

	assert.Equal(t, opts.ListenAddr, tr.ListenAddr)

	assert.Nil(t, tr.ListenAndAccept())
}
func TestTCPTransporterClose(t *testing.T) {
	// Create a TCP listener
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	// Verify that the listener is accepting connections
	// You can write additional logic here to perform connection tests

	// Close the listener
	err = listener.Close()
	if err != nil {
		t.Fatalf("Failed to close listener: %v", err)
	}

	// Attempt to connect to the closed listener
	_, err = net.Dial("tcp", ":8080")
	if err == nil {
		t.Fatalf("Expected connection to fail, but it succeeded")
	}
}
