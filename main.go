package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pborman/getopt/v2"
)

func main() {
	output := getopt.StringLong(
		"outputDir", 'o', "", "output directory. Each package will be placed as child directory here.")
	outputHTML := getopt.StringLong(
		"outputHtml", 't', "", "path to output file where the HTML templates are written."+
			" defaults to ${outputDir}/templates.html.")
	getopt.Parse()
	args := getopt.Args()

	if strings.Contains(*output, "..") {
		fmt.Fprintf(os.Stderr, "[error] illegal outputDir: %s\n", *output)
		os.Stderr.WriteString("[error] may not contain `..` (must target dir in current module)\n")
	}
	if filepath.IsAbs(*output) {
		fmt.Fprintf(os.Stderr, "[error] illegal outputDir: %s\n", *output)
		os.Stderr.WriteString("[error] must be relative path\n")
	}

	if *outputHTML == "" {
		*outputHTML = filepath.Join(*output, "templates.html")
	}

	info, err := os.Stat(*output)
	if err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(*output, os.ModePerm)
		} else {
			panic("unable to access output directory " + *output)
		}
	} else if !info.IsDir() {
		panic("output path is not a directory: " + *output)
	}

	if len(args) == 0 {
		panic("must give at least one input file")
	}

	var p processor
	if !p.init(*output) {
		os.Exit(1)
	}
	for i := range args {
		p.process(args[i])
	}
	p.dump(*outputHTML, *output)
}
