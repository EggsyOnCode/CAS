package storage

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStore(t *testing.T) {
	key := "myTestFile.txt"
	opts := &StoreOpts{
		PathTransformFunc: DefaultPathName,
	}
	store := NewStore(opts)

	//reader
	r := bytes.NewReader([]byte("this is new jpg file"))

	assert.Nil(t, store.writeStream(key, r))
}
