package book

import (
	"testing"
)

func TestStdin(t *testing.T) {
	input, _ := NewStdinInput()
	input.Close()
}

func TestFile(t *testing.T) {
	input, _ := NewFileInput("")
	input.Close()
}

func TestBytes(t *testing.T) {
	b := make([]byte, 0)
	input, _ := NewBytesInput(b)
	input.Close()
}

func TestEachLineBytes(t *testing.T) {
	b := []byte("foo\nbar\n")
	res := ""
	input, _ := NewBytesInput(b)
	input.EachLineBytes(func(text []byte) {
		res += string(text)
	})
	input.Close()
	if res != "foobar" {
		t.Error("EachLineBytes doesn't read bytes.")
	}
}

func TestEachLineBytesLong(t *testing.T) {
	b := make([]byte, 1000*1000)
	for i := len(b) - 1; i >= 0; i-- {
		b[i] = 'a'
	}
	res := ""
	input, _ := NewBytesInput(b)
	input.EachLineBytes(func(text []byte) {
		res += string(text)
	})
	input.Close()
	if res != string(b) {
		t.Error("EachLineBytes doesn't read bytes.")
	}
}

func TestEachLineBytesTooLongAndSkipLine(t *testing.T) {
	b := make([]byte, 1000*1000*3)
	for i := len(b) - 1; i >= 0; i-- {
		b[i] = 'a'
	}
	var res bool
	input, _ := NewBytesInput(b)
	input.EachLineBytes(func(text []byte) {
		res = (text != nil)
	})
	input.Close()
	if res {
		t.Error("EachLineBytes doesn't skip bytes.")
	}
}
