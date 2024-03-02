package p2p

import (
	"encoding/gob"
	"io"
)

type Decoder interface {
	Decode(io.Reader, *RPC) error
}

type GOBDecoder struct{}

//gobs are binary streams being exchanged between the emitter (encoder) and the recevier(Decoder)
func (dec GOBDecoder) Decode(r io.Reader, msg *RPC) error {
	return gob.NewDecoder(r).Decode(msg)
}


type DefaultDecoder struct{}

func (dec DefaultDecoder) Decode(r io.Reader, msg *RPC) error {
	//1kb buffer for the msg from the peer
	buf :=make([]byte, 1024)
	n, err := r.Read(buf)
	if err != nil{
		return err
	}
	// n is the no of bytes that reader could read from and put into buffer; so we
	// start readifn from the start of the buffer until n
	msg.Payload = buf[:n]
	return nil
}