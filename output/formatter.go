package output

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

func writeFormatted(goCode string, file string) {
	fmtcmd := exec.Command("goimports")

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
