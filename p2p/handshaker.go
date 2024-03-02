package p2p

type Handshake func(any) error

func NOPHandshake(any) error {
	return nil
}
