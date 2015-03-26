package book

import (
	"fmt"
	"github.com/hata/goseq"
	"io"
	"runtime"
)

const (
	lineNumberStartAt = LineNumber(1)
)

type PageIndex byte
type FoundHandler func(params *FoundParams)

type FoundCountType uint64

type FindParams struct {
	Patterns            []string
	AfterContextLength  int
	BeforeContextLength int
	IgnoreCase          bool
	FixedStrings        bool
	Handler             FoundHandler
}

type FoundParams struct {
	FindParams
	file          string
	lineNumber    LineNumber
	page          Page
	linePosInPage int
}

type Chapter interface {
	io.Closer
	Find(findParams *FindParams) FoundCountType
}

type chapter struct {
	file          string
	parallelCount int
	executor      goseq.Executor
	afterContext  int
	beforeContext int
	foundHandler  FoundHandler

	pages       []Page
	futures     []goseq.Future
	matchers    []Matcher
	foundParams []FoundParams
	foundCounts []FoundCountType

	newInputFunc func() (Input, error)
}

type chapterStdin struct {
	chapter
}

type chapterFile struct {
	chapter
}

type chapterBytes struct {
	chapter
	bytes []byte
}

func NewChapterStdin() Chapter {
	return newChapterStdin()
}

func NewChapterFile(file string) Chapter {
	return newChapterFile(file)
}

func NewChapterBytes(bytes []byte) Chapter {
	return newChapterBytes(bytes)
}

func DefaultFoundHandler(params *FoundParams) {
	for _, lineBytes := range params.page.LineBytesBeforeAndAfter(
		params.linePosInPage, params.BeforeContextLength, params.AfterContextLength) {
		if lineBytes != nil {
			fmt.Println(string(lineBytes))
		}
	}
}

func LineNumberFoundHandler(params *FoundParams) {
	for _, lineBytes := range params.page.LineBytesBeforeAndAfter(
		params.linePosInPage, params.BeforeContextLength, params.AfterContextLength) {
		if lineBytes != nil {
			fmt.Println(params.lineNumber, ": ", string(lineBytes))
		}
	}
}

func FileNameFoundHandler(params *FoundParams) {
	for _, lineBytes := range params.page.LineBytesBeforeAndAfter(
		params.linePosInPage, params.BeforeContextLength, params.AfterContextLength) {
		if lineBytes != nil {
			fmt.Println(params.file, ": ", string(lineBytes))
		}
	}
}

func FileNameLineNumberFoundHandler(params *FoundParams) {
	for _, lineBytes := range params.page.LineBytesBeforeAndAfter(
		params.linePosInPage, params.BeforeContextLength, params.AfterContextLength) {
		if lineBytes != nil {
			fmt.Println(params.file, ": ", params.lineNumber, ": ", string(lineBytes))
		}
	}
}

func newChapterStdin() (c *chapterStdin) {
	c = new(chapterStdin)
	c.file = ""
	c.parallelCount = runtime.NumCPU() + 2
	c.executor = goseq.NewExecutor(c.parallelCount)
	c.newInputFunc = c.newInput
	c.foundHandler = DefaultFoundHandler
	return
}

func newChapterFile(file string) (c *chapterFile) {
	c = new(chapterFile)
	c.file = file
	c.parallelCount = runtime.NumCPU() + 2
	c.executor = goseq.NewExecutor(c.parallelCount)
	c.newInputFunc = c.newInput
	c.foundHandler = DefaultFoundHandler
	return
}

func newChapterBytes(bytes []byte) (c *chapterBytes) {
	c = new(chapterBytes)
	c.file = ""
	c.parallelCount = runtime.NumCPU() + 2
	c.executor = goseq.NewExecutor(c.parallelCount)
	c.bytes = bytes
	c.newInputFunc = c.newInput
	c.foundHandler = DefaultFoundHandler
	return
}

func (c *chapter) findPageText(pageIndex PageIndex) (count FoundCountType, err error) {
	currentPage := c.pages[pageIndex]
	currentPageLen := currentPage.Length()
	foundHandler := c.foundHandler
	for i := 0; i < currentPageLen; i++ {
		line := currentPage.LineBytesAt(i)
		if c.matchers[pageIndex].Match(line) {
			count++
			if foundHandler != nil {
				c.foundParams[pageIndex].page = currentPage
				c.foundParams[pageIndex].linePosInPage = i
				c.foundParams[pageIndex].lineNumber = currentPage.StartLineNumber() + LineNumber(i) + lineNumberStartAt
				foundHandler(&(c.foundParams[pageIndex]))
			}
		}
	}
	c.foundCounts[pageIndex] += count
	return
}

func (c *chapter) submitPage(pageIndex PageIndex) (oldFuture goseq.Future) {
	if c.futures[pageIndex] != nil {
		oldFuture = c.futures[pageIndex]
		oldFuture.Result()
	}
	c.futures[pageIndex] = c.executor.Execute(func() (goseq.Any, error) {
		return c.findPageText(pageIndex)
	})
	return
}

// TODO: This should be able to break searching files when gorep only need to find a line or not.
func (c *chapter) Find(findParams *FindParams) FoundCountType {
	if findParams == nil {
		return 0
	}

	c.afterContext = findParams.AfterContextLength
	c.beforeContext = findParams.BeforeContextLength
	c.foundHandler = findParams.Handler
	pageBufSize := 1024 * 2
	pageCounter := 0
	pageIndex := PageIndex(0)

	c.pages = make([]Page, c.parallelCount)
	c.futures = make([]goseq.Future, c.parallelCount)
	c.matchers = make([]Matcher, c.parallelCount)
	c.foundCounts = make([]FoundCountType, c.parallelCount)
	c.foundParams = make([]FoundParams, c.parallelCount)

	for i := len(c.pages) - 1; i >= 0; i-- {
		c.pages[i] = NewPage(&PageParams{Capacity: pageBufSize})
		c.pages[i].Reset()
		if findParams.FixedStrings {
			c.matchers[i] = NewBytesMatcher(findParams.Patterns, findParams.IgnoreCase)
		} else {
			c.matchers[i] = NewRegexpMatcher(findParams.Patterns, findParams.IgnoreCase)
		}
		c.foundParams[i].AfterContextLength = findParams.AfterContextLength
		c.foundParams[i].BeforeContextLength = findParams.BeforeContextLength
		c.foundParams[i].FixedStrings = findParams.FixedStrings
		c.foundParams[i].Handler = findParams.Handler
		c.foundParams[i].IgnoreCase = findParams.IgnoreCase
		c.foundParams[i].Patterns = findParams.Patterns
		c.foundParams[i].file = c.file
	}

	input, _ := c.createInput()
	defer input.Close()

	deferredPageIndex := PageIndex(0)
	deferred := false
	lineNum := LineNumber(0)

	input.EachLineBytes(func(line []byte) {
		currentPage := c.pages[pageIndex]
		if !currentPage.IsEnoughCapacity() {
			previousPage := currentPage

			if findParams.AfterContextLength == 0 {
				c.submitPage(pageIndex)
			} else {
				deferred = true
				deferredPageIndex = pageIndex
			}

			pageCounter++
			pageIndex = PageIndex(pageCounter % c.parallelCount)
			currentPage = c.pages[pageIndex]
			currentPage.Reset()
			currentPage.SetStartLineNumber(lineNum)
			previousPage.Next(currentPage)
		}

		currentPage.AddLineBytes(line)
		lineNum++

		if deferred && findParams.AfterContextLength < currentPage.Length() {
			deferred = false
			c.submitPage(deferredPageIndex)
		}
	})

	if deferred {
		c.submitPage(deferredPageIndex)
	}
	// Flush last chapter.
	c.submitPage(pageIndex)

	for _, future := range c.futures {
		if future != nil {
			future.Result()
		}
	}

	var totalFoundCount FoundCountType
	for _, count := range c.foundCounts {
		totalFoundCount += count
	}

	return totalFoundCount
}

func (c *chapter) Close() error {
	c.executor.Stop()
	return nil
}

func (c *chapter) createInput() (Input, error) {
	return c.newInputFunc()
}

func (c *chapterStdin) newInput() (Input, error) {
	return NewStdinInput()
}

func (c *chapterFile) newInput() (Input, error) {
	return NewFileInput(c.file)

}

func (c *chapterBytes) newInput() (Input, error) {
	return NewBytesInput(c.bytes)
}
