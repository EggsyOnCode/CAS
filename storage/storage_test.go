package storage

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCASPathTransformFunc(t *testing.T) {
	key := "key"
	expectedPaths := "a62f2/225bf/70bfa/ccbc7/f1ef2/a3978/36717/377de"

	assert.Equal(t, CASPathTransformFunc(key), expectedPaths)
}

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
