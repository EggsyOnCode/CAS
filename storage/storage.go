package storage

import (
	"fmt"
	"io"
	"os"
)

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

	//we'll make a directory in the local fs with the pahtname
	// in that path we'll create a new file
	// we'll copy the contents of the io.Reader into that buffer/file

	pathName := key
	err := os.MkdirAll(pathName, os.ModePerm)
	if err != nil {
		return err
	}

	fileName := "abc.txt"
	pathAndFileName := pathName + "/"+ fileName
	//creating new file
	file, err := os.Create(pathAndFileName)
	if err != nil{
		return nil
	}

	n , err := io.Copy(file, r)
	if err != nil{
		return err
	}

	fmt.Printf("%d bytes written to disk at location %s", n, pathAndFileName)

	return nil
}
