package main

import (
	"math/bits"
)

const NUM_BITS_FIRST_LOOKUP = 9

type SymbolLengthPair struct {
	Symbol uint32
	Length uint32
}

type HuffmanDecoder struct {
	lookup        []SymbolLengthPair
	primaryMask   uint32
	secondaryMask uint32
}

func NewHuffmanDecoder(codebook *Codebook) *HuffmanDecoder {
	var nbits uint32
	var secondaryMask uint32
	if codebook.MaxLength > NUM_BITS_FIRST_LOOKUP {
		nbits = NUM_BITS_FIRST_LOOKUP
		secondaryMask = (1 << (codebook.MaxLength - NUM_BITS_FIRST_LOOKUP)) - 1
	} else {
		nbits = codebook.MaxLength
	}
	var primaryMask uint32 = (1 << nbits) - 1

	lookup := make([]SymbolLengthPair, 1<<nbits)
	for symbol, pair := range codebook.Book {
		if pair.Length == 0 {
			continue
		}

		bitcode := uint32(bits.Reverse16(uint16(pair.Bitcode)))
		bitcode >>= 16 - pair.Length
		if pair.Length <= nbits {
			delta := nbits - pair.Length
			for idx := uint32(0); idx < 1<<delta; idx++ {
				lookup[int(bitcode|(idx<<pair.Length))] = SymbolLengthPair{uint32(symbol), pair.Length}
			}
		} else {
			base := int(bitcode & primaryMask)
			var offset uint32
			if lookup[base].Symbol == 0 {
				offset = uint32(len(lookup))
				lookup[base] = SymbolLengthPair{offset, pair.Length}
				newLen := len(lookup) + (1 << (codebook.MaxLength - nbits))
				lookup = append(lookup, make([]SymbolLengthPair, newLen-len(lookup))...)

			} else {
				offset = lookup[base].Symbol
			}

			secondaryLen := pair.Length - nbits
			base = int(offset + ((bitcode >> nbits) & secondaryMask))
			for idx := 0; idx < 1<<(codebook.MaxLength-pair.Length); idx++ {
				lookup[base+(idx<<int(secondaryLen))] = SymbolLengthPair{uint32(symbol), pair.Length}
			}

		}
	}

	return &HuffmanDecoder{lookup, primaryMask, secondaryMask}
}

func (d *HuffmanDecoder) Decode(bits uint32) (*SymbolLengthPair, error) {
	pair := d.lookup[(bits & d.primaryMask)]
	if pair.Length == 0 {
		return nil, NewError(HuffmanDecoderCodeNotFound)
	}
	if pair.Length <= NUM_BITS_FIRST_LOOKUP {
		return &pair, nil
	}
	base := int(pair.Symbol)
	idx := (bits >> NUM_BITS_FIRST_LOOKUP) & d.secondaryMask
	return &d.lookup[base+int(idx)], nil
}
