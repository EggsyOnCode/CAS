package p2p

import (
	"encoding/gob"
	"io"
)

type Decoder interface {
	Decode(io.Reader, *RPC) error
}

type GOBDecoder struct{}

// gobs are binary streams being exchanged between the emitter (encoder) and the recevier(Decoder)
func (dec GOBDecoder) Decode(r io.Reader, msg *RPC) error {
	return gob.NewDecoder(r).Decode(msg)
}

type DefaultDecoder struct{}

func (dec DefaultDecoder) Decode(r io.Reader, msg *RPC) error {
	//peekBuffer to peak at teh first byte of the rpc msg which would
	// indicate the purpose of the msg
	// if a stream then we'll NOT Decode and block the thread in the writeStream until the streaming is done
	peekBuf := make([]byte, 1)
	if _, err := r.Read(peekBuf); err != nil {
		return err
	}

	// if the first byte is 0x2 then it's a stream
	if peekBuf[0] == IncomingStream {
		msg.Stream = true
		return nil // cuz we don't want to decode the stream in this case
	}

	//1kb buffer for the msg from the peer
	buf := make([]byte, 1024)
	n, err := r.Read(buf)
	if err != nil {
		return err
	}
	// n is the no of bytes that reader could read from and put into buffer; so we
	// start readifn from the start of the buffer until n
	msg.Payload = buf[:n]
	return nil
}
