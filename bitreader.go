package main

import (
	"encoding/binary"
	"errors"
	"io"
)

const bufferSize = 16 << 10

type BitRead interface {
	PeekBits() (uint32, error)
	Consume(n int)
	ByteAlign()
	ReadBits(n int) (uint32, error)
	HasDataLeft() (bool, error)
	Read(p []byte) (n int, err error)
	ReadExact(p []byte) error
}

type BitReader struct {
	reader     io.Reader
	nbits      int
	buf        []byte
	begin, cap int
}

func NewBitReader(reader io.Reader) *BitReader {
	return &BitReader{
		reader: reader,
		nbits:  0,
		buf:    make([]byte, bufferSize),
		begin:  0,
		cap:    0,
	}
}

func (r *BitReader) Read(b []byte) (n int, err error) {
	r.ByteAlign()
	n = min(len(b), len(r.buffer()))
	copy(b, r.buffer()[:n])
	r.begin += n

	m, err := r.reader.Read(b[n:])
	n += m
	return
}

func (r *BitReader) PeekBits() (uint32, error) {
	for len(r.buffer()) < 4 {
		n, err := r.fillBuf()
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, errors.New("unexpected EOF")
		}
	}
	bits := binary.LittleEndian.Uint32(r.buffer())
	// right shift by the number of consumed bits
	return bits >> r.nbits, nil
}

func (r *BitReader) Consume(n int) {
	r.nbits += n
	r.begin += (r.nbits / 8)
	r.nbits %= 8
}

func (r *BitReader) ByteAlign() {
	if r.nbits > 0 {
		r.nbits = 0
		r.begin += 1
	}
}

func (r *BitReader) HasDataLeft() (bool, error) {
	if len(r.buffer()) > 0 {
		return true, nil
	}
	n, err := r.fillBuf()
	return n > 0, err
}

func (r *BitReader) ReadBits(n int) (uint32, error) {
	bits, err := r.PeekBits()
	r.Consume(n)
	return bits & ((1 << n) - 1), err
}

const BUFFER_SIZE int = 16 << 10

func (r *BitReader) buffer() []byte {
	return r.buf[r.begin:r.cap]
}

func (r *BitReader) bitLen() int {
	return len(r.buffer())*8 - r.nbits
}

func (r *BitReader) fillBuf() (int, error) {
	copy(r.buf, r.buffer())
	r.cap -= r.begin
	r.begin = 0
	n, err := r.reader.Read(r.buf[r.cap:])

	r.cap += n
	return n, err
}

func (r *BitReader) ReadExact(p []uint8) error {
	begin := 0
	for begin < len(p) {
		n, err := r.Read(p[begin:])
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrUnexpectedEOF
		}
		begin += n
	}
	return nil
}
