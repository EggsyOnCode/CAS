package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCP(t *testing.T) {
	listenAddr := ":4040"
	tr := NewTCPTransporter(listenAddr)

	assert.Equal(t, listenAddr, tr.ListenAddr)
}
