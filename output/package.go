package output

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/flyx/tbc/data"
)

// PackageWriter writes the Go code for a component into a file.
type PackageWriter struct {
	Syms        *data.Symbols
	PackageName string
	PackagePath string
}

func (pw *PackageWriter) writeFormatted(goCode string, file string) {
	fmtcmd := exec.Command("gofmt")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	fmtcmd.Stdout = &stdout
	fmtcmd.Stderr = &stderr

	stdin, err := fmtcmd.StdinPipe()
	if err != nil {
		panic("unable to create stdin pipe: " + err.Error())
	}
	io.WriteString(stdin, goCode)
	stdin.Close()

	if err := fmtcmd.Run(); err != nil {
		log.Println("error while formatting: " + err.Error())
		log.Println("stderr output:")
		log.Println(stderr.String())
		log.Println("input:")
		log.Println(goCode)
		panic("failed to format Go code")
	}

	if err := ioutil.WriteFile(file, []byte(stdout.String()), os.ModePerm); err != nil {
		panic("failed to write file '" + file + "': " + err.Error())
	}
}

func wrapperForType(k data.VariableType) string {
	switch k {
	case data.StringVar:
		return "StringValue"
	case data.IntVar:
		return "IntValue"
	case data.BoolVar:
		return "BoolValue"
	default:
		panic("unsupported type")
	}
}

// WriteComponent writes a component of the package to a file.
func (pw *PackageWriter) WriteComponent(name string, c *data.Component) {
	b := strings.Builder{}
	fileHeader.Execute(&b, struct {
		PackageName string
		Deps        map[string]struct{}
	}{pw.PackageName, c.Dependencies})
	if c.NeedsList && c.Handlers != nil {
		componentController.Execute(&b, struct {
			Name     string
			Handlers map[string]data.Handler
		}{Name: name, Handlers: c.Handlers})
	}

	if err := component.Execute(&b, c); err != nil {
		panic(err)
	}

	if c.Handlers != nil {
		if c.NeedsController {
			b.WriteString("// SetController defines which object handles the captured events\n")
			b.WriteString("// of this component. If set to nil, default behavior will take over.\n")
			fmt.Fprintf(&b, "func (o *%s) SetController(c %sController) {\n", name, name)
			b.WriteString("o.c = c\n}\n")
		}
		for hName, h := range c.Handlers {
			fmt.Fprintf(&b, "func (o *%s) call%s(", name, hName)
			first := true
			for pName := range h.Params {
				if first {
					first = false
				} else {
					b.WriteString(", ")
				}
				fmt.Fprintf(&b, "%s runtime.BoundValue", pName)
			}
			b.WriteString(") bool {\n")
			if c.NeedsController {
				b.WriteString("if o.c == nil {\nreturn false\n}\n")
			}
			for pName, pType := range h.Params {
				fmt.Fprintf(&b, "_%s := runtime.%s{BoundValue: %s}\n",
					pName, wrapperForType(pType), pName)
			}
			if c.NeedsController {
				fmt.Fprintf(&b, "return o.c.%s(", hName)
			} else {
				fmt.Fprintf(&b, "return o.%s(", hName)
			}
			first = true
			for pName := range h.Params {
				if first {
					first = false
				} else {
					b.WriteString(", ")
				}
				fmt.Fprintf(&b, "_%s.Get()", pName)
			}
			b.WriteString(")\n}\n")
		}
	}

	if c.NeedsList {
		if err := componentList.Execute(&b, name); err != nil {
			panic(err)
		}
	}

	pw.writeFormatted(b.String(), filepath.Join(pw.PackagePath, strings.ToLower(name)+".go"))
}
