package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/flyx/askew/packages"

	"github.com/pborman/getopt/v2"
)

func main() {
	output := getopt.StringLong(
		"outputDir", 'o', ".", "output directory for index.html")
	excludes := getopt.ListLong("exclude", 'e',
		"comma-separated list of directories to exclude. "+
			"allows patterns (which must be quoted in a typical shell). "+
			"relative to the directory given at command line, or to cwd if no directory is given.")
	getopt.Parse()
	var err error
	outputDirPath, err := filepath.Abs(*output)
	if err != nil {
		panic(err)
	}

	args := getopt.Args()
	if len(args) == 1 {
		if err := os.Chdir(args[0]); err != nil {
			os.Stdout.WriteString("[error] cannot process directory: " + err.Error() + "\n")
			os.Exit(1)
		}
	} else if len(args) > 0 {
		os.Stdout.WriteString("[error] unexpected arguments:\n")
		for i := 1; i < len(args); i++ {
			os.Stdout.WriteString("[error]   " + args[i] + "\n")
		}
		os.Exit(1)
	}

	info, err := os.Stat(*output)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(*output, os.ModePerm)
			if err != nil {
				panic("unable to create output directory " + *output)
			}
		} else {
			panic("unable to access output directory " + *output)
		}
	} else if !info.IsDir() {
		panic("output path is not a directory: " + *output)
	}

	base, err := packages.Discover(*excludes)
	if err != nil {
		os.Stdout.WriteString("[error] " + err.Error() + "\n")
		os.Exit(1)
	}
	order, err := packages.Sort(base.ImportPath, base.Packages)
	if err != nil {
		os.Stdout.WriteString("[error] " + err.Error() + "\n")
		os.Exit(1)
	}

	fmt.Printf("importPath = %s\n", base.ImportPath)
	for pRelPath, pkg := range base.Packages {
		fmt.Printf("package %s at relpath '%s'\n", pkg.Name, pRelPath)
	}

	var p processor
	p.init(base)
	for _, path := range order {
		if err := p.processMacros(path); err != nil {
			os.Stdout.WriteString("[error] " + err.Error() + "\n")
			os.Exit(1)
		}
	}
	for _, path := range order {
		if err := p.processComponents(path); err != nil {
			os.Stdout.WriteString("[error] " + err.Error() + "\n")
			os.Exit(1)
		}
	}

	os.Stdout.WriteString("[info] generating code\n")
	if err := p.dump(outputDirPath); err != nil {
		os.Stdout.WriteString("[error] " + err.Error() + "\n")
		os.Exit(1)
	}
}
