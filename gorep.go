package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/hata/gorep/book"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
)

const (
	AppName         = "gorep"
	AppUsage        = "Experimental implementation of a subset of grep command"
	AppAuthor       = "hata"
	appHelpTemplate = `
NAME:
   {{.Name}} - {{.Usage}}
USAGE:
   {{.Name}} {{if .Flags}}[options] {{end}}[pattern] [files...]
VERSION:
   {{.Version}}
AUTHOR(S): 
   {{.Author}}{{if .Flags}}
OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}{{end}}
	`
)

func main() {
	cpus := runtime.NumCPU()
	runtime.GOMAXPROCS(cpus)

	cli.AppHelpTemplate = appHelpTemplate

	app := cli.NewApp()
	app.Name = AppName
	app.Version = Version
	app.Usage = AppUsage
	app.Author = AppAuthor
	app.Email = ""
	app.Flags = Flags
	app.Action = flagAction

	app.Run(os.Args)

	var handler book.FoundHandler
	totalCount := book.FoundCountType(0)

	if len(appOptions.patterns) == 0 {
		os.Exit(1)
		return
	}

	if !appOptions.fixedStrings {
		for _, p := range appOptions.patterns {
			_, err := regexp.Compile(p)
			if err != nil {
				fmt.Println("Failed to compile regexp.", err)
				return
			}
		}
	}

	if appOptions.count || appOptions.quiet || appOptions.filesWithMatches || appOptions.filesWithoutMatch {
		handler = nil
	} else if appOptions.lineNumber && appOptions.showFilenameFlag {
		handler = book.FileNameLineNumberFoundHandler
	} else if appOptions.lineNumber {
		handler = book.LineNumberFoundHandler
	} else if appOptions.showFilenameFlag {
		handler = book.FileNameFoundHandler
	} else {
		handler = book.DefaultFoundHandler
	}

	for _, aFile := range appOptions.files {
		totalCount += parsePath(aFile, appOptions, handler)
	}

	if appOptions.count && !appOptions.quiet {
		fmt.Println(uint64(totalCount))
	}

	if totalCount > 0 {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}

func parsePath(rootPath string, appOptions AppOptions, handler book.FoundHandler) (count book.FoundCountType) {
	fstat, ferr := os.Stat(rootPath)
	if ferr != nil {
		return
	}

	if !appOptions.recursive && fstat.IsDir() {
		fmt.Println(AppName, ": ", rootPath, ": Is a directory")
		return
	}

	filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			if !os.SameFile(fstat, info) {
				count += parsePath(path, appOptions, handler)
			}
		} else {
			n := parseFile(path, appOptions, handler)

			if (n > 0 && appOptions.filesWithMatches) || (n == 0 && appOptions.filesWithoutMatch) {
				fmt.Println(path)
			}

			count += n
		}
		return nil
	})
	return
}

func parseFile(file string, appOptions AppOptions, handler book.FoundHandler) book.FoundCountType {
	c := book.NewChapterFile(file)
	defer c.Close()

	return c.Find(&book.FindParams{
		Patterns:            appOptions.patterns,
		AfterContextLength:  appOptions.afterContext,
		BeforeContextLength: appOptions.beforeContext,
		IgnoreCase:          appOptions.ignoreCase,
		FixedStrings:        appOptions.fixedStrings,
		Handler:             handler,
	})
}
