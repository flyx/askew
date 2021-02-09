package parsers

import "github.com/flyx/askew/data"

// HandlerSpec describes the content of an <a:handler> node.
type HandlerSpec struct {
	Name    string
	Params  []data.Param
	Returns *data.ParamType
}

// ParseHandlers parses the content of a <a:handlers> element.
func ParseHandlers(s string) ([]HandlerSpec, error) {
	p := GeneralParser{Buffer: s}
	p.Init()
	if err := p.Parse(int(rulehandlers)); err != nil {
		return nil, err
	}
	p.Execute()
	return p.handlers, p.err
}
