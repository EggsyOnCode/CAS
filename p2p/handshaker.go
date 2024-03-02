package p2p

type Handshake func(Peer) error

func NOPHandshake(Peer) error {
	return nil
}
