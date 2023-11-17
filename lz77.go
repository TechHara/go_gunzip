package main

const END_OF_BLOCK = 256
const MAX_DISTANCE = 1 << 15 // 32kB
const MAX_LENGTH = 258

type Code uint8

const (
	Literal Code = iota
	EndOfBlock
	Dictionary
)

type CodeData struct {
	Tag      Code
	Value    uint8
	Distance uint16
	Length   uint16
}

func NewLiteral(value uint8) CodeData {
	return CodeData{Literal, value, 0, 0}
}

func NewEndOfBlock() CodeData {
	return CodeData{EndOfBlock, 0, 0, 0}
}

func NewDictionary(length uint16, distance uint16) CodeData {
	return CodeData{Dictionary, 0, distance, length}
}

type Result uint32

const (
	Done Result = iota
	WindowsIsFull
)

type DecodeResult struct {
	Tag Result
	N   uint32
}

func Decode(window []uint8, boundary int, reader BitRead, llDecoder *HuffmanDecoder, distDecoder *HuffmanDecoder) (*DecodeResult, error) {
	idx := boundary
	if idx+MAX_LENGTH >= len(window) {
		return &DecodeResult{WindowsIsFull, uint32(idx - boundary)}, nil
	}
	for {
		code, err := ReadNextCode(reader, llDecoder, distDecoder)
		if err != nil {
			return nil, err
		}
		if code.Tag == Literal {
			window[idx] = code.Value
			idx += 1
		} else if code.Tag == Dictionary {
			distance := int(code.Distance)
			length := int(code.Length)
			if distance > idx {
				return nil, NewError(DistanceTooMuch)
			}
			begin := idx - distance
			for length > 0 {
				n := min(distance, length)
				copy(window[idx:idx+n], window[begin:begin+n])
				idx += n
				length -= n
				distance += n
			}
		} else if code.Tag == EndOfBlock {
			return &DecodeResult{Done, uint32(idx - boundary)}, nil
		}
		if idx+MAX_LENGTH >= len(window) {
			return &DecodeResult{WindowsIsFull, uint32(idx - boundary)}, nil
		}
	}
}

func ReadNextCode(reader BitRead, llDecoder *HuffmanDecoder, distDecoder *HuffmanDecoder) (CodeData, error) {
	bitcode, err := reader.PeekBits()
	if err != nil {
		return CodeData{}, err
	}
	pair, err := llDecoder.Decode(bitcode)
	if err != nil {
		return CodeData{}, err
	}
	reader.Consume(int(pair.Length))
	if pair.Symbol == END_OF_BLOCK {
		return NewEndOfBlock(), nil
	} else if pair.Symbol < END_OF_BLOCK {
		return NewLiteral(uint8(pair.Symbol)), nil
	}
	bitsLength := SYMBOL2BITS_LENGTH[int(pair.Symbol&0xFF)]
	length, err := reader.ReadBits(int(bitsLength[0]))
	if err != nil {
		return CodeData{}, err
	}
	bitsLength[1] += length
	bitcode, err = reader.PeekBits()
	if err != nil {
		return CodeData{}, err
	}
	pair, err = distDecoder.Decode(bitcode)
	if err != nil {
		return CodeData{}, err
	}
	reader.Consume(int(pair.Length))
	bitsDistance := SYMBOL2BITS_DISTANCE[int(pair.Symbol)]
	dist, err := reader.ReadBits(int(bitsDistance[0]))
	if err != nil {
		return CodeData{}, err
	}
	bitsDistance[1] += dist
	return NewDictionary(uint16(bitsLength[1]), uint16(bitsDistance[1])), nil
}

var SYMBOL2BITS_LENGTH = [][2]uint32{
	{0, 0},
	{0, 3},
	{0, 4},
	{0, 5},
	{0, 6},
	{0, 7},
	{0, 8},
	{0, 9},
	{0, 10},
	{1, 11},
	{1, 13},
	{1, 15},
	{1, 17},
	{2, 19},
	{2, 23},
	{2, 27},
	{2, 31},
	{3, 35},
	{3, 43},
	{3, 51},
	{3, 59},
	{4, 67},
	{4, 83},
	{4, 99},
	{4, 115},
	{5, 131},
	{5, 163},
	{5, 195},
	{5, 227},
	{0, 258},
}

var SYMBOL2BITS_DISTANCE = [][2]uint32{
	{0, 1},
	{0, 2},
	{0, 3},
	{0, 4},
	{1, 5},
	{1, 7},
	{2, 9},
	{2, 13},
	{3, 17},
	{3, 25},
	{4, 33},
	{4, 49},
	{5, 65},
	{5, 97},
	{6, 129},
	{6, 193},
	{7, 257},
	{7, 385},
	{8, 513},
	{8, 769},
	{9, 1025},
	{9, 1537},
	{10, 2049},
	{10, 3073},
	{11, 4097},
	{11, 6145},
	{12, 8193},
	{12, 12289},
	{13, 16385},
	{13, 24577},
}
