package book

import (
	"testing"
)

func TestNewChapterFile(t *testing.T) {
	c := newChapterFile("filename")
	if c.file != "filename" {
		t.Error("file field is not set.")
	}
}

func TestFind(t *testing.T) {
	c := newChapterBytes([]byte("foo\nbar\n"))
	c.Find(&FindParams{
		Patterns:            []string{"foo"},
		AfterContextLength:  0,
		BeforeContextLength: 0,
		IgnoreCase:          false,
		FixedStrings:        true,
	})

}
