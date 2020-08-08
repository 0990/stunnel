package stunnel

import (
	"bytes"
	"fmt"
	"testing"
)

func TestRSA(t *testing.T) {
	pri, pub, err := GenRSAKey(1024)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("public key:\n%s\n", string(pub))
	fmt.Printf("private key:\n%s\n", string(pri))

	data := []byte("hello 0990")
	chipher, err := RSAEncrypt(pub, data)
	if err != nil {
		t.Fatal(err)
	}

	out, err := RSADecrypt(pri, chipher)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(data, out) {
		t.FailNow()
	}
}
