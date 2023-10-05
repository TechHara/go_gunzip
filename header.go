package main

import (
	"bytes"
	"encoding/binary"
	"io"
)

const (
	ID1      = 0x1f
	ID2      = 0x8b
	DEFLATE  = 8
	FTEXT    = 1
	FHCRC    = 2
	FEXTRA   = 4
	FNAME    = 8
	FCOMMENT = 16
)

type Header struct {
	Header     [10]byte
	ExtraField []byte
	Name       []byte
	Comment    []byte
	Crc16      uint16
	Size       int // header size
}

func ReadHeader(r BitRead) (*Header, error) {
	var h Header
	err := r.ReadExact(h.Header[:])
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(h.Header[:3], []byte{ID1, ID2, DEFLATE}) || h.Header[3]&0b11100000 != 0 {
		return nil, NewError(InvalidGzHeader)
	}
	h.Size = len(h.Header)
	if h.getFlg()&FEXTRA != 0 {
		var n uint16
		err := binary.Read(r, binary.LittleEndian, &n)
		if err != nil {
			return nil, err
		}
		h.Size += 2 // size of n
		h.ExtraField = make([]byte, n)
		err = binary.Read(r, binary.LittleEndian, &h.ExtraField)
		if err != nil {
			return nil, err
		}
		h.Size += int(n) // size of extra field
	}
	if h.getFlg()&FNAME != 0 {
		h.Name, err = readUntil(r, 0)
		if err != nil {
			return nil, err
		}
		h.Size += len(h.Name) // size of name
	}
	if h.getFlg()&FCOMMENT != 0 {
		h.Comment, err = readUntil(r, 0)
		if err != nil {
			return nil, err
		}
		h.Size += len(h.Comment) // size of comment
	}
	if h.getFlg()&FHCRC != 0 {
		err := binary.Read(r, binary.LittleEndian, &h.Crc16)
		if err != nil {
			return nil, err
		}
		h.Size += 2 // size of crc16
	}
	return &h, nil
}

// getFlg returns the FLG byte of the header
func (h *Header) getFlg() byte {
	return h.Header[3]
}

func readUntil(r io.Reader, delim byte) ([]byte, error) {
	var buf bytes.Buffer
	var b [1]byte
	for {
		err := binary.Read(r, binary.LittleEndian, &b)
		if err != nil {
			return nil, err
		}
		buf.Write(b[:])
		if b[0] == delim {
			break
		}
	}
	return buf.Bytes(), nil
}
