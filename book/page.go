package book

import (
//	"io"
//	"bytes"
)

const (
	initialBufferSize = 128
)

type LineBytes []byte
type LineNumber uint64

type PageParams struct {
	Capacity        int
	StartLineNumber LineNumber
}

// Manage some lines.
type page struct {
	// Default array for pool slices.
	mainPool []byte

	// This is pooled buffer for slice referenced by lines
	pools []LineBytes
	lines [][]byte
	// This is filled length of lines
	length int
	// This is the same as len(pools)
	capacity int

	previousPage Page
	nextPage     Page

	startLineNumber LineNumber
}

type Page interface {
	AddLineBytes(line LineBytes)
	IsEnoughCapacity() bool
	Reset()
	Length() int
	Capacity() int
	LineBytesAt(index int) LineBytes

	previous(previousPage Page)
	Next(nextPage Page)

	LineBytesBeforeAndAfter(index, beforeLength, afterLength int) []LineBytes
	StartLineNumber() LineNumber
	SetStartLineNumber(newLineNumber LineNumber)
}

// e.g.
// page := NewPage(&PageParams{Capacity:1024, StartLineNumber:100})
func NewPage(params *PageParams) Page {
	page := newPage(params)
	return page
}

func newPage(params *PageParams) *page {
	p := new(page)

	p.mainPool = make([]byte, params.Capacity*initialBufferSize)
	p.pools = make([]LineBytes, params.Capacity)
	p.lines = make([][]byte, params.Capacity)

	first := 0
	for i := 0; i < params.Capacity; i++ {
		p.pools[i] = p.mainPool[first : first+initialBufferSize]
		first += initialBufferSize
	}

	p.length = 0
	p.capacity = params.Capacity
	p.startLineNumber = params.StartLineNumber
	return p
}

func (p *page) AddLineBytes(line LineBytes) {
	ln := len(line)
	if cap(p.pools[p.length]) < ln {
		p.pools[p.length] = make(LineBytes, ln<<1) // TODO: Check this incremented size is ok ??
	}
	p.lines[p.length] = p.pools[p.length][0:ln]
	copy(p.lines[p.length], line)
	p.length++
}

func (p *page) IsEnoughCapacity() bool {
	return p.length < p.capacity
}

func (p *page) Reset() {
	p.length = 0
	p.nextPage = nil
	p.previousPage = nil
	p.startLineNumber = 0
}

func (p *page) Length() int {
	return p.length
}

func (p *page) Capacity() int {
	return p.capacity
}

func (p *page) LineBytesAt(index int) LineBytes {
	return p.lines[index]
}

func (p *page) previous(previousPage Page) {
	p.previousPage = previousPage
}

func (p *page) Next(nextPage Page) {
	p.nextPage = nextPage
	nextPage.previous(p)
}

func (p *page) LineBytesBeforeAndAfter(index, beforeLength, afterLength int) []LineBytes {
	paragraph := make([]LineBytes, beforeLength+afterLength+1)

	paragraph[beforeLength] = p.LineBytesAt(index)

	for i := 1; i <= beforeLength; i++ {
		if index-i >= 0 {
			paragraph[beforeLength-i] = p.LineBytesAt(index - i)
		} else if p.previousPage != nil {
			n := p.previousPage.Length() + (index - i)
			paragraph[beforeLength-i] = p.previousPage.LineBytesAt(n)
		}
	}

	for i := 1; i <= afterLength; i++ {
		if index+i < p.length {
			paragraph[beforeLength+i] = p.LineBytesAt(index + i)
		} else if p.nextPage != nil {
			// How we can check nextPage is valid or not ?
			// This will be handled by a caller(main) side.
			paragraph[beforeLength+i] = p.nextPage.LineBytesAt((index + i) - p.length)
		}
	}

	return paragraph
}

func (p *page) StartLineNumber() LineNumber {
	return p.startLineNumber
}

func (p *page) SetStartLineNumber(newLineNumber LineNumber) {
	p.startLineNumber = newLineNumber
}
