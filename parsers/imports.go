package parsers

// ParseImports parses the content of an a:import element.
func ParseImports(s string) (map[string]string, error) {
	p := GeneralParser{Buffer: s, imports: make(map[string]string)}
	p.Init()
	if err := p.Parse(int(ruleimports)); err != nil {
		return nil, err
	}
	p.Execute()
	return p.imports, p.err
}
