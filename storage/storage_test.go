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

func TestStoreDelete(t *testing.T) {
	key := "myTestFile.txt"
	opts := &StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	store := NewStore(opts)

	//reader
	// test_strig := "this is new jpg file testing for deletion"
	// r := bytes.NewReader([]byte(test_strig))
	// assert.Nil(t, store.writeStream(key, r))

	assert.Nil(t, store.Delete(key))
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
func TestStoreRead(t *testing.T){
	key := "myTestFile.txt"
	opts := &StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	store := NewStore(opts)
	//reader
	//testing reading func
	buf, _ := store.readStream(key)
	r_data := []byte("this is new jpg file")
	data, _ := io.ReadAll(buf)
	fmt.Println(string(data))
	assert.Equal(t, data, r_data)
}