package crypto

import (
	"crypto/cipher"
	"io"
	"math/rand"
	"time"
)

//const payloadSizeMask = 0x3FFF // 16*1024 - 1
const payloadSizeMask = 65535

func init() {
	rand.Seed(time.Now().UnixNano())
}

type conn struct {
	io.ReadWriter
	r *reader
	w *writer
}

func NewConn(rw io.ReadWriter, aead cipher.AEAD) io.ReadWriter {
	return &conn{
		ReadWriter: rw,
		r:          NewReader(rw, aead),
		w:          NewWriter(rw, aead),
	}
}

func (p *conn) Read(b []byte) (int, error) {
	return p.r.Read(b)
}

func (c *conn) Write(b []byte) (int, error) {
	return c.w.Write(b)
}

type writer struct {
	io.Writer
	cipher.AEAD
	nonce []byte
	buf   []byte
}

func NewWriter(w io.Writer, aead cipher.AEAD) *writer {
	return &writer{
		Writer: w,
		AEAD:   aead,
		buf:    make([]byte, 2+aead.Overhead()+payloadSizeMask+aead.Overhead()),
		nonce:  make([]byte, aead.NonceSize()),
	}
}

func (w *writer) RandomNonce() {

}

func (w *writer) Write(b []byte) (int, error) {
	n, err := w.write(b)
	return int(n), err
}

// nonce + payloadsize(encrypt)+ payload(encrypt)
func (w *writer) write(b []byte) (n int64, err error) {
	if len(b) == 0 {
		return 0, nil
	}

	//create new nonce
	rand.Read(w.nonce)

	buf := w.buf
	sizeBuf := buf[w.NonceSize():]
	payloadBuf := buf[2+w.Overhead()+w.NonceSize() : 2+w.Overhead()+w.NonceSize()+payloadSizeMask]
	nr := len(b)
	//copy(payloadBuf, b)

	n += int64(nr)
	end := w.NonceSize() + 2 + w.Overhead() + nr + w.Overhead()
	buf = buf[:end]
	//payloadBuf = payloadBuf[:nr]
	sizeBuf[0], sizeBuf[1] = byte(nr>>8), byte(nr)

	//nonce
	copy(buf, w.nonce)
	//payloadsize
	w.Seal(sizeBuf[:0], w.nonce, sizeBuf[:2], nil)
	//payload
	w.Seal(payloadBuf[:0], w.nonce, b, nil)
	_, ew := w.Writer.Write(buf)
	if ew != nil {
		return n, ew
	}

	return n, nil
}

type reader struct {
	io.Reader
	cipher.AEAD
	nonce    []byte
	buf      []byte
	leftover []byte
}

func NewReader(r io.Reader, aead cipher.AEAD) *reader {
	return &reader{
		Reader: r,
		AEAD:   aead,
		nonce:  make([]byte, aead.NonceSize()),
		buf:    make([]byte, payloadSizeMask+aead.Overhead()),
	}
}

func (r *reader) Read(b []byte) (int, error) {
	if len(r.leftover) > 0 {
		n := copy(b, r.leftover)
		r.leftover = r.leftover[n:]
		return n, nil
	}

	n, err := r.read()
	m := copy(b, r.buf[:n])
	if m < n {
		r.leftover = r.buf[m:n]
	}
	return m, err
}

func (r *reader) read() (int, error) {
	_, err := io.ReadFull(r.Reader, r.nonce)
	if err != nil {
		return 0, err
	}
	buf := r.buf[:2+r.Overhead()]
	_, err = io.ReadFull(r.Reader, buf)
	if err != nil {
		return 0, err
	}
	_, err = r.Open(buf[:0], r.nonce, buf, nil)
	if err != nil {
		return 0, err
	}

	size := (int(buf[0])<<8 + int(buf[1])) & payloadSizeMask
	buf = r.buf[:size+r.Overhead()]
	_, err = io.ReadFull(r.Reader, buf)
	if err != nil {
		return 0, err
	}
	_, err = r.Open(buf[:0], r.nonce, buf, nil)
	if err != nil {
		return 0, err
	}
	return size, nil
}
