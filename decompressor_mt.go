package main

import (
	"io"
)

type DecompressorMultithreaded struct {
	c <-chan struct {
		*Produce
		error
	}
	buf      []uint8
	begin    int
	checksum Checksum
}

func NewDecompressorMultithreaded(reader io.Reader) *DecompressorMultithreaded {
	bitreader := NewBitReader(reader)
	producer := NewProducer(bitreader)
	checksum := NewCrc32()

	c := make(chan struct {
		*Produce
		error
	})
	go produce(producer, c)

	return &DecompressorMultithreaded{c, make([]uint8, 0), 0, checksum}
}

func produce(producer *Producer, c chan<- struct {
	*Produce
	error
}) {
	for {
		produce, err := producer.Next()
		done := produce == nil
		c <- struct {
			*Produce
			error
		}{produce, err}
		if done {
			break
		}
	}
}

func (d *DecompressorMultithreaded) fillBuffer() (int, error) {
	for pair := range d.c {
		produce := pair.Produce
		err := pair.error

		if err != nil {
			return 0, err
		}
		if produce == nil {
			return 0, nil
		}
		if produce.Tag == ProduceHeader {
			// nothing to do
		} else if produce.Tag == ProduceFooter {
			footer := produce.Foot
			if d.checksum.Checksum() != footer.Crc32 {
				return 0, NewError(ChecksumMismatch)
			}
			if d.checksum.Len()&0xFFFFFFFF != int(footer.Size) {
				return 0, NewError(SizeMismatch)
			}
			d.checksum.ResetLen()
		} else if produce.Tag == ProduceData {
			xs := produce.Data
			if len(xs) == 0 {
				continue
			}
			d.checksum.Update(xs)
			d.buf = xs
			d.begin = 0
			return len(xs), nil
		}
	}
	panic("unreachable")
}

func (d *DecompressorMultithreaded) Read(buf []uint8) (int, error) {
	nbytes := 0
	idx := 0
	for {
		n := min(len(buf[idx:]), len(d.buf[d.begin:]))
		copy(buf[idx:idx+n], d.buf[d.begin:d.begin+n])
		idx += n
		nbytes += n
		d.begin += n
		if idx == len(buf) {
			break
		}
		filled, err := d.fillBuffer()
		if err != nil {
			return nbytes, err
		}
		if filled == 0 {
			break
		}
	}

	return nbytes, nil
}
