package main

import (
	"github.com/codegangsta/cli"
	"github.com/hata/gorep/book"
	"log"
	"os"
)

type AppOptions struct {
	afterContext      int
	beforeContext     int
	context           int
	count             bool
	fixedStrings      bool
	file              string
	withFilename      bool
	noFilename        bool
	ignoreCase        bool
	filesWithoutMatch bool
	filesWithMatches  bool
	lineNumber        bool
	quiet             bool
	recursive         bool

	// Created from command line options.
	patterns         []string
	files            []string
	showFilenameFlag bool
}

var appOptions AppOptions

var Flags = []cli.Flag{
	flagAfterContext,
	flagBeforeContext,
	flagContext,
	flagCount,
	flagFixedStrings,
	flagFile,
	flagWithFilename,
	flagNoFilename,
	//    flagHelp,
	flagIgnoreCase,
	flagFilesWithoutMatch,
	flagFilesWithMatches,
	flagLineNumber,
	flagQuiet,
	flagRecursive,
	//    flagVersion,
}

var flagAfterContext = cli.IntFlag{
	Name:  "after-context, A",
	Usage: "Print num lines of trailing context after each match. See also the -B and -C options",
}

var flagBeforeContext = cli.IntFlag{
	Name:  "before-context, B",
	Usage: "Print num lines of leading context before each match. See also the -A and -C options.",
}

var flagContext = cli.IntFlag{
	Name:  "context, C",
	Usage: "Print num lines of leading and trailing context surrounding each match.",
}

var flagCount = cli.BoolFlag{
	Name:  "count, c",
	Usage: "Only a count of selected lines is written to standard output.",
}

var flagFixedStrings = cli.BoolFlag{
	Name:  "fixed-strings, F",
	Usage: "Interpret pattern as a set of fixed strings",
}

var flagFile = cli.StringFlag{
	Name:  "file, f",
	Usage: "Read one or more newline separated patterns from file.",
}

var flagWithFilename = cli.BoolFlag{
	Name:  "with-filename, H",
	Usage: "Always print filename headers with output lines",
}

var flagNoFilename = cli.BoolFlag{
	Name:  "no-filename",
	Usage: "Never print filename headers (i.e. filenames) with output lines.",
}

/*
var flagHelp = cli.Flag{
	Name:  "help",
	Usage: "",
}*/

var flagIgnoreCase = cli.BoolFlag{
	Name:  "ignore-case, i",
	Usage: "Perform case insensitive matching.  By default, grep is case sensitive.",
}

var flagFilesWithoutMatch = cli.BoolFlag{
	Name:  "files-without-match, L",
	Usage: "Only the names of files not containing selected lines are written to standard output.",
}

var flagFilesWithMatches = cli.BoolFlag{
	Name:  "files-with-matches, l",
	Usage: "Only the names of files containing selected lines are written to standard output.",
}

var flagLineNumber = cli.BoolFlag{
	Name:  "line-number, n",
	Usage: "Each output line is preceded by its relative line number in the file, starting at line 1.",
}

var flagQuiet = cli.BoolFlag{
	Name:  "quiet, q",
	Usage: "Quiet mode: suppress normal output.",
}

var flagRecursive = cli.BoolFlag{
	Name:  "recursive, r",
	Usage: "Recursively search subdirectories listed.",
}

/*
var flagVersion = cli.Flag{
	Name:  "version",
	Usage: "",
}*/

func flagAction(c *cli.Context) {
	appOptions.afterContext = c.Int("after-context")
	appOptions.beforeContext = c.Int("before-context")
	appOptions.context = c.Int("context")
	appOptions.count = c.Bool("count")
	appOptions.fixedStrings = c.Bool("fixed-strings")
	appOptions.file = c.String("file")
	appOptions.withFilename = c.Bool("with-filename")
	appOptions.noFilename = c.Bool("no-filename")
	appOptions.ignoreCase = c.Bool("ignore-case")
	appOptions.filesWithoutMatch = c.Bool("files-without-match")
	appOptions.filesWithMatches = c.Bool("files-with-matches")
	appOptions.lineNumber = c.Bool("line-number")
	appOptions.quiet = c.Bool("quiet")
	appOptions.recursive = c.Bool("recursive")

	if len(appOptions.file) > 0 {
		appOptions.patterns = readPatternsFile(appOptions.file)
	}

	length := len(c.Args())
	if length > 0 {
		startFileIndex := 0
		if appOptions.patterns == nil {
			appOptions.patterns = make([]string, 1)
			appOptions.patterns[0] = c.Args()[0]
			startFileIndex++
		}
		if length > startFileIndex {
			appOptions.files = make([]string, length-startFileIndex)
			for i := startFileIndex; i < length; i++ {
				appOptions.files[i-startFileIndex] = c.Args()[i]
				dirFlag, dirErr := isDir(appOptions.files[i-startFileIndex])
				if dirFlag && dirErr == nil {
					appOptions.showFilenameFlag = true
				}
			}
		} else {
			appOptions.files = nil
		}
	} else {
		cli.ShowAppHelp(c)
		return
	}

	if appOptions.context > 0 && appOptions.afterContext == 0 && appOptions.beforeContext == 0 {
		appOptions.afterContext = appOptions.context
		appOptions.beforeContext = appOptions.context
	}

	if !appOptions.showFilenameFlag {
		appOptions.showFilenameFlag = appOptions.withFilename
		if !appOptions.showFilenameFlag && appOptions.noFilename {
			appOptions.showFilenameFlag = false
		} else {
			if len(appOptions.files) > 1 || appOptions.recursive {
				appOptions.showFilenameFlag = true
			} else { // TODO: If there is a directory, then showFilenameFlag should be true if my understanding is correct.
				appOptions.showFilenameFlag = false
			}
		}
	}
}

func debug(v ...interface{}) {
	if os.Getenv("DEBUG") != "" {
		log.Println(v...)
	}
}

func assert(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func isDir(path string) (bool, error) {
	fstat, err := os.Stat(path)
	if err != nil {
		return false, err
	} else {
		return fstat.IsDir(), nil
	}
}

func readPatternsFile(path string) []string {
	patterns := make([]string, 0, 10)

	input, err := book.NewFileInput(path)
	defer input.Close()

	if err != nil {
		return nil
	}

	input.EachLineBytes(func(text []byte) {
		patterns = append(patterns, string(text))
	})

	return patterns
}
