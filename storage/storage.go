package storage

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/EggsyOnCode/CAS/encrypt"
)

var defaultRoot = "xenNet"

func CASPathTransformFunc(key string) Pathkey {
	//reutnrs a sha1 hash of hte key
	// eg key
	hash := sha1.Sum([]byte(key))

	//hashStr rep the hex encoded hash of the key like 0xsdfsdf
	hashStr := hex.EncodeToString(hash[:])
	// eg: a1357f312ce120ba9b5c2fbc1be02e2a7b64e4db
	blocksize := 5
	sliceLen := len(hashStr) / blocksize
	paths := make([]string, sliceLen)

	for i := 0; i < sliceLen; i++ {
		from, to := i*blocksize, (i*blocksize)+blocksize
		paths[i] = hashStr[from:to]
	}
	return Pathkey{
		PathName: strings.Join(paths, "/"),
		// eg a1357/f312c/e120b/a9b5c/2fbc1/be02e/2a7b6/4e4db
		Original: hashStr,
	}
}

type PathTransformFunc func(string) Pathkey

var DefaultPathTransformFunc = func(key string) Pathkey {
	return Pathkey{
		PathName: key,
		Original: key,
	}
}

type StoreOpts struct {
	// path transform will hash the content of the file given to it then some transformation of it
	PathTransformFunc PathTransformFunc
	//root folder name
	Root string
}

// Store will represent a machine that will be responsible for storing files given to it on a Distributed CAS
type Store struct {
	StoreOpts
}

type Pathkey struct {
	PathName string
	Original string
}

func (p *Pathkey) Fullpath() string {
	// we can derive a file's path using its filename
	// this is imp so that we can retrive it
	return fmt.Sprintf("%s/%s", p.PathName, p.Original)
}
func (s *Store) Has(key string) bool {
	pathKey := s.PathTransformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, pathKey.Fullpath())

	_, err := os.Stat(fullPathWithRoot)
	return !errors.Is(err, os.ErrNotExist)
}

func NewStore(opts *StoreOpts) *Store {
	if opts.PathTransformFunc == nil {
		opts.PathTransformFunc = DefaultPathTransformFunc
	}
	if len(opts.Root) == 0 {
		opts.Root = defaultRoot
	}
	return &Store{StoreOpts: *opts}
}

// Deleting a file using its key
func (s *Store) Delete(key string) error {
	pathkey := s.PathTransformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s", s.StoreOpts.Root, pathkey.Fullpath())
	err := os.Remove(fullPathWithRoot)
	if err != nil {
		log.Printf("failed to delete file at path %s: %s", pathkey.Fullpath(), err)
		return err
	}
	log.Printf("file at path %s has been removed", pathkey.Fullpath())

	// Traverse up the directory hierarchy and delete empty directories
	currentDir := s.StoreOpts.Root + "/" + pathkey.PathName
	for currentDir != "" {
		files, err := ioutil.ReadDir(currentDir)
		if err != nil {
			log.Printf("error reading directory %s: %s", currentDir, err)
			return err
		}
		if len(files) > 0 {
			break // Directory is not empty, stop traversing
		}
		err = os.RemoveAll(currentDir)
		if err != nil {
			log.Printf("failed to delete directory %s: %s", currentDir, err)
			return err
		}
		// log.Printf("directory %s has been removed", currentDir)

		// Move up to the parent directory
		currentDir = filepath.Dir(currentDir)
	}
	return nil
}

// returns the fileSize of a file given its key
func (s *Store) GetFileSize(key string) (int64, error) {
	pathkey := s.PathTransformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s", s.StoreOpts.Root, pathkey.Fullpath())
	fil, err := os.Stat(fullPathWithRoot)
	if err != nil {
		return 0, err
	}
	return fil.Size(), nil
}

// Reading file contents using its key
// TODO: instead of copying the returned reader to a buffer, we can return the reader directly
func (s *Store) Read(key string, ext string) (io.Reader, error) {
	f, err := s.readStream(key, ext)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, f)
	return buf, err
}

func (s *Store) readStream(key string, ext string) (io.ReadCloser, error) {
	pathkey := s.PathTransformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s", s.StoreOpts.Root, pathkey.Fullpath())
	if ext != "" {
		fullPathWithRoot = fullPathWithRoot + "." + ext
	}
	return os.Open(fullPathWithRoot)
}

func (s *Store) Write(key string, r io.Reader) (int64, error) {
	return s.writeStream(key, r)
}

func (s *Store) WriteDecrypt(encKey []byte, key string, r io.Reader) (string, int64, error) {
	//creating new file
	file, err := s.openFileForWriting(key)
	if err != nil {
		return "", 0, nil
	}

	ext, n, err := encrypt.CopyDecrypt(encKey, r, file)
	if err != nil {
		return "", 0, nil
	}
	os.Rename(file.Name(), file.Name()+"."+ext)

	return ext, int64(n), nil
}
func (s *Store) openFileForWriting(key string) (*os.File, error) {
	pathkey := s.PathTransformFunc(key)
	fullFolderPathWithRoot := s.StoreOpts.Root + "/" + pathkey.PathName
	err := os.MkdirAll(fullFolderPathWithRoot, os.ModePerm)
	if err != nil {
		return nil, err
	}

	fullPathWithRoot := fmt.Sprintf("%s/%s", s.StoreOpts.Root, pathkey.Fullpath())
	//creating new file
	return os.Create(fullPathWithRoot)
}

// returns file size being streamed
func (s *Store) writeStream(key string, r io.Reader) (int64, error) {
	// the pathname would be some modified hash of the content of hte file
	// io.REader thats passed will be holding the byte slice of the file contents

	//we'll make a directory in the local fs with the pathname
	// in that path we'll create a new file
	// we'll copy the contents of the io.Reader into that buffer/file

	//creating new file
	file, err := s.openFileForWriting(key)
	if err != nil {
		return 0, nil
	}

	return io.Copy(file, r)
}
