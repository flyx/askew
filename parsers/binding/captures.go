package binding

import "github.com/flyx/askew/data"

// ParseCapture parses the content of an a:capture attribute.
func ParseCapture(s string) ([]data.UnboundEventMapping, error) {
	p := BindingParser{Buffer: s, paramMappings: make(map[string]data.BoundValue),
		eventHandling: data.AutoPreventDefault}
	p.Init()
	if err := p.Parse(int(rulecaptures)); err != nil {
		return nil, err
	}
	p.Execute()
	return p.eventMappings, p.err
}
