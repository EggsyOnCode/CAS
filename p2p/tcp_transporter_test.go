package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCP(t *testing.T) {
	opts := TCPTransportOpts{
		ListenAddr: ":3000",
		Handshakefunc: NOPHandshake,
		Decoder: DefaultDecoder{},
	}
	tr := NewTCPTransporter(opts)

	assert.Equal(t, opts.ListenAddr, tr.ListenAddr)

	assert.Nil(t, tr.ListenAndAccept())
}
