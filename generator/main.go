package main

import (
	"os"

	"github.com/pborman/getopt/v2"
)

func main() {
	output := getopt.StringLong(
		"output", 'o', "", "output file to write HTML to.")
	packagePath := getopt.StringLong(
		"package", 'o', "", "package path where *.go files should be written to."+
			" last element of the path determines package name.")
	getopt.Parse()
	args := getopt.Args()

	info, err := os.Stat(*packagePath)
	if err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(*packagePath, os.ModePerm)
		} else {
			panic("unable to access package directory " + *output)
		}
	} else if !info.IsDir() {
		panic("package path is not a directory: " + *output)
	}

	if len(args) == 0 {
		panic("must give at least one input file")
	}

	p := processor{}
	for i := range args {
		p.process(args[i])
	}
	p.dump(*output, *packagePath)
}
