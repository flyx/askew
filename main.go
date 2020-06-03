package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flyx/tbc/data"

	"github.com/pborman/getopt/v2"
)

func main() {
	skeleton := getopt.StringLong("skeleton", 's', "", "Skeleton HTML file. If given, will be the "+
		"base for the output HTML.")
	output := getopt.StringLong(
		"outputDir", 'o', "", "output directory. Each package will be placed as child directory here.")
	outputHTML := getopt.StringLong(
		"outputHtml", 'i', "", "path to the HTML output file. This file will contain the rendered skeleton "+
			"in one has been given, or a list of templates if not. defaults to ${outputDir}/<name>.html where "+
			"<name> is \"templates\" if no skeleton is given and ${outputDir}/index.html if a skeleton is given.")
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
		if *skeleton == "" {
			*outputHTML = filepath.Join(*output, "templates.html")
		} else {
			*outputHTML = filepath.Join(*output, "index.html")
		}
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
	var s *data.Skeleton
	if *skeleton != "" {
		s, err = readSkeleton(&p.syms, *skeleton)
		if err != nil {
			panic(err)
		}
	}
	p.dump(s, *outputHTML, *output)
}
