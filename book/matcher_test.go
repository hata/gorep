package book

import (
	"bytes"
	"testing"
)

func patternsForTest(pattern string, patterns ...string) []string {
	patternArray := make([]string, len(patterns)+1)
	patternArray[0] = pattern
	for i, p := range patterns {
		patternArray[i+1] = p
	}
	return patternArray
}

func TestNewRegexpMatcher(t *testing.T) {
	m := newRegexpMatcher(patternsForTest("foo", "bar"), false)

	if m.patterns[0] != "foo" {
		t.Error("Pattern 'foo' is not set.")
	}
	if m.patterns[1] != "bar" {
		t.Error("Pattern 'bar' is not set.")
	}

	if m.ignoreCase {
		t.Error("ignoreCase is not set.")
	}

	if len(m.regexpPatterns) != 2 {
		t.Error("regexpPatterns has a wrong length.")
	}

	if !m.regexpPatterns[0].MatchString("foo") {
		t.Error("regexPatterns cannot match configured regular expression.")
	}

	if !m.regexpPatterns[1].MatchString("bar") {
		t.Error("regexPatterns cannot match configured regular expression.")
	}
}

func TestRegexpMatch(t *testing.T) {
	m := newRegexpMatcher(patternsForTest("foo.*", "bar.*"), false)
	if !m.Match([]byte("foo")) {
		t.Error("Match for regexp doesn't work.")
	}
	if !m.Match([]byte("ffoo1")) {
		t.Error("Match for regexp doesn't work.")
	}
	if !m.Match([]byte("bar")) {
		t.Error("Match for regexp doesn't work.")
	}
	if !m.Match([]byte("bbar1")) {
		t.Error("Match for regexp doesn't work.")
	}
	if m.Match([]byte("Foo")) {
		t.Error("Match for regexp shouldn't match ignore case.")
	}
}

func TestRegexpMatchIgnoreCase(t *testing.T) {
	m := newRegexpMatcher(patternsForTest("foo.*"), true)
	if !m.Match([]byte("foo")) {
		t.Error("Match for regexp doesn't work.")
	}
	if !m.Match([]byte("ffoo1")) {
		t.Error("Match for regexp doesn't work.")
	}
	if !m.Match([]byte("Foo")) {
		t.Error("Match for regexp shouldn't match ignore case.")
	}
	if !m.Match([]byte("FOO")) {
		t.Error("Match for regexp shouldn't match ignore case.")
	}
}

func TestNewBytesMatcher(t *testing.T) {
	m := newBytesMatcher(patternsForTest("foo", "bar"), false)

	if m.patterns[0] != "foo" {
		t.Error("Pattern 'foo' is not set.")
	}
	if m.patterns[1] != "bar" {
		t.Error("Pattern 'bar' is not set.")
	}

	if m.ignoreCase {
		t.Error("ignoreCase is not set.")
	}

	if len(m.bytesPatterns) != 2 {
		t.Error("bytesPatterns has a wrong length.")
	}

	if !bytes.Equal(m.bytesPatterns[0], []byte("foo")) {
		t.Error("bytesPatterns cannot match configured patterns.")
	}

	if !bytes.Equal(m.bytesPatterns[1], []byte("bar")) {
		t.Error("bytesPatterns cannot match configured patterns.")
	}
}

func TestBytesMatch(t *testing.T) {
	m := newBytesMatcher(patternsForTest("foo", "bar"), false)
	if !m.Match([]byte("foo")) {
		t.Error("Match for bytes doesn't work.")
	}
	if !m.Match([]byte("ffoo1")) {
		t.Error("Match for bytes doesn't work.")
	}
	if !m.Match([]byte("bar")) {
		t.Error("Match for bytes doesn't work.")
	}
	if !m.Match([]byte("bbar1")) {
		t.Error("Match for bytes doesn't work.")
	}
	if m.Match([]byte("Foo")) {
		t.Error("Match for bytes shouldn't match ignore case.")
	}
}

func TestBytesMatchIgnoreCase(t *testing.T) {
	m := newBytesMatcher(patternsForTest("foo"), true)
	if m.Match([]byte("fo")) {
		t.Error("Shorter text doesn't match.")
	}
	if !m.Match([]byte("foo")) {
		t.Error("Match for text doesn't work.")
	}
	if !m.Match([]byte("Foo")) {
		t.Error("Match for text doesn't work.")
	}
	if !m.Match([]byte("FOO")) {
		t.Error("Match for text doesn't work.")
	}
	if !m.Match([]byte("ffoo1")) {
		t.Error("Match for text doesn't work.")
	}
	if !m.Match([]byte("fFoo1")) {
		t.Error("Match for text doesn't work.")
	}
	if !m.Match([]byte("FFOO1")) {
		t.Error("Match for text doesn't work.")
	}
}

func TestBytesMatcherContainsIgnoreCase(t *testing.T) {
	m := newBytesMatcher(patternsForTest("foo"), true)
	if !m.containsIgnoreCase([]byte("foo"), []byte("FOO"), []byte("foo")) {
		t.Error("containsIgnoreCase should match text.")
	}
	if !m.containsIgnoreCase([]byte("foo"), []byte("FOO"), []byte("Foo")) {
		t.Error("containsIgnoreCase should match text.")
	}
	if !m.containsIgnoreCase([]byte("foo"), []byte("FOO"), []byte("FOO")) {
		t.Error("containsIgnoreCase should match text.")
	}
	if !m.containsIgnoreCase([]byte("foo"), []byte("FOO"), []byte("xFoOy")) {
		t.Error("containsIgnoreCase should match text.")
	}
	if !m.containsIgnoreCase([]byte("foo"), []byte("FOO"), []byte("FoFoOy")) {
		t.Error("containsIgnoreCase should match text.")
	}
	if !m.containsIgnoreCase([]byte("foo"), []byte("FOO"), []byte("FoFoO")) {
		t.Error("containsIgnoreCase should match text.")
	}

	if m.containsIgnoreCase([]byte("foo"), []byte("FOO"), []byte("")) {
		t.Error("containsIgnoreCase should not match text.")
	}
	if m.containsIgnoreCase([]byte("foo"), []byte("FOO"), []byte("F")) {
		t.Error("containsIgnoreCase should not match text.")
	}
	if m.containsIgnoreCase([]byte("foo"), []byte("FOO"), []byte("Fo")) {
		t.Error("containsIgnoreCase should not match text.")
	}
	if m.containsIgnoreCase([]byte("foo"), []byte("FOO"), []byte("FoF")) {
		t.Error("containsIgnoreCase should not match text.")
	}
	if m.containsIgnoreCase([]byte("foo"), []byte("FOO"), []byte("FoFo")) {
		t.Error("containsIgnoreCase should not match text.")
	}
}
