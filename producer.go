package main

type State int

const (
	StateHeader State = iota
	StateBlock
	StateFooter
)

type ProduceTag int

const (
	ProduceHeader ProduceTag = iota
	ProduceFooter
	ProduceData
)

type Produce struct {
	Tag  ProduceTag
	Head *Header
	Foot *Footer
	Data []uint8
}

type Producer struct {
	reader    BitRead
	state     State
	memberIdx int
	window    SlidingWindow
}

func NewProducer(reader BitRead) *Producer {
	return &Producer{reader, StateHeader, 0, *NewSlidingWindow()}
}

// returns nil as producer if done
func (p *Producer) Next() (*Produce, error) {
	if p.state == StateHeader {
		dataLeft, err := p.reader.HasDataLeft()
		if err != nil {
			return nil, err
		}
		if !dataLeft {
			if p.memberIdx == 0 {
				return nil, NewError(EmptyInput)
			} else {
				return nil, nil
			}
		}
		p.state = StateBlock
		p.memberIdx += 1
		header, err := ReadHeader(p.reader)
		return &Produce{ProduceHeader, header, nil, nil}, err
	} else if p.state == StateBlock {
		header, err := p.reader.ReadBits(3)
		if err != nil {
			return nil, err
		}
		if header&1 == 1 {
			p.state = StateFooter
		}

		if header&0b110 == 0b000 {
			return p.inflateBlock0()
		} else if header&0b110 == 0b010 {
			return p.inflateBlock1()
		} else if header&0b110 == 0b100 {
			return p.inflateBlock2()
		} else {
			return nil, NewError(InvalidBlockType)
		}
	} else if p.state == StateFooter {
		p.state = StateHeader
		footer, err := ReadFooter(p.reader)
		return &Produce{ProduceFooter, nil, footer, nil}, err
	}
	panic("unreachable")
}

func (p *Producer) inflateBlock0() (*Produce, error) {
	p.reader.ByteAlign()
	length, err := p.reader.ReadBits(16)
	if err != nil {
		return nil, err
	}
	nlength, err := p.reader.ReadBits(16)
	if err != nil {
		return nil, err
	}
	if length^nlength != 0xFFFF {
		return nil, NewError(BlockType0LenMismatch)
	}
	buf := make([]uint8, int(length))
	err = p.reader.ReadExact(buf)
	if err != nil {
		return nil, err
	}
	copy(p.window.WriteBuffer()[:length], buf)
	p.window.Slide((int(length)))
	return &Produce{ProduceData, nil, nil, buf}, nil
}

func (p *Producer) inflateBlock1() (*Produce, error) {
	llDecoder := NewHuffmanDecoder(NewDefaultLLCodebook())
	distDecoder := NewHuffmanDecoder(NewDefaultDistCodebook())
	return p.inflate(llDecoder, distDecoder)
}

func (p *Producer) inflateBlock2() (*Produce, error) {
	llDecoder, distDecoder, err := p.readDynamicCodebooks()
	if err != nil {
		return nil, err
	}
	return p.inflate(llDecoder, distDecoder)
}

func (p *Producer) inflate(llDecoder *HuffmanDecoder, distDecoder *HuffmanDecoder) (*Produce, error) {
	iter := NewCodeIterator(p.reader, llDecoder, distDecoder)
	done := false
	buf := make([]uint8, 0)
	for {
		boundary := p.window.Boundary
		decodeResult, err := Decode(p.window.Data, boundary, iter)
		if err != nil {
			return nil, err
		}
		if decodeResult.Tag == Done {
			done = true
		}

		n := int(decodeResult.N)
		buf = append(buf, p.window.WriteBuffer()[:n]...)

		p.window.Slide(n)
		if done {
			break
		}
	}
	return &Produce{ProduceData, nil, nil, buf}, nil
}

func (p *Producer) readDynamicCodebooks() (*HuffmanDecoder, *HuffmanDecoder, error) {
	hlit, err := p.reader.ReadBits(5)
	if err != nil {
		return nil, nil, err
	}
	hlit += 257

	hdist, err := p.reader.ReadBits(5)
	if err != nil {
		return nil, nil, err
	}
	hdist += 1

	hclen, err := p.reader.ReadBits(4)
	if err != nil {
		return nil, nil, err
	}
	hclen += 4

	clLengths := make([]uint32, 19)
	for i, idx := range []int{16, 17, 18, 0, 8, 7, 9, 6, 10, 5, 11, 4, 12, 3, 13, 2, 14, 1, 15} {
		if i >= int(hclen) {
			break
		}
		length, err := p.reader.ReadBits(3)
		if err != nil {
			return nil, nil, err
		}
		clLengths[idx] = length
	}
	clCodebook, err := NewCodebook(clLengths)
	if err != nil {
		return nil, nil, err
	}
	clDecoder := NewHuffmanDecoder(clCodebook)

	numCodes := int(hlit + hdist)
	lengths := make([]uint32, 0)
	for len(lengths) < numCodes {
		bits, err := p.reader.PeekBits()
		if err != nil {
			return nil, nil, err
		}
		pair, err := clDecoder.Decode(bits)
		if err != nil {
			return nil, nil, err
		}
		p.reader.Consume(int(pair.Length))
		if pair.Symbol <= 15 {
			lengths = append(lengths, pair.Symbol)
		} else if pair.Symbol == 16 {
			length, err := p.reader.ReadBits(2)
			if err != nil {
				return nil, nil, err
			}
			length += 3
			x := lengths[len(lengths)-1]
			for i := 0; i < int(length); i++ {
				lengths = append(lengths, x)
			}
		} else if pair.Symbol == 17 {
			length, err := p.reader.ReadBits(3)
			if err != nil {
				return nil, nil, err
			}
			length += 3
			for i := 0; i < int(length); i++ {
				lengths = append(lengths, 0)
			}
		} else if pair.Symbol == 18 {
			length, err := p.reader.ReadBits(7)
			if err != nil {
				return nil, nil, err
			}
			length += 11
			for i := 0; i < int(length); i++ {
				lengths = append(lengths, 0)
			}
		} else {
			panic("unreachable")
		}
	}

	if len(lengths) != numCodes {
		return nil, nil, NewError(ReadDynamicCodebook)
	}

	llCodes, err := NewCodebook(lengths[:hlit])
	if err != nil {
		return nil, nil, err
	}
	distCodes, err := NewCodebook(lengths[hlit:])
	if err != nil {
		return nil, nil, err
	}
	return NewHuffmanDecoder(llCodes), NewHuffmanDecoder(distCodes), nil
}
