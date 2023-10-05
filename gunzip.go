package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	reader := os.Stdin
	writer := os.Stdout

	defer reader.Close()
	defer writer.Close()

	args := os.Args
	var decompressor io.Reader
	if len(args) == 2 && args[1] == "-t" {
		decompressor = NewDecompressorMultithreaded(reader)
	} else if len(args) == 1 {
		decompressor = NewDecompressor(reader)
	} else {
		fmt.Printf("Usage: %s [-t]\n", args[0])
		os.Exit(-1)
	}

	_, err := io.Copy(writer, decompressor)
	if err != nil {
		log.Fatal(err)
	}
}
