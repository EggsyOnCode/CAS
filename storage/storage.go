package storage

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log"
	"os"
	"strings"
)

func CASPathTransformFunc(key string) string {
	//reutnrs a sha1 hash of hte key
	// eg key
	hash := sha1.Sum([]byte(key))

	//hashStr rep the hex encoded hash of the key like 0xsdfsdf
	hashStr := hex.EncodeToString(hash[:])
	// eg: a1357f312ce120ba9b5c2fbc1be02e2a7b64e4db
	log.Printf(" hash str is %s", hashStr)

	blocksize := 5
	sliceLen := len(hashStr) / blocksize
	paths := make([]string, sliceLen)

	for i := 0; i < sliceLen; i++ {
		from, to := i*blocksize, (i*blocksize)+blocksize
		paths[i] = hashStr[from:to]
	}

	return strings.Join(paths, "/")
	// eg a1357/f312c/e120b/a9b5c/2fbc1/be02e/2a7b6/4e4db 
}

type PathTransformFunc func(string) string

var DefaultPathName = func(key string) string {
	return key
}

type StoreOpts struct {
	// path transform will hash the content of the file given to it then some transformation of it
	PathTransformFunc PathTransformFunc
}

// Store will represent a machine that will be responsible for storing files given to it on a Distributed CAS
type Store struct {
	StoreOpts
}

func NewStore(opts *StoreOpts) *Store {
	return &Store{StoreOpts: *opts}
}

func (s *Store) writeStream(key string, r io.Reader) error {
	// the pathname would be some modified hash of the content of hte file
	// io.REader thats passed will be holding the byte slice of the file contents

	//we'll make a directory in the local fs with the pathname
	// in that path we'll create a new file
	// we'll copy the contents of the io.Reader into that buffer/file

	pathName := key
	err := os.MkdirAll(pathName, os.ModePerm)
	if err != nil {
		return err
	}

	fileName := "abc.txt"
	pathAndFileName := pathName + "/" + fileName
	//creating new file
	file, err := os.Create(pathAndFileName)
	if err != nil {
		return nil
	}

	n, err := io.Copy(file, r)
	if err != nil {
		return err
	}

	log.Printf("%d bytes written to disk at location %s", n, pathAndFileName)

	return nil
}
