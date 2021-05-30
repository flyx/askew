package output

import (
	"bytes"
	"go/build"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

var goimportsPath = filepath.Join(build.Default.GOPATH, "bin", "goimports")

func init() {
	info, err := os.Stat(goimportsPath)
	if os.IsNotExist(err) {
		panic("`goimports` missing, please install via `go get golang.org/x/tools/cmd/goimports`")
	}
	if err != nil {
		panic("failed to access `goimports`: " + err.Error())
	}
	if !info.Mode().IsRegular() {
		panic("`goimports` is not a regular file")
	}
}

func writeFormatted(goCode string, file string) {
	fmtcmd := exec.Command(goimportsPath)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	fmtcmd.Stdout = &stdout
	fmtcmd.Stderr = &stderr

	stdin, err := fmtcmd.StdinPipe()
	if err != nil {
		panic("unable to create stdin pipe: " + err.Error())
	}
	err = fmtcmd.Start()
	if err != nil {
		panic("unable to start goimports process for formatting the code: " + err.Error())
	}
	io.WriteString(stdin, goCode)
	stdin.Close()

	if err := fmtcmd.Wait(); err != nil {
		log.Println("error while formatting: " + err.Error())
		log.Println("stderr output:")
		log.Println(stderr.String())
		log.Println("input:")
		log.Println(goCode)
		panic("failed to format Go code")
	}

	if err := ioutil.WriteFile(file, stdout.Bytes(), os.ModePerm); err != nil {
		panic("failed to write file '" + file + "': " + err.Error())
	}
}
