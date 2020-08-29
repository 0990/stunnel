package crypto

import (
	"bytes"
	"crypto/cipher"
	"io"
)

const payloadSizeMask = 0x3FFF // 16*1024 - 1

type conn struct {
	io.ReadWriter
	r *reader
	w *writer
}

func NewConn(rw io.ReadWriter, aead cipher.AEAD) io.ReadWriter {
	return &conn{
		ReadWriter: rw,
		r:          newReader(rw, aead),
		w:          newWriter(rw, aead),
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

func newWriter(w io.Writer, aead cipher.AEAD) *writer {
	return &writer{
		Writer: w,
		AEAD:   aead,
		buf:    make([]byte, 2+aead.Overhead()+payloadSizeMask+aead.Overhead()),
		nonce:  make([]byte, aead.NonceSize()),
	}
}

func (w *writer) Write(b []byte) (int, error) {
	n, err := w.ReadFrom(bytes.NewBuffer(b))
	return int(n), err
}

func (w *writer) ReadFrom(r io.Reader) (n int64, err error) {
	for {
		buf := w.buf
		payloadBuf := buf[2+w.Overhead() : 2+w.Overhead()+payloadSizeMask]
		nr, er := r.Read(payloadBuf)
		if nr > 0 {
			n += int64(nr)
			buf = buf[:2+w.Overhead()+nr+w.Overhead()]
			payloadBuf = payloadBuf[:nr]
			buf[0], buf[1] = byte(nr>>8), byte(nr)
			w.Seal(buf[:0], w.nonce, buf[:2], nil)
			w.Seal(payloadBuf[:0], w.nonce, payloadBuf, nil)
			_, ew := w.Writer.Write(buf)
			if ew != nil {
				err = ew
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return n, err
}

type reader struct {
	io.Reader
	cipher.AEAD
	nonce    []byte
	buf      []byte
	leftover []byte
}

func newReader(r io.Reader, aead cipher.AEAD) *reader {
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
	buf := r.buf[:2+r.Overhead()]
	_, err := io.ReadFull(r.Reader, buf)
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
