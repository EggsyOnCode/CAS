package p2p

// Peer rep the remote node client
type Peer interface {
}

// Transporter ; the interface that the transporter implement i.e tcp,udp,websockets etc have to conform to
type Transporter interface {
	ListenAndAccept() error
}
