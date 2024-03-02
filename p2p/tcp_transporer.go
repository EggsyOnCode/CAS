package p2p

import (
	"net"
	"sync"
)

type TCPTransporter struct {
	ListenAddr string
	Listener   net.Listener

	mu sync.RWMutex
	peers map[net.Addr]Peer
}

//here we could've returned Transpoort interface but that would've been bad for teesting because we wouldn't have had access to Listener etc objects
func NewTCPTransporter(listenAddr string) *TCPTransporter{
	return &TCPTransporter{
		ListenAddr: listenAddr,
	}
}