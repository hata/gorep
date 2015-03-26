package book

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

type Matcher interface {
	Match(bytes []byte) bool
}

type matcher struct {
	patterns   []string
	ignoreCase bool
}

type regexpMatcher struct {
	matcher
	regexpPatterns []*regexp.Regexp
}

type bytesMatcher struct {
	matcher
	bytesPatterns      [][]byte
	upperBytesPatterns [][]byte
}

func NewRegexpMatcher(patterns []string, ignoreCase bool) Matcher {
	return newRegexpMatcher(patterns, ignoreCase)
}

func NewBytesMatcher(patterns []string, ignoreCase bool) Matcher {
	return newBytesMatcher(patterns, ignoreCase)
}

// TODO: ignoreCase should be supported.
func newRegexpMatcher(patterns []string, ignoreCase bool) (m *regexpMatcher) {
	m = new(regexpMatcher)
	m.patterns = patterns
	m.ignoreCase = ignoreCase

	m.regexpPatterns = make([]*regexp.Regexp, len(m.patterns))
	for i, p := range m.patterns {
		if ignoreCase {
			p = m.toIgnoreRegexpPattern(p)
		}
		m.regexpPatterns[i], _ = regexp.Compile(p)
	}

	return
}

func (m *regexpMatcher) Match(textBytes []byte) bool {
	for _, p := range m.regexpPatterns {
		if p.Match(textBytes) {
			return true
		}
	}

	return false
}

// TODO: This doesn't handle Unicode. So, if double byte characters are used, then it may create
// wrong bytes.
func (m *regexpMatcher) toIgnoreRegexpPattern(pattern string) string {
	var buffer bytes.Buffer
	patternBytes := []byte(pattern)
	lowerBytes := []byte(strings.ToLower(pattern))
	upperBytes := []byte(strings.ToUpper(pattern))
	patternLen := len(patternBytes)

	for i := 0; i < patternLen; i++ {
		if lowerBytes[i] != upperBytes[i] {
			buffer.WriteString(fmt.Sprintf("(%s|%s)", string(lowerBytes[i]), string(upperBytes[i])))
		} else {
			buffer.WriteString(string(patternBytes[i]))
		}
	}

	return buffer.String()
}

func newBytesMatcher(patterns []string, ignoreCase bool) (m *bytesMatcher) {
	m = new(bytesMatcher)

	m.patterns = patterns
	m.ignoreCase = ignoreCase

	m.bytesPatterns = make([][]byte, len(patterns))

	if ignoreCase {
		m.upperBytesPatterns = make([][]byte, len(patterns))
	}

	for i, p := range patterns {
		if ignoreCase {
			m.bytesPatterns[i] = []byte(strings.ToLower(p))
			m.upperBytesPatterns[i] = []byte(strings.ToUpper(p))
		} else {
			m.bytesPatterns[i] = []byte(p)
		}
	}

	return
}

func (m *bytesMatcher) Match(textBytes []byte) bool {
	for i, p := range m.bytesPatterns {
		if len(textBytes) >= len(p) {
			if m.ignoreCase {
				up := m.upperBytesPatterns[i]
				if m.containsIgnoreCase(p, up, textBytes) {
					return true
				}
			} else {
				if bytes.Contains(textBytes, p) {
					return true
				}
			}
		}
	}

	return false
}

func (m *bytesMatcher) containsIgnoreCase(lowerPattern, upperPattern, textBytes []byte) bool {
	matchPos := 0
	textLen := len(textBytes)
	patternLen := len(lowerPattern)

	// TODO: There is a better matcching way.
	for startPos := 0; startPos < textLen; startPos++ {
		if textLen-startPos < patternLen {
			return false
		}
		matchPos = 0

		for i := startPos; i < textLen; i++ {
			b := textBytes[i]
			if b == lowerPattern[matchPos] || b == upperPattern[matchPos] {
				matchPos++
				if matchPos == patternLen {
					return true
				}
			} else {
				break
			}
		}
	}

	return false
}
