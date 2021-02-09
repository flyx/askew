title: The Code Generator
----

The `askew` command line tool takes the following parameters:

    askew [options] [dir]

Possible options are

 * `-o path`, `--outputDir=path`: Specify the directory where the `index.html` file should be placed. Defaults to the current directory.

The `dir` parameter must be a path to a directory containing a Go module or a subdirectory thereof.
If left out, the current directory is used.

## Dependencies

You can reference Askew files in other packages as long as they are in the same module.
Askew is currently unable to depend on Askew source files in a different module.

## Using go generate

You can put a comment like this in any `.go` file in the module's main directory:

```go
//go:generate askew
```

This will run `askew` (which must be available in your PATH) when you issue `go generate` on the command line.
Add options as necessary.
