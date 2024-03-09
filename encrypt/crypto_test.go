package encrypt

import (
	"bytes"
	"fmt"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestCopyEncryptnDecrypt(t *testing.T) {
	payload := "this is new jpg file"
	key := newEncryptionKey()
	src := bytes.NewReader([]byte(payload))
	dst := new(bytes.Buffer)
	n, err := CopyEncrypt(key, src, dst)
	assert.Nil(t, err)
	fmt.Println(dst.String())

	//testing decrypt
	out := new(bytes.Buffer)
	randN, err:= CopyDecrypt(key, dst, out)
	assert.Nil(t, err)
	assert.Equal(t, n, randN)
	fmt.Println(out.String())
	assert.Equal(t, out.String(), payload)
}
