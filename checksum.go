package main

import (
	"hash/crc32"
)

type Checksum interface {
	Update(xs []byte)
	Checksum() uint32
	Len() int
	ResetLen()
}

type Crc32 struct {
	sum uint32
	n   int
}

func NewCrc32() *Crc32 {
	return &Crc32{0, 0}
}

func (c *Crc32) Update(xs []byte) {
	c.n += len(xs)
	c.sum = crc32.Update(c.sum, crc32.IEEETable, xs)
}

func (c *Crc32) Checksum() uint32 {
	sum := c.sum
	c.sum = 0
	return sum
}

func (c *Crc32) Len() int {
	return c.n
}

func (c *Crc32) ResetLen() {
	c.n = 0
}
