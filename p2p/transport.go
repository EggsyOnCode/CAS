package p2p

import "net"

// Peer rep the remote node client
type Peer interface {
	// every peer must implement this so that we can close the conn with them
	Close() error
	RemoteAddr() net.Addr
}

// Transporter ; the interface that the transporter implement i.e tcp,udp,websockets etc have to conform to
type Transporter interface {
	ListenAndAccept() error
	// consume will be fed with a channel of RPC messages
	Consume() <-chan RPC
	Close() error
	Dial(string) error
}
