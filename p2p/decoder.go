package p2p

import (
	"encoding/gob"
	"io"
)

type Decoder interface {
	Decode(io.Reader, any) error
}

type GOBDecoder struct{}

//gobs are binary streams being exchanged between the emitter (encoder) and the recevier(Decoder)
func (dec GOBDecoder) Decode(r io.Reader, v any) error {
	return gob.NewDecoder(r).Decode(v)
}
