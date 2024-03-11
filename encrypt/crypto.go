package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	"github.com/h2non/filetype"
)

func GetFileType(data []byte) string {
	kind, _ := filetype.Get(data)
	if kind == filetype.Unknown {
		fmt.Println("Unknown file type")
	} else {
		fmt.Printf("File type matched: %s\n", kind.Extension)
		return kind.Extension
	}
	return ""
}
func NewEncryptionKey() []byte {
	keyBuf := make([]byte, 32)
	io.ReadFull(rand.Reader, keyBuf)
	return keyBuf
}
func CopyDecrypt(key []byte, src io.Reader, dst io.Writer) (string, int, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", 0, err
	}

	// Read the IV from the given io.Reader which, in our case should be the
	// the block.BlockSize() bytes we read.
	//iv here is a nonce
	iv := make([]byte, block.BlockSize())
	if _, err := src.Read(iv); err != nil {
		return "", 0, err
	}

	var (
		buf    = make([]byte, 32*1024)
		stream = cipher.NewCTR(block, iv)
		nw     = block.BlockSize()
		ext    string
		i      = 0
	)

	for {
		n, err := src.Read(buf)
		if n > 0 {
			stream.XORKeyStream(buf, buf[:n])
			nn, err := dst.Write(buf[:n])
			if err != nil {
				return "", 0, err
			}
			nw += nn
		}
		if i == 0 {
			ext = GetFileType(buf)
			fmt.Printf("File type: %s\n", ext)
			i++
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", 0, err
		}
		i++
	}

	// Determine the file type using the data buffer

	return ext, nw, err
}

func CopyEncrypt(key []byte, src io.Reader, dst io.Writer) (int, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}

	iv := make([]byte, block.BlockSize()) // 16 bytes
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return 0, err
	}

	// prepend the IV to the file.
	if _, err := dst.Write(iv); err != nil {
		return 0, err
	}

	var (
		buf    = make([]byte, 32*1024)
		stream = cipher.NewCTR(block, iv)
		nw     = block.BlockSize()
	)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			stream.XORKeyStream(buf, buf[:n])
			nn, err := dst.Write(buf[:n])
			if err != nil {
				return 0, err
			}
			nw += nn
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
	}
	return nw, nil
}
