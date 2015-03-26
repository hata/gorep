package book

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
)

const (
	defaultBufferSize           = 4096
	defaultOverflowNewArraySize = 16
	defaultMaxLineSize          = 1024 * 1024 * 2 // 1MB
)

type LineBytesHandler func(text []byte)

type Input interface {
	EachLineBytes(handler LineBytesHandler) error
	io.Closer
}

type baseInput struct {
	fileReader *bufio.Reader
}

type fileInput struct {
	baseInput
	fileDescriptor *os.File
}

type stdinInput struct {
	fileInput
}

type bytesInput struct {
	baseInput
	bytes []byte
}

func NewStdinInput() (input Input, err error) {
	return newStdinInput()
}

func NewFileInput(file string) (input Input, err error) {
	return newFileInput(file)
}

func NewBytesInput(bytes []byte) (input Input, err error) {
	return newBytesInput(bytes)
}

func newStdinInput() (input *stdinInput, err error) {
	input = new(stdinInput)
	input.fileDescriptor = os.Stdin
	input.fileReader = bufio.NewReaderSize(input.fileDescriptor, defaultBufferSize)
	err = nil
	return
}

func newFileInput(file string) (input *fileInput, err error) {
	input = new(fileInput)
	input.fileDescriptor, err = os.Open(file)
	input.fileReader = bufio.NewReaderSize(input.fileDescriptor, defaultBufferSize)
	return
}

func newBytesInput(byteData []byte) (input *bytesInput, err error) {
	input = new(bytesInput)
	input.fileReader = bufio.NewReaderSize(bytes.NewReader(byteData), defaultBufferSize)
	return
}

func (input *baseInput) EachLineBytes(handler LineBytesHandler) error {
	return eachLineBytes(input.fileReader, handler)
}

// TODO: Need to set a limit to avoid continuing allocation. It is like
// to skip parsing this line.
func eachLineBytes(reader *bufio.Reader, handler LineBytesHandler) error {
	for {
		line, isPrefix, err := reader.ReadLine()
		if isPrefix {
			total := 0
			n := 0
			lineFragments := make([][]byte, defaultOverflowNewArraySize)
			lineFragments[n] = line
			total += len(line)
			for isPrefix {
				n++
				line, isPrefix, err = reader.ReadLine()
				if n >= len(lineFragments) {
					newFragments := make([][]byte, len(lineFragments)+defaultOverflowNewArraySize)
					copy(newFragments, lineFragments)
					lineFragments = newFragments
				}
				lineFragments[n] = line
				total += len(line)
				if err != nil {
					break
				}

				// When the reading line size is too big, then
				// skip the line.
				if total > defaultMaxLineSize {
					for isPrefix {
						line, isPrefix, err = reader.ReadLine()
					}
					fmt.Println("Skip a line.")
					line = nil
					break
				}
			}

			if total < defaultMaxLineSize {
				first := 0
				line = make([]byte, total)
				for i := 0; i < len(lineFragments); i++ {
					partLen := len(lineFragments[i])
					if partLen > 0 {
						dest := line[first : first+partLen]
						copy(dest, lineFragments[i])
						first += partLen
					}
				}
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		} else if line != nil {
			handler(line)
		}
	}

	return nil
}

func (input *stdinInput) Close() error {
	return nil
}

func (input *fileInput) Close() error {
	if input.fileDescriptor != nil {
		return input.fileDescriptor.Close()
	} else {
		return nil
	}
}

func (input *bytesInput) Close() error {
	return nil
}
