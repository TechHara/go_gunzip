package main

type Codebook struct {
	Book      []CodeLengthPair
	MaxLength uint32
}

type CodeLengthPair struct {
	Bitcode uint32
	Length  uint32
}

const MAX_CODELENGTH = 15
const MAX_LL_SYMBOL = 288

func NewCodebook(lengths []uint32) (*Codebook, error) {
	err := NewError(InvalidCodeLengths)

	if len(lengths) == 0 || len(lengths) > int(MAX_LL_SYMBOL)+1 {
		return nil, err
	}

	book := make([]CodeLengthPair, len(lengths))
	var maxLen uint32

	var blCount [MAX_CODELENGTH + 1]uint32
	for i, l := range lengths {
		blCount[l] += 1
		book[i] = CodeLengthPair{0, l}
		maxLen = max(maxLen, l)
	}

	if maxLen > MAX_CODELENGTH {
		return nil, err
	}

	var nextCode [MAX_CODELENGTH + 1]uint32
	var code uint32
	blCount[0] = 0
	for bits := uint32(1); bits <= maxLen; bits++ {
		code = (code + blCount[bits-1]) << 1
		nextCode[bits] = code
	}

	for i := 0; i < len(book); i++ {
		l := int(book[i].Length)
		if l != 0 {
			book[i].Bitcode = nextCode[l]
			nextCode[l] += 1
		}
	}

	return &Codebook{book, maxLen}, nil
}

func NewDefaultLLCodebook() *Codebook {
	lengths := []uint32{
		8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
		8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
		8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
		8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
		8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 9,
		9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9,
		9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9,
		9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9,
		9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 7, 7, 7, 7, 7,
		7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 8, 8, 8, 8, 8, 8, 8, 8,
	}
	codebook, _ := NewCodebook(lengths)
	return codebook
}

func NewDefaultDistCodebook() *Codebook {
	lengths := []uint32{
		5, 5, 5, 5, 5, 5, 5, 5, 5, 5,
		5, 5, 5, 5, 5, 5, 5, 5, 5, 5,
		5, 5, 5, 5, 5, 5, 5, 5, 5, 5,
	}
	codebook, _ := NewCodebook(lengths)
	return codebook
}
