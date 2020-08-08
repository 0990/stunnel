package stunnel

import (
	"bytes"
	"fmt"
	"testing"
)

func Test_Cipher(t *testing.T) {
	ahead, err := createAesGcmAEAD(kdf("xujialong", 32))
	if err != nil {
		t.Fatal(err)
	}

	data := []byte{1, 2, 3}
	//in := bytes.NewBuffer(data)

	rw := &ReadWriter{buf: make([]byte, 100)}

	c := newCipherConn(rw, ahead)
	//	out := make([]byte, payloadSizeMask)
	_, err = c.Write(data)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(rw.buf)

	//rw2 := &ReadWriter{buf: rw.buf}
	c2 := newCipherConn(bytes.NewBuffer(rw.buf), ahead)

	buf := make([]byte, 100)
	n, _ := c2.Read(buf)
	fmt.Println(buf[0:n])

}

type ReadWriter struct {
	buf []byte
}

func (p *ReadWriter) Write(b []byte) (int, error) {
	n := copy(p.buf, b)
	return int(n), nil
}

func (w *ReadWriter) Read(b []byte) (int, error) {
	copy(b, w.buf)
	return len(w.buf), nil
}
