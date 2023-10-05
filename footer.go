package main

import (
	"encoding/binary"
	"io"
)

type Footer struct {
	Crc32 uint32
	Size  uint32
}

func ReadFooter(reader io.Reader) (*Footer, error) {
	var f Footer
	err := binary.Read(reader, binary.LittleEndian, &f)
	return &f, err
}
