package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"errors"
	"io"
)

func CreateAesGcmAead(key []byte) (cipher.AEAD, error) {
	if len(key) != 32 {
		return nil, errors.New("key len!=32")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return aead, nil
}

func StringToAesKey(password string, keyLen int) []byte {
	var b, prev []byte
	h := md5.New()
	for len(b) < keyLen {
		h.Write(prev)
		h.Write([]byte(password))
		c := h.Sum(b)
		b = c
		prev = b[len(b)-h.Size():]
		h.Reset()
	}
	return b[:keyLen]
}

func Copy(dst io.Writer, src io.Reader) error {
	_, err := io.Copy(dst, src)
	return err
}
