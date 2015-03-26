package book

import (
	"testing"
)

func TestNewPageLength(t *testing.T) {
	p := newPage(&PageParams{Capacity: 16})
	if p.length != 0 {
		t.Error("New Page should be zero length.")
	}
}

func TestNewPageCapacity(t *testing.T) {
	p := newPage(&PageParams{Capacity: 16})
	if p.capacity != 16 {
		t.Error("New Page should be initialized capacity.")
	}
	p2 := NewPage(&PageParams{Capacity: 16})
	if p2.Capacity() != 16 {
		t.Error("NewPage should return a new instance.")
	}
}

func TestAddLineBytes(t *testing.T) {
	p := newPage(&PageParams{Capacity: 3})
	p.AddLineBytes([]byte("foo"))
	if p.Length() != 1 {
		t.Error("Length is increased by adding a new line.")
	}
	p.AddLineBytes([]byte("bar"))
	if p.Length() != 2 {
		t.Error("Length is increased by adding a new line.")
	}
	if !p.IsEnoughCapacity() {
		t.Error("There is enough size to add a new line.")
	}
	p.AddLineBytes([]byte("hoge"))
	if p.Length() != 3 {
		t.Error("Length is increased by adding a new line.")
	}
	if p.IsEnoughCapacity() {
		t.Error("There is no enough space to add a new line.")
	}
}

func TestLineBytesAt(t *testing.T) {
	p := newPage(&PageParams{Capacity: 3})
	p.AddLineBytes([]byte("foo"))
	if string(p.LineBytesAt(0)) != "foo" {
		t.Error("The added line is not the same as expected bytes.")
	}
}

func TestReset(t *testing.T) {
	p := newPage(&PageParams{Capacity: 3})
	p.AddLineBytes([]byte("foo"))
	p.Reset()
	if p.Length() != 0 {
		t.Error("Reset doesn't work well.")
	}
	if p.StartLineNumber() != 0 {
		t.Error("Reset doesn't reset startLineNumber.")
	}
}

func TestUpdateNewText(t *testing.T) {
	p := newPage(&PageParams{Capacity: 3})
	p.AddLineBytes([]byte("foo"))
	p.AddLineBytes([]byte("foo"))
	p.AddLineBytes([]byte("foo"))
	p.Reset()
	p.AddLineBytes([]byte("bar"))
	if string(p.LineBytesAt(0)) != "bar" {
		t.Error("Updating a new line didn't work well.")
	}
}

func TestLongTextUpdate(t *testing.T) {
	p := newPage(&PageParams{Capacity: 3})
	textBytes := make([]byte, initialBufferSize*2)
	for i := 0; i < initialBufferSize*2; i++ {
		textBytes[i] = 'a'
	}
	text := string(textBytes)
	p.AddLineBytes(textBytes)
	if string(p.LineBytesAt(0)) != text {
		t.Error("Long text should work well.")
	}
}

func TestNextAndPrevious(t *testing.T) {
	p := newPage(&PageParams{Capacity: 1})
	nextPage := newPage(&PageParams{Capacity: 1})

	p.Next(nextPage)
	if p.nextPage != nextPage || nextPage.previousPage != p {
		t.Error("Next and Previous is not set well.")
	}
}

func TestLineBytesBeforeAndAfter(t *testing.T) {
	prevPage := newPage(&PageParams{Capacity: 3})
	prevPage.AddLineBytes([]byte("bar1"))
	prevPage.AddLineBytes([]byte("bar2"))
	currPage := newPage(&PageParams{Capacity: 3})
	currPage.AddLineBytes([]byte("foo1"))
	currPage.AddLineBytes([]byte("foo2"))
	nextPage := newPage(&PageParams{Capacity: 3})
	nextPage.AddLineBytes([]byte("hoge1"))
	nextPage.AddLineBytes([]byte("hoge2"))

	prevPage.Next(currPage)
	currPage.Next(nextPage)

	lines := prevPage.LineBytesBeforeAndAfter(0, 1, 1)
	if lines[0] != nil {
		t.Error("Cannot get correct line bytes.", string(lines[0]))
	}
	if string(lines[1]) != "bar1" {
		t.Error("Cannot get correct line bytes.", string(lines[1]))
	}
	if string(lines[2]) != "bar2" {
		t.Error("Cannot get correct line bytes.", string(lines[2]))
	}

	lines = currPage.LineBytesBeforeAndAfter(1, 2, 2)
	if string(lines[0]) != "bar2" {
		t.Error("Cannot get correct line bytes.", string(lines[0]))
	}
	if string(lines[1]) != "foo1" {
		t.Error("Cannot get correct line bytes.", string(lines[1]))
	}
	if string(lines[2]) != "foo2" {
		t.Error("Cannot get correct line bytes.", string(lines[2]))
	}
	if string(lines[3]) != "hoge1" {
		t.Error("Cannot get correct line bytes.", string(lines[3]))
	}
	if string(lines[4]) != "hoge2" {
		t.Error("Cannot get correct line bytes.", string(lines[4]))
	}

	lines = nextPage.LineBytesBeforeAndAfter(0, 1, 3)
	if string(lines[0]) != "foo2" {
		t.Error("Cannot get correct line bytes.", string(lines[0]))
	}
	if string(lines[1]) != "hoge1" {
		t.Error("Cannot get correct line bytes.", string(lines[1]))
	}
	if string(lines[2]) != "hoge2" {
		t.Error("Cannot get correct line bytes.", string(lines[2]))
	}
	if lines[3] != nil {
		t.Error("Cannot get correct line bytes.", string(lines[3]))
	}

	currPage.Reset()
	if currPage.nextPage != nil {
		t.Error("Reset should clear nextPage.")
	}
	if currPage.previousPage != nil {
		t.Error("Reset should clear previousPage.")
	}
}

func TestStartLineNumber(t *testing.T) {
	p := newPage(&PageParams{Capacity: 3, StartLineNumber: 4})
	if p.startLineNumber != 4 {
		t.Error("Cannot set StartLineNumber.")
	}
	if p.StartLineNumber() != 4 {
		t.Error("StartLineNumber() doesn't return a correct value.")
	}
}

func TestSetStartLineNumber(t *testing.T) {
	p := newPage(&PageParams{Capacity: 3})
	if p.startLineNumber != 0 {
		t.Error("Initial startLineNumber is incorrect.")
	}
	p.SetStartLineNumber(4)
	if p.startLineNumber != 4 {
		t.Error("SetStartLineNumber is not set correctly.")
	}
}
