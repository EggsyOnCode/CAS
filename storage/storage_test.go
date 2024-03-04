package storage

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCASPathTransformFunc(t *testing.T) {
	key := "key"
	expectedPaths := "a62f2/225bf/70bfa/ccbc7/f1ef2/a3978/36717/377de"

	assert.Equal(t, CASPathTransformFunc(key).PathName, expectedPaths)

}

func TestStore(t *testing.T) {
	key := "myTestFile.txt"
	opts := &StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	store := NewStore(opts)

	//reader
	r := bytes.NewReader([]byte("this is new jpg file"))
	r_data := []byte("this is new jpg file")

	assert.Nil(t, store.writeStream(key, r))

	//testing reading func
	buf, _ := store.readStream(key)

	data, _ := io.ReadAll(buf)
	fmt.Println(string(data))
	assert.Equal(t, data, r_data)
}
