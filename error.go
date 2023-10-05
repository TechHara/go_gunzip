package main

import (
	"fmt"
)

// Error is a custom error type for the package
type Error struct {
	Kind ErrorKind
}

// ErrorKind is an enum for the kinds of errors
type ErrorKind int

const (
	StdIoError ErrorKind = iota
	EmptyInput
	InvalidGzHeader
	InvalidBlockType
	BlockType0LenMismatch
	InvalidCodeLengths
	HuffmanDecoderCodeNotFound
	DistanceTooMuch
	EndOfBlockNotFound
	ReadDynamicCodebook
	ChecksumMismatch
	SizeMismatch
)

// Error implements the error interface for Error type
func (e *Error) Error() string {
	return fmt.Sprintf("%v", e.Kind)
}

// NewError creates a new Error with the given kind and message
func NewError(kind ErrorKind) *Error {
	return &Error{Kind: kind}
}
