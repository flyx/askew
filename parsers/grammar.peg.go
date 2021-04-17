package parsers

import (
	"errors"
	"strings"
	"github.com/flyx/askew/data"
	"fmt"
	"math"
	"sort"
	"strconv"
)

const endSymbol rune = 1114112

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	rulee
	ruleassignments
	rulebindings
	rulebinding
	ruleautovar
	ruletypedvar
	ruleisp
	ruleassignment
	rulebound
	ruleself
	ruledataset
	ruleprop
	rulestyle
	ruleclass
	ruleform
	rulegoExpr
	ruleevent
	rulehtmlid
	rulejsid
	ruleexpr
	rulecommaless
	rulenumber
	ruleoperators
	rulestring
	ruleenclosed
	ruleparens
	rulebraces
	rulebrackets
	ruleinner
	ruleidentifier
	rulefields
	rulefsep
	rulefield
	rulename
	ruletype
	rulesname
	ruleqname
	rulearray
	rulemap
	rulechan
	rulefunc
	rulekeytype
	rulepointer
	rulecaptures
	rulecapture
	rulehandlername
	ruleeventid
	rulemappings
	rulemappingstart
	rulemapping
	rulemappingname
	ruletags
	ruletag
	ruletagname
	ruletagarg
	rulefor
	ruleforVar
	rulehandlers
	rulehandler
	ruleparamname
	ruleparam
	rulecparams
	rulecparam
	rulevar
	ruleargs
	rulearg
	ruleimports
	ruleimport
	ruleAction0
	rulePegText
	ruleAction1
	ruleAction2
	ruleAction3
	ruleAction4
	ruleAction5
	ruleAction6
	ruleAction7
	ruleAction8
	ruleAction9
	ruleAction10
	ruleAction11
	ruleAction12
	ruleAction13
	ruleAction14
	ruleAction15
	ruleAction16
	ruleAction17
	ruleAction18
	ruleAction19
	ruleAction20
	ruleAction21
	ruleAction22
	ruleAction23
	ruleAction24
	ruleAction25
	ruleAction26
	ruleAction27
	ruleAction28
	ruleAction29
	ruleAction30
	ruleAction31
	ruleAction32
	ruleAction33
	ruleAction34
	ruleAction35
	ruleAction36
	ruleAction37
	ruleAction38
	ruleAction39
	ruleAction40
	ruleAction41

	rulePre
	ruleIn
	ruleSuf
)

var rul3s = [...]string{
	"Unknown",
	"e",
	"assignments",
	"bindings",
	"binding",
	"autovar",
	"typedvar",
	"isp",
	"assignment",
	"bound",
	"self",
	"dataset",
	"prop",
	"style",
	"class",
	"form",
	"goExpr",
	"event",
	"htmlid",
	"jsid",
	"expr",
	"commaless",
	"number",
	"operators",
	"string",
	"enclosed",
	"parens",
	"braces",
	"brackets",
	"inner",
	"identifier",
	"fields",
	"fsep",
	"field",
	"name",
	"type",
	"sname",
	"qname",
	"array",
	"map",
	"chan",
	"func",
	"keytype",
	"pointer",
	"captures",
	"capture",
	"handlername",
	"eventid",
	"mappings",
	"mappingstart",
	"mapping",
	"mappingname",
	"tags",
	"tag",
	"tagname",
	"tagarg",
	"for",
	"forVar",
	"handlers",
	"handler",
	"paramname",
	"param",
	"cparams",
	"cparam",
	"var",
	"args",
	"arg",
	"imports",
	"import",
	"Action0",
	"PegText",
	"Action1",
	"Action2",
	"Action3",
	"Action4",
	"Action5",
	"Action6",
	"Action7",
	"Action8",
	"Action9",
	"Action10",
	"Action11",
	"Action12",
	"Action13",
	"Action14",
	"Action15",
	"Action16",
	"Action17",
	"Action18",
	"Action19",
	"Action20",
	"Action21",
	"Action22",
	"Action23",
	"Action24",
	"Action25",
	"Action26",
	"Action27",
	"Action28",
	"Action29",
	"Action30",
	"Action31",
	"Action32",
	"Action33",
	"Action34",
	"Action35",
	"Action36",
	"Action37",
	"Action38",
	"Action39",
	"Action40",
	"Action41",

	"Pre_",
	"_In_",
	"_Suf",
}

type node32 struct {
	token32
	up, next *node32
}

func (node *node32) print(depth int, buffer string) {
	for node != nil {
		for c := 0; c < depth; c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[node.pegRule], strconv.Quote(string(([]rune(buffer)[node.begin:node.end]))))
		if node.up != nil {
			node.up.print(depth+1, buffer)
		}
		node = node.next
	}
}

func (node *node32) Print(buffer string) {
	node.print(0, buffer)
}

type element struct {
	node *node32
	down *element
}

/* ${@} bit structure for abstract syntax tree */
type token32 struct {
	pegRule
	begin, end, next uint32
}

func (t *token32) isZero() bool {
	return t.pegRule == ruleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token32) isParentOf(u token32) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token32) getToken32() token32 {
	return token32{pegRule: t.pegRule, begin: uint32(t.begin), end: uint32(t.end), next: uint32(t.next)}
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", rul3s[t.pegRule], t.begin, t.end, t.next)
}

type tokens32 struct {
	tree    []token32
	ordered [][]token32
}

func (t *tokens32) trim(length int) {
	t.tree = t.tree[0:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) Order() [][]token32 {
	if t.ordered != nil {
		return t.ordered
	}

	depths := make([]int32, 1, math.MaxInt16)
	for i, token := range t.tree {
		if token.pegRule == ruleUnknown {
			t.tree = t.tree[:i]
			break
		}
		depth := int(token.next)
		if length := len(depths); depth >= length {
			depths = depths[:depth+1]
		}
		depths[depth]++
	}
	depths = append(depths, 0)

	ordered, pool := make([][]token32, len(depths)), make([]token32, len(t.tree)+len(depths))
	for i, depth := range depths {
		depth++
		ordered[i], pool, depths[i] = pool[:depth], pool[depth:], 0
	}

	for i, token := range t.tree {
		depth := token.next
		token.next = uint32(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type state32 struct {
	token32
	depths []int32
	leaf   bool
}

func (t *tokens32) AST() *node32 {
	tokens := t.Tokens()
	stack := &element{node: &node32{token32: <-tokens}}
	for token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	return stack.node
}

func (t *tokens32) PreOrder() (<-chan state32, [][]token32) {
	s, ordered := make(chan state32, 6), t.Order()
	go func() {
		var states [8]state32
		for i := range states {
			states[i].depths = make([]int32, len(ordered))
		}
		depths, state, depth := make([]int32, len(ordered)), 0, 1
		write := func(t token32, leaf bool) {
			S := states[state]
			state, S.pegRule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.pegRule, t.begin, t.end, uint32(depth), leaf
			copy(S.depths, depths)
			s <- S
		}

		states[state].token32 = ordered[0][0]
		depths[0]++
		state++
		a, b := ordered[depth-1][depths[depth-1]-1], ordered[depth][depths[depth]]
	depthFirstSearch:
		for {
			for {
				if i := depths[depth]; i > 0 {
					if c, j := ordered[depth][i-1], depths[depth-1]; a.isParentOf(c) &&
						(j < 2 || !ordered[depth-1][j-2].isParentOf(c)) {
						if c.end != b.begin {
							write(token32{pegRule: ruleIn, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token32{pegRule: rulePre, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.pegRule != ruleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.pegRule != ruleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token32{pegRule: ruleSuf, begin: b.end, end: a.end}, true)
				}

				depth--
				if depth > 0 {
					a, b, c = ordered[depth-1][depths[depth-1]-1], a, ordered[depth][depths[depth]]
					parent = a.isParentOf(b)
					continue
				}

				break depthFirstSearch
			}
		}

		close(s)
	}()
	return s, ordered
}

func (t *tokens32) PrintSyntax() {
	tokens, ordered := t.PreOrder()
	max := -1
	for token := range tokens {
		if !token.leaf {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[36m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", rul3s[token.pegRule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", rul3s[token.pegRule])
		} else {
			for c, end := token.begin, token.end; c < end; c++ {
				if i := int(c); max+1 < i {
					for j := max; j < i; j++ {
						fmt.Printf("skip %v %v\n", j, token.String())
					}
					max = i
				} else if i := int(c); i <= max {
					for j := i; j <= max; j++ {
						fmt.Printf("dupe %v %v\n", j, token.String())
					}
				} else {
					max = int(c)
				}
				fmt.Printf("%v", c)
				for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
					fmt.Printf(" \x1B[34m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", rul3s[token.pegRule])
			}
			fmt.Printf("\n")
		}
	}
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	tokens, _ := t.PreOrder()
	for token := range tokens {
		for c := 0; c < int(token.next); c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[token.pegRule], strconv.Quote(string(([]rune(buffer)[token.begin:token.end]))))
	}
}

func (t *tokens32) Add(rule pegRule, begin, end, depth uint32, index int) {
	t.tree[index] = token32{pegRule: rule, begin: uint32(begin), end: uint32(end), next: uint32(depth)}
}

func (t *tokens32) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.getToken32()
		}
		close(s)
	}()
	return s
}

func (t *tokens32) Error() []token32 {
	ordered := t.Order()
	length := len(ordered)
	tokens, length := make([]token32, length), length-1
	for i := range tokens {
		o := ordered[length-i]
		if len(o) > 1 {
			tokens[i] = o[len(o)-2].getToken32()
		}
	}
	return tokens
}

func (t *tokens32) Expand(index int) {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
}

type GeneralParser struct {
	eventHandling                         data.EventHandling
	expr, tagname, handlername, eventName string
	paramnames                            []string
	names                                 []string
	keytype, valuetype                    *data.ParamType
	fields                                []*data.Field
	bv                                    data.BoundValue
	goVal                                 data.GoValue
	paramMappings                         map[string]data.BoundValue
	paramIndex                            int
	params                                []data.Param
	isVar                                 bool
	err                                   error

	assignments   []data.Assignment
	varMappings   []data.VariableMapping
	eventMappings []data.UnboundEventMapping
	handlers      []HandlerSpec
	cParams       []data.ComponentParam
	imports       map[string]string

	Buffer string
	buffer []rune
	rules  [112]func() bool
	Parse  func(rule ...int) error
	Reset  func()
	Pretty bool
	tokens32
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer []rune, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer {
		if c == '\n' {
			line, symbol = line+1, 0
		} else {
			symbol++
		}
		if i == positions[j] {
			translations[positions[j]] = textPosition{line, symbol}
			for j++; j < length; j++ {
				if i != positions[j] {
					continue search
				}
			}
			break search
		}
	}

	return translations
}

type parseError struct {
	p   *GeneralParser
	max token32
}

func (e *parseError) Error() string {
	tokens, error := []token32{e.max}, "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.buffer, positions)
	format := "parse error near %v (line %v symbol %v - line %v symbol %v):\n%v\n"
	if e.p.Pretty {
		format = "parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n"
	}
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf(format,
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			strconv.Quote(string(e.p.buffer[begin:end])))
	}

	return error
}

func (p *GeneralParser) PrintSyntaxTree() {
	p.tokens32.PrintSyntaxTree(p.Buffer)
}

func (p *GeneralParser) Highlighter() {
	p.PrintSyntax()
}

func (p *GeneralParser) Execute() {
	buffer, _buffer, text, begin, end := p.Buffer, p.buffer, "", 0, 0
	for token := range p.Tokens() {
		switch token.pegRule {

		case rulePegText:
			begin, end = int(token.begin), int(token.end)
			text = string(_buffer[begin:end])

		case ruleAction0:

			p.varMappings = append(p.varMappings,
				data.VariableMapping{Value: p.bv, Variable: p.goVal})
			p.goVal.Type = nil
			p.bv.IDs = nil

		case ruleAction1:

			p.goVal.Name = buffer[begin:end]

		case ruleAction2:

			p.goVal.Type = p.valuetype
			p.valuetype = nil

		case ruleAction3:

			p.assignments = append(p.assignments, data.Assignment{Expression: p.expr,
				Target: p.bv})
			p.bv.IDs = nil

		case ruleAction4:

			p.bv.Kind = data.BoundSelf

		case ruleAction5:

			p.bv.Kind = data.BoundDataset

		case ruleAction6:

			p.bv.Kind = data.BoundProperty

		case ruleAction7:

			p.bv.Kind = data.BoundStyle

		case ruleAction8:

			p.bv.Kind = data.BoundClass

		case ruleAction9:

			p.bv.Kind = data.BoundFormValue

		case ruleAction10:

			p.bv.Kind = data.BoundExpr
			p.bv.IDs = append(p.bv.IDs, p.expr)

		case ruleAction11:

			p.bv.Kind = data.BoundEventValue
			if len(p.bv.IDs) == 0 {
				p.bv.IDs = append(p.bv.IDs, "")
			}

		case ruleAction12:

			p.bv.IDs = append(p.bv.IDs, buffer[begin:end])

		case ruleAction13:

			p.bv.IDs = append(p.bv.IDs, buffer[begin:end])

		case ruleAction14:

			p.expr = buffer[begin:end]

		case ruleAction15:

			var expr *string
			if p.expr != "" {
				expr = new(string)
				*expr = p.expr
			}
			for _, name := range p.names {
				p.fields = append(p.fields, &data.Field{Name: name, Type: p.valuetype, DefaultValue: expr})
			}
			p.expr = ""
			p.valuetype = nil
			p.names = nil

		case ruleAction16:

			p.names = append(p.names, buffer[begin:end])

		case ruleAction17:

			switch name := buffer[begin:end]; name {
			case "int":
				p.valuetype = &data.ParamType{Kind: data.IntType}
			case "bool":
				p.valuetype = &data.ParamType{Kind: data.BoolType}
			case "string":
				p.valuetype = &data.ParamType{Kind: data.StringType}
			default:
				p.valuetype = &data.ParamType{Kind: data.NamedType, Name: name}
			}

		case ruleAction18:

			name := buffer[begin:end]
			if name == "js.Value" {
				p.valuetype = &data.ParamType{Kind: data.JSValueType}
			} else {
				p.valuetype = &data.ParamType{Kind: data.NamedType, Name: name}
			}

		case ruleAction19:

			p.valuetype = &data.ParamType{Kind: data.ArrayType, ValueType: p.valuetype}

		case ruleAction20:

			p.valuetype = &data.ParamType{Kind: data.MapType, KeyType: p.keytype, ValueType: p.valuetype}

		case ruleAction21:

			p.valuetype = &data.ParamType{Kind: data.ChanType, ValueType: p.valuetype}

		case ruleAction22:

			p.valuetype = &data.ParamType{Kind: data.FuncType, ValueType: p.valuetype,
				Params: p.params}
			p.params = nil

		case ruleAction23:

			p.keytype = p.valuetype

		case ruleAction24:

			p.valuetype = &data.ParamType{Kind: data.PointerType, ValueType: p.valuetype}

		case ruleAction25:

			p.eventMappings = append(p.eventMappings, data.UnboundEventMapping{
				Event: p.eventName, Handler: p.handlername, ParamMappings: p.paramMappings,
				Handling: p.eventHandling})
			p.eventHandling = data.AutoPreventDefault
			p.expr = ""
			p.paramMappings = make(map[string]data.BoundValue)

		case ruleAction26:

			p.handlername = buffer[begin:end]

		case ruleAction27:

			p.eventName = buffer[begin:end]

		case ruleAction28:

			p.paramIndex = 0
			p.tagname = ""

		case ruleAction29:

			if p.tagname == "" {
				if p.paramIndex == -1 {
					p.err = errors.New("unnamed parameter mapping after named one")
					return
				}
				p.tagname = fmt.Sprintf("~%v", p.paramIndex)
				p.paramIndex++
			} else {
				if _, ok := p.paramMappings[p.tagname]; ok {
					p.err = errors.New("duplicate param: " + p.tagname)
					return
				}
				p.paramIndex = -1
			}
			p.paramMappings[p.tagname] = p.bv
			p.tagname = ""
			p.bv.IDs = nil

		case ruleAction30:

			p.tagname = buffer[begin:end]

		case ruleAction31:

			switch p.tagname {
			case "preventDefault":
				if p.eventHandling != data.AutoPreventDefault {
					p.err = errors.New("duplicate preventDefault")
					return
				}
				switch len(p.names) {
				case 0:
					p.eventHandling = data.PreventDefault
				case 1:
					switch p.names[0] {
					case "true":
						p.eventHandling = data.PreventDefault
					case "false":
						p.eventHandling = data.DontPreventDefault
					case "ask":
						p.eventHandling = data.AskPreventDefault
					default:
						p.err = fmt.Errorf("unsupported value for preventDefault: %s", p.names[0])
						return
					}
				default:
					p.err = errors.New("too many parameters for preventDefault")
					return
				}
			default:
				p.err = errors.New("unknown tag: " + p.tagname)
				return
			}
			p.names = nil

		case ruleAction32:

			p.tagname = buffer[begin:end]

		case ruleAction33:

			p.names = append(p.names, buffer[begin:end])

		case ruleAction34:

			p.names = append(p.names, buffer[begin:end])

		case ruleAction35:

			p.handlers = append(p.handlers, HandlerSpec{
				Name: p.handlername, Params: p.params, Returns: p.valuetype})
			p.valuetype = nil
			p.params = nil

		case ruleAction36:

			p.paramnames = append(p.paramnames, buffer[begin:end])

		case ruleAction37:

			name := p.paramnames[len(p.paramnames)-1]
			p.paramnames = p.paramnames[:len(p.paramnames)-1]
			for _, para := range p.params {
				if para.Name == name {
					p.err = errors.New("duplicate param name: " + para.Name)
					return
				}
			}

			p.params = append(p.params, data.Param{Name: name, Type: p.valuetype})
			p.valuetype = nil

		case ruleAction38:

			p.cParams = append(p.cParams, data.ComponentParam{
				Name: p.tagname, Type: *p.valuetype, IsVar: p.isVar})
			p.valuetype = nil
			p.isVar = false

		case ruleAction39:

			p.isVar = true

		case ruleAction40:

			p.names = append(p.names, p.expr)

		case ruleAction41:

			path := buffer[begin:end]
			if p.tagname == "" {
				lastDot := strings.LastIndexByte(path, '/')
				if lastDot == -1 {
					p.tagname = path
				} else {
					p.tagname = path[lastDot+1:]
				}
			}
			if _, ok := p.imports[p.tagname]; ok {
				p.err = errors.New("duplicate import name: " + p.tagname)
				return
			}
			p.imports[p.tagname] = path
			p.tagname = ""

		}
	}
	_, _, _, _, _ = buffer, _buffer, text, begin, end
}

func (p *GeneralParser) Init() {
	p.buffer = []rune(p.Buffer)
	if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != endSymbol {
		p.buffer = append(p.buffer, endSymbol)
	}

	tree := tokens32{tree: make([]token32, math.MaxInt16)}
	var max token32
	position, depth, tokenIndex, buffer, _rules := uint32(0), uint32(0), 0, p.buffer, p.rules

	p.Parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokens32 = tree
		if matches {
			p.trim(tokenIndex)
			return nil
		}
		return &parseError{p, max}
	}

	p.Reset = func() {
		position, tokenIndex, depth = 0, 0, 0
	}

	add := func(rule pegRule, begin uint32) {
		tree.Expand(tokenIndex)
		tree.Add(rule, begin, position, depth, tokenIndex)
		tokenIndex++
		if begin != position && position > max.end {
			max = token32{rule, begin, position, depth}
		}
	}

	matchDot := func() bool {
		if buffer[position] != endSymbol {
			position++
			return true
		}
		return false
	}

	/*matchChar := func(c byte) bool {
		if buffer[position] == c {
			position++
			return true
		}
		return false
	}*/

	/*matchRange := func(lower byte, upper byte) bool {
		if c := buffer[position]; c >= lower && c <= upper {
			position++
			return true
		}
		return false
	}*/

	_rules = [...]func() bool{
		nil,
		/* 0 e <- <(assignments / bindings / captures / fields / for / handlers / cparams / args / imports)> */
		func() bool {
			position0, tokenIndex0, depth0 := position, tokenIndex, depth
			{
				position1 := position
				depth++
				{
					position2, tokenIndex2, depth2 := position, tokenIndex, depth
					if !_rules[ruleassignments]() {
						goto l3
					}
					goto l2
				l3:
					position, tokenIndex, depth = position2, tokenIndex2, depth2
					if !_rules[rulebindings]() {
						goto l4
					}
					goto l2
				l4:
					position, tokenIndex, depth = position2, tokenIndex2, depth2
					if !_rules[rulecaptures]() {
						goto l5
					}
					goto l2
				l5:
					position, tokenIndex, depth = position2, tokenIndex2, depth2
					if !_rules[rulefields]() {
						goto l6
					}
					goto l2
				l6:
					position, tokenIndex, depth = position2, tokenIndex2, depth2
					if !_rules[rulefor]() {
						goto l7
					}
					goto l2
				l7:
					position, tokenIndex, depth = position2, tokenIndex2, depth2
					if !_rules[rulehandlers]() {
						goto l8
					}
					goto l2
				l8:
					position, tokenIndex, depth = position2, tokenIndex2, depth2
					if !_rules[rulecparams]() {
						goto l9
					}
					goto l2
				l9:
					position, tokenIndex, depth = position2, tokenIndex2, depth2
					if !_rules[ruleargs]() {
						goto l10
					}
					goto l2
				l10:
					position, tokenIndex, depth = position2, tokenIndex2, depth2
					if !_rules[ruleimports]() {
						goto l0
					}
				}
			l2:
				depth--
				add(rulee, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 assignments <- <(isp* assignment isp* ((',' / ';') isp* assignment isp*)* !.)> */
		func() bool {
			position11, tokenIndex11, depth11 := position, tokenIndex, depth
			{
				position12 := position
				depth++
			l13:
				{
					position14, tokenIndex14, depth14 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l14
					}
					goto l13
				l14:
					position, tokenIndex, depth = position14, tokenIndex14, depth14
				}
				if !_rules[ruleassignment]() {
					goto l11
				}
			l15:
				{
					position16, tokenIndex16, depth16 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l16
					}
					goto l15
				l16:
					position, tokenIndex, depth = position16, tokenIndex16, depth16
				}
			l17:
				{
					position18, tokenIndex18, depth18 := position, tokenIndex, depth
					{
						position19, tokenIndex19, depth19 := position, tokenIndex, depth
						if buffer[position] != rune(',') {
							goto l20
						}
						position++
						goto l19
					l20:
						position, tokenIndex, depth = position19, tokenIndex19, depth19
						if buffer[position] != rune(';') {
							goto l18
						}
						position++
					}
				l19:
				l21:
					{
						position22, tokenIndex22, depth22 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l22
						}
						goto l21
					l22:
						position, tokenIndex, depth = position22, tokenIndex22, depth22
					}
					if !_rules[ruleassignment]() {
						goto l18
					}
				l23:
					{
						position24, tokenIndex24, depth24 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l24
						}
						goto l23
					l24:
						position, tokenIndex, depth = position24, tokenIndex24, depth24
					}
					goto l17
				l18:
					position, tokenIndex, depth = position18, tokenIndex18, depth18
				}
				{
					position25, tokenIndex25, depth25 := position, tokenIndex, depth
					if !matchDot() {
						goto l25
					}
					goto l11
				l25:
					position, tokenIndex, depth = position25, tokenIndex25, depth25
				}
				depth--
				add(ruleassignments, position12)
			}
			return true
		l11:
			position, tokenIndex, depth = position11, tokenIndex11, depth11
			return false
		},
		/* 2 bindings <- <(isp* binding isp* ((',' / ';') isp* binding isp*)* !.)> */
		func() bool {
			position26, tokenIndex26, depth26 := position, tokenIndex, depth
			{
				position27 := position
				depth++
			l28:
				{
					position29, tokenIndex29, depth29 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l29
					}
					goto l28
				l29:
					position, tokenIndex, depth = position29, tokenIndex29, depth29
				}
				if !_rules[rulebinding]() {
					goto l26
				}
			l30:
				{
					position31, tokenIndex31, depth31 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l31
					}
					goto l30
				l31:
					position, tokenIndex, depth = position31, tokenIndex31, depth31
				}
			l32:
				{
					position33, tokenIndex33, depth33 := position, tokenIndex, depth
					{
						position34, tokenIndex34, depth34 := position, tokenIndex, depth
						if buffer[position] != rune(',') {
							goto l35
						}
						position++
						goto l34
					l35:
						position, tokenIndex, depth = position34, tokenIndex34, depth34
						if buffer[position] != rune(';') {
							goto l33
						}
						position++
					}
				l34:
				l36:
					{
						position37, tokenIndex37, depth37 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l37
						}
						goto l36
					l37:
						position, tokenIndex, depth = position37, tokenIndex37, depth37
					}
					if !_rules[rulebinding]() {
						goto l33
					}
				l38:
					{
						position39, tokenIndex39, depth39 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l39
						}
						goto l38
					l39:
						position, tokenIndex, depth = position39, tokenIndex39, depth39
					}
					goto l32
				l33:
					position, tokenIndex, depth = position33, tokenIndex33, depth33
				}
				{
					position40, tokenIndex40, depth40 := position, tokenIndex, depth
					if !matchDot() {
						goto l40
					}
					goto l26
				l40:
					position, tokenIndex, depth = position40, tokenIndex40, depth40
				}
				depth--
				add(rulebindings, position27)
			}
			return true
		l26:
			position, tokenIndex, depth = position26, tokenIndex26, depth26
			return false
		},
		/* 3 binding <- <(bound isp* ':' isp* (autovar / typedvar) Action0)> */
		func() bool {
			position41, tokenIndex41, depth41 := position, tokenIndex, depth
			{
				position42 := position
				depth++
				if !_rules[rulebound]() {
					goto l41
				}
			l43:
				{
					position44, tokenIndex44, depth44 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l44
					}
					goto l43
				l44:
					position, tokenIndex, depth = position44, tokenIndex44, depth44
				}
				if buffer[position] != rune(':') {
					goto l41
				}
				position++
			l45:
				{
					position46, tokenIndex46, depth46 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l46
					}
					goto l45
				l46:
					position, tokenIndex, depth = position46, tokenIndex46, depth46
				}
				{
					position47, tokenIndex47, depth47 := position, tokenIndex, depth
					if !_rules[ruleautovar]() {
						goto l48
					}
					goto l47
				l48:
					position, tokenIndex, depth = position47, tokenIndex47, depth47
					if !_rules[ruletypedvar]() {
						goto l41
					}
				}
			l47:
				if !_rules[ruleAction0]() {
					goto l41
				}
				depth--
				add(rulebinding, position42)
			}
			return true
		l41:
			position, tokenIndex, depth = position41, tokenIndex41, depth41
			return false
		},
		/* 4 autovar <- <(<identifier> Action1)> */
		func() bool {
			position49, tokenIndex49, depth49 := position, tokenIndex, depth
			{
				position50 := position
				depth++
				{
					position51 := position
					depth++
					if !_rules[ruleidentifier]() {
						goto l49
					}
					depth--
					add(rulePegText, position51)
				}
				if !_rules[ruleAction1]() {
					goto l49
				}
				depth--
				add(ruleautovar, position50)
			}
			return true
		l49:
			position, tokenIndex, depth = position49, tokenIndex49, depth49
			return false
		},
		/* 5 typedvar <- <('(' isp* autovar isp+ type isp* ')' Action2)> */
		func() bool {
			position52, tokenIndex52, depth52 := position, tokenIndex, depth
			{
				position53 := position
				depth++
				if buffer[position] != rune('(') {
					goto l52
				}
				position++
			l54:
				{
					position55, tokenIndex55, depth55 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l55
					}
					goto l54
				l55:
					position, tokenIndex, depth = position55, tokenIndex55, depth55
				}
				if !_rules[ruleautovar]() {
					goto l52
				}
				if !_rules[ruleisp]() {
					goto l52
				}
			l56:
				{
					position57, tokenIndex57, depth57 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l57
					}
					goto l56
				l57:
					position, tokenIndex, depth = position57, tokenIndex57, depth57
				}
				if !_rules[ruletype]() {
					goto l52
				}
			l58:
				{
					position59, tokenIndex59, depth59 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l59
					}
					goto l58
				l59:
					position, tokenIndex, depth = position59, tokenIndex59, depth59
				}
				if buffer[position] != rune(')') {
					goto l52
				}
				position++
				if !_rules[ruleAction2]() {
					goto l52
				}
				depth--
				add(ruletypedvar, position53)
			}
			return true
		l52:
			position, tokenIndex, depth = position52, tokenIndex52, depth52
			return false
		},
		/* 6 isp <- <(' ' / '\t')> */
		func() bool {
			position60, tokenIndex60, depth60 := position, tokenIndex, depth
			{
				position61 := position
				depth++
				{
					position62, tokenIndex62, depth62 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l63
					}
					position++
					goto l62
				l63:
					position, tokenIndex, depth = position62, tokenIndex62, depth62
					if buffer[position] != rune('\t') {
						goto l60
					}
					position++
				}
			l62:
				depth--
				add(ruleisp, position61)
			}
			return true
		l60:
			position, tokenIndex, depth = position60, tokenIndex60, depth60
			return false
		},
		/* 7 assignment <- <(isp* bound isp* '=' isp* expr Action3)> */
		func() bool {
			position64, tokenIndex64, depth64 := position, tokenIndex, depth
			{
				position65 := position
				depth++
			l66:
				{
					position67, tokenIndex67, depth67 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l67
					}
					goto l66
				l67:
					position, tokenIndex, depth = position67, tokenIndex67, depth67
				}
				if !_rules[rulebound]() {
					goto l64
				}
			l68:
				{
					position69, tokenIndex69, depth69 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l69
					}
					goto l68
				l69:
					position, tokenIndex, depth = position69, tokenIndex69, depth69
				}
				if buffer[position] != rune('=') {
					goto l64
				}
				position++
			l70:
				{
					position71, tokenIndex71, depth71 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l71
					}
					goto l70
				l71:
					position, tokenIndex, depth = position71, tokenIndex71, depth71
				}
				if !_rules[ruleexpr]() {
					goto l64
				}
				if !_rules[ruleAction3]() {
					goto l64
				}
				depth--
				add(ruleassignment, position65)
			}
			return true
		l64:
			position, tokenIndex, depth = position64, tokenIndex64, depth64
			return false
		},
		/* 8 bound <- <(self / ((&('E' | 'e') event) | (&('F' | 'f') form) | (&('G' | 'g') goExpr) | (&('C' | 'c') class) | (&('S' | 's') style) | (&('P' | 'p') prop) | (&('D' | 'd') dataset)))> */
		func() bool {
			position72, tokenIndex72, depth72 := position, tokenIndex, depth
			{
				position73 := position
				depth++
				{
					position74, tokenIndex74, depth74 := position, tokenIndex, depth
					if !_rules[ruleself]() {
						goto l75
					}
					goto l74
				l75:
					position, tokenIndex, depth = position74, tokenIndex74, depth74
					{
						switch buffer[position] {
						case 'E', 'e':
							if !_rules[ruleevent]() {
								goto l72
							}
							break
						case 'F', 'f':
							if !_rules[ruleform]() {
								goto l72
							}
							break
						case 'G', 'g':
							if !_rules[rulegoExpr]() {
								goto l72
							}
							break
						case 'C', 'c':
							if !_rules[ruleclass]() {
								goto l72
							}
							break
						case 'S', 's':
							if !_rules[rulestyle]() {
								goto l72
							}
							break
						case 'P', 'p':
							if !_rules[ruleprop]() {
								goto l72
							}
							break
						default:
							if !_rules[ruledataset]() {
								goto l72
							}
							break
						}
					}

				}
			l74:
				depth--
				add(rulebound, position73)
			}
			return true
		l72:
			position, tokenIndex, depth = position72, tokenIndex72, depth72
			return false
		},
		/* 9 self <- <(('s' / 'S') ('e' / 'E') ('l' / 'L') ('f' / 'F') isp* '(' isp* ')' Action4)> */
		func() bool {
			position77, tokenIndex77, depth77 := position, tokenIndex, depth
			{
				position78 := position
				depth++
				{
					position79, tokenIndex79, depth79 := position, tokenIndex, depth
					if buffer[position] != rune('s') {
						goto l80
					}
					position++
					goto l79
				l80:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
					if buffer[position] != rune('S') {
						goto l77
					}
					position++
				}
			l79:
				{
					position81, tokenIndex81, depth81 := position, tokenIndex, depth
					if buffer[position] != rune('e') {
						goto l82
					}
					position++
					goto l81
				l82:
					position, tokenIndex, depth = position81, tokenIndex81, depth81
					if buffer[position] != rune('E') {
						goto l77
					}
					position++
				}
			l81:
				{
					position83, tokenIndex83, depth83 := position, tokenIndex, depth
					if buffer[position] != rune('l') {
						goto l84
					}
					position++
					goto l83
				l84:
					position, tokenIndex, depth = position83, tokenIndex83, depth83
					if buffer[position] != rune('L') {
						goto l77
					}
					position++
				}
			l83:
				{
					position85, tokenIndex85, depth85 := position, tokenIndex, depth
					if buffer[position] != rune('f') {
						goto l86
					}
					position++
					goto l85
				l86:
					position, tokenIndex, depth = position85, tokenIndex85, depth85
					if buffer[position] != rune('F') {
						goto l77
					}
					position++
				}
			l85:
			l87:
				{
					position88, tokenIndex88, depth88 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l88
					}
					goto l87
				l88:
					position, tokenIndex, depth = position88, tokenIndex88, depth88
				}
				if buffer[position] != rune('(') {
					goto l77
				}
				position++
			l89:
				{
					position90, tokenIndex90, depth90 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l90
					}
					goto l89
				l90:
					position, tokenIndex, depth = position90, tokenIndex90, depth90
				}
				if buffer[position] != rune(')') {
					goto l77
				}
				position++
				if !_rules[ruleAction4]() {
					goto l77
				}
				depth--
				add(ruleself, position78)
			}
			return true
		l77:
			position, tokenIndex, depth = position77, tokenIndex77, depth77
			return false
		},
		/* 10 dataset <- <(('d' / 'D') ('a' / 'A') ('t' / 'T') ('a' / 'A') ('s' / 'S') ('e' / 'E') ('t' / 'T') isp* '(' isp* htmlid isp* ')' Action5)> */
		func() bool {
			position91, tokenIndex91, depth91 := position, tokenIndex, depth
			{
				position92 := position
				depth++
				{
					position93, tokenIndex93, depth93 := position, tokenIndex, depth
					if buffer[position] != rune('d') {
						goto l94
					}
					position++
					goto l93
				l94:
					position, tokenIndex, depth = position93, tokenIndex93, depth93
					if buffer[position] != rune('D') {
						goto l91
					}
					position++
				}
			l93:
				{
					position95, tokenIndex95, depth95 := position, tokenIndex, depth
					if buffer[position] != rune('a') {
						goto l96
					}
					position++
					goto l95
				l96:
					position, tokenIndex, depth = position95, tokenIndex95, depth95
					if buffer[position] != rune('A') {
						goto l91
					}
					position++
				}
			l95:
				{
					position97, tokenIndex97, depth97 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l98
					}
					position++
					goto l97
				l98:
					position, tokenIndex, depth = position97, tokenIndex97, depth97
					if buffer[position] != rune('T') {
						goto l91
					}
					position++
				}
			l97:
				{
					position99, tokenIndex99, depth99 := position, tokenIndex, depth
					if buffer[position] != rune('a') {
						goto l100
					}
					position++
					goto l99
				l100:
					position, tokenIndex, depth = position99, tokenIndex99, depth99
					if buffer[position] != rune('A') {
						goto l91
					}
					position++
				}
			l99:
				{
					position101, tokenIndex101, depth101 := position, tokenIndex, depth
					if buffer[position] != rune('s') {
						goto l102
					}
					position++
					goto l101
				l102:
					position, tokenIndex, depth = position101, tokenIndex101, depth101
					if buffer[position] != rune('S') {
						goto l91
					}
					position++
				}
			l101:
				{
					position103, tokenIndex103, depth103 := position, tokenIndex, depth
					if buffer[position] != rune('e') {
						goto l104
					}
					position++
					goto l103
				l104:
					position, tokenIndex, depth = position103, tokenIndex103, depth103
					if buffer[position] != rune('E') {
						goto l91
					}
					position++
				}
			l103:
				{
					position105, tokenIndex105, depth105 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l106
					}
					position++
					goto l105
				l106:
					position, tokenIndex, depth = position105, tokenIndex105, depth105
					if buffer[position] != rune('T') {
						goto l91
					}
					position++
				}
			l105:
			l107:
				{
					position108, tokenIndex108, depth108 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l108
					}
					goto l107
				l108:
					position, tokenIndex, depth = position108, tokenIndex108, depth108
				}
				if buffer[position] != rune('(') {
					goto l91
				}
				position++
			l109:
				{
					position110, tokenIndex110, depth110 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l110
					}
					goto l109
				l110:
					position, tokenIndex, depth = position110, tokenIndex110, depth110
				}
				if !_rules[rulehtmlid]() {
					goto l91
				}
			l111:
				{
					position112, tokenIndex112, depth112 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l112
					}
					goto l111
				l112:
					position, tokenIndex, depth = position112, tokenIndex112, depth112
				}
				if buffer[position] != rune(')') {
					goto l91
				}
				position++
				if !_rules[ruleAction5]() {
					goto l91
				}
				depth--
				add(ruledataset, position92)
			}
			return true
		l91:
			position, tokenIndex, depth = position91, tokenIndex91, depth91
			return false
		},
		/* 11 prop <- <(('p' / 'P') ('r' / 'R') ('o' / 'O') ('p' / 'P') isp* '(' isp* htmlid isp* ')' Action6)> */
		func() bool {
			position113, tokenIndex113, depth113 := position, tokenIndex, depth
			{
				position114 := position
				depth++
				{
					position115, tokenIndex115, depth115 := position, tokenIndex, depth
					if buffer[position] != rune('p') {
						goto l116
					}
					position++
					goto l115
				l116:
					position, tokenIndex, depth = position115, tokenIndex115, depth115
					if buffer[position] != rune('P') {
						goto l113
					}
					position++
				}
			l115:
				{
					position117, tokenIndex117, depth117 := position, tokenIndex, depth
					if buffer[position] != rune('r') {
						goto l118
					}
					position++
					goto l117
				l118:
					position, tokenIndex, depth = position117, tokenIndex117, depth117
					if buffer[position] != rune('R') {
						goto l113
					}
					position++
				}
			l117:
				{
					position119, tokenIndex119, depth119 := position, tokenIndex, depth
					if buffer[position] != rune('o') {
						goto l120
					}
					position++
					goto l119
				l120:
					position, tokenIndex, depth = position119, tokenIndex119, depth119
					if buffer[position] != rune('O') {
						goto l113
					}
					position++
				}
			l119:
				{
					position121, tokenIndex121, depth121 := position, tokenIndex, depth
					if buffer[position] != rune('p') {
						goto l122
					}
					position++
					goto l121
				l122:
					position, tokenIndex, depth = position121, tokenIndex121, depth121
					if buffer[position] != rune('P') {
						goto l113
					}
					position++
				}
			l121:
			l123:
				{
					position124, tokenIndex124, depth124 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l124
					}
					goto l123
				l124:
					position, tokenIndex, depth = position124, tokenIndex124, depth124
				}
				if buffer[position] != rune('(') {
					goto l113
				}
				position++
			l125:
				{
					position126, tokenIndex126, depth126 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l126
					}
					goto l125
				l126:
					position, tokenIndex, depth = position126, tokenIndex126, depth126
				}
				if !_rules[rulehtmlid]() {
					goto l113
				}
			l127:
				{
					position128, tokenIndex128, depth128 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l128
					}
					goto l127
				l128:
					position, tokenIndex, depth = position128, tokenIndex128, depth128
				}
				if buffer[position] != rune(')') {
					goto l113
				}
				position++
				if !_rules[ruleAction6]() {
					goto l113
				}
				depth--
				add(ruleprop, position114)
			}
			return true
		l113:
			position, tokenIndex, depth = position113, tokenIndex113, depth113
			return false
		},
		/* 12 style <- <(('s' / 'S') ('t' / 'T') ('y' / 'Y') ('l' / 'L') ('e' / 'E') isp* '(' isp* htmlid isp* ')' Action7)> */
		func() bool {
			position129, tokenIndex129, depth129 := position, tokenIndex, depth
			{
				position130 := position
				depth++
				{
					position131, tokenIndex131, depth131 := position, tokenIndex, depth
					if buffer[position] != rune('s') {
						goto l132
					}
					position++
					goto l131
				l132:
					position, tokenIndex, depth = position131, tokenIndex131, depth131
					if buffer[position] != rune('S') {
						goto l129
					}
					position++
				}
			l131:
				{
					position133, tokenIndex133, depth133 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l134
					}
					position++
					goto l133
				l134:
					position, tokenIndex, depth = position133, tokenIndex133, depth133
					if buffer[position] != rune('T') {
						goto l129
					}
					position++
				}
			l133:
				{
					position135, tokenIndex135, depth135 := position, tokenIndex, depth
					if buffer[position] != rune('y') {
						goto l136
					}
					position++
					goto l135
				l136:
					position, tokenIndex, depth = position135, tokenIndex135, depth135
					if buffer[position] != rune('Y') {
						goto l129
					}
					position++
				}
			l135:
				{
					position137, tokenIndex137, depth137 := position, tokenIndex, depth
					if buffer[position] != rune('l') {
						goto l138
					}
					position++
					goto l137
				l138:
					position, tokenIndex, depth = position137, tokenIndex137, depth137
					if buffer[position] != rune('L') {
						goto l129
					}
					position++
				}
			l137:
				{
					position139, tokenIndex139, depth139 := position, tokenIndex, depth
					if buffer[position] != rune('e') {
						goto l140
					}
					position++
					goto l139
				l140:
					position, tokenIndex, depth = position139, tokenIndex139, depth139
					if buffer[position] != rune('E') {
						goto l129
					}
					position++
				}
			l139:
			l141:
				{
					position142, tokenIndex142, depth142 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l142
					}
					goto l141
				l142:
					position, tokenIndex, depth = position142, tokenIndex142, depth142
				}
				if buffer[position] != rune('(') {
					goto l129
				}
				position++
			l143:
				{
					position144, tokenIndex144, depth144 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l144
					}
					goto l143
				l144:
					position, tokenIndex, depth = position144, tokenIndex144, depth144
				}
				if !_rules[rulehtmlid]() {
					goto l129
				}
			l145:
				{
					position146, tokenIndex146, depth146 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l146
					}
					goto l145
				l146:
					position, tokenIndex, depth = position146, tokenIndex146, depth146
				}
				if buffer[position] != rune(')') {
					goto l129
				}
				position++
				if !_rules[ruleAction7]() {
					goto l129
				}
				depth--
				add(rulestyle, position130)
			}
			return true
		l129:
			position, tokenIndex, depth = position129, tokenIndex129, depth129
			return false
		},
		/* 13 class <- <(('c' / 'C') ('l' / 'L') ('a' / 'A') ('s' / 'S') ('s' / 'S') isp* '(' isp* htmlid isp* (',' isp* htmlid isp*)* ')' Action8)> */
		func() bool {
			position147, tokenIndex147, depth147 := position, tokenIndex, depth
			{
				position148 := position
				depth++
				{
					position149, tokenIndex149, depth149 := position, tokenIndex, depth
					if buffer[position] != rune('c') {
						goto l150
					}
					position++
					goto l149
				l150:
					position, tokenIndex, depth = position149, tokenIndex149, depth149
					if buffer[position] != rune('C') {
						goto l147
					}
					position++
				}
			l149:
				{
					position151, tokenIndex151, depth151 := position, tokenIndex, depth
					if buffer[position] != rune('l') {
						goto l152
					}
					position++
					goto l151
				l152:
					position, tokenIndex, depth = position151, tokenIndex151, depth151
					if buffer[position] != rune('L') {
						goto l147
					}
					position++
				}
			l151:
				{
					position153, tokenIndex153, depth153 := position, tokenIndex, depth
					if buffer[position] != rune('a') {
						goto l154
					}
					position++
					goto l153
				l154:
					position, tokenIndex, depth = position153, tokenIndex153, depth153
					if buffer[position] != rune('A') {
						goto l147
					}
					position++
				}
			l153:
				{
					position155, tokenIndex155, depth155 := position, tokenIndex, depth
					if buffer[position] != rune('s') {
						goto l156
					}
					position++
					goto l155
				l156:
					position, tokenIndex, depth = position155, tokenIndex155, depth155
					if buffer[position] != rune('S') {
						goto l147
					}
					position++
				}
			l155:
				{
					position157, tokenIndex157, depth157 := position, tokenIndex, depth
					if buffer[position] != rune('s') {
						goto l158
					}
					position++
					goto l157
				l158:
					position, tokenIndex, depth = position157, tokenIndex157, depth157
					if buffer[position] != rune('S') {
						goto l147
					}
					position++
				}
			l157:
			l159:
				{
					position160, tokenIndex160, depth160 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l160
					}
					goto l159
				l160:
					position, tokenIndex, depth = position160, tokenIndex160, depth160
				}
				if buffer[position] != rune('(') {
					goto l147
				}
				position++
			l161:
				{
					position162, tokenIndex162, depth162 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l162
					}
					goto l161
				l162:
					position, tokenIndex, depth = position162, tokenIndex162, depth162
				}
				if !_rules[rulehtmlid]() {
					goto l147
				}
			l163:
				{
					position164, tokenIndex164, depth164 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l164
					}
					goto l163
				l164:
					position, tokenIndex, depth = position164, tokenIndex164, depth164
				}
			l165:
				{
					position166, tokenIndex166, depth166 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l166
					}
					position++
				l167:
					{
						position168, tokenIndex168, depth168 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l168
						}
						goto l167
					l168:
						position, tokenIndex, depth = position168, tokenIndex168, depth168
					}
					if !_rules[rulehtmlid]() {
						goto l166
					}
				l169:
					{
						position170, tokenIndex170, depth170 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l170
						}
						goto l169
					l170:
						position, tokenIndex, depth = position170, tokenIndex170, depth170
					}
					goto l165
				l166:
					position, tokenIndex, depth = position166, tokenIndex166, depth166
				}
				if buffer[position] != rune(')') {
					goto l147
				}
				position++
				if !_rules[ruleAction8]() {
					goto l147
				}
				depth--
				add(ruleclass, position148)
			}
			return true
		l147:
			position, tokenIndex, depth = position147, tokenIndex147, depth147
			return false
		},
		/* 14 form <- <(('f' / 'F') ('o' / 'O') ('r' / 'R') ('m' / 'M') isp* '(' isp* htmlid isp* ')' Action9)> */
		func() bool {
			position171, tokenIndex171, depth171 := position, tokenIndex, depth
			{
				position172 := position
				depth++
				{
					position173, tokenIndex173, depth173 := position, tokenIndex, depth
					if buffer[position] != rune('f') {
						goto l174
					}
					position++
					goto l173
				l174:
					position, tokenIndex, depth = position173, tokenIndex173, depth173
					if buffer[position] != rune('F') {
						goto l171
					}
					position++
				}
			l173:
				{
					position175, tokenIndex175, depth175 := position, tokenIndex, depth
					if buffer[position] != rune('o') {
						goto l176
					}
					position++
					goto l175
				l176:
					position, tokenIndex, depth = position175, tokenIndex175, depth175
					if buffer[position] != rune('O') {
						goto l171
					}
					position++
				}
			l175:
				{
					position177, tokenIndex177, depth177 := position, tokenIndex, depth
					if buffer[position] != rune('r') {
						goto l178
					}
					position++
					goto l177
				l178:
					position, tokenIndex, depth = position177, tokenIndex177, depth177
					if buffer[position] != rune('R') {
						goto l171
					}
					position++
				}
			l177:
				{
					position179, tokenIndex179, depth179 := position, tokenIndex, depth
					if buffer[position] != rune('m') {
						goto l180
					}
					position++
					goto l179
				l180:
					position, tokenIndex, depth = position179, tokenIndex179, depth179
					if buffer[position] != rune('M') {
						goto l171
					}
					position++
				}
			l179:
			l181:
				{
					position182, tokenIndex182, depth182 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l182
					}
					goto l181
				l182:
					position, tokenIndex, depth = position182, tokenIndex182, depth182
				}
				if buffer[position] != rune('(') {
					goto l171
				}
				position++
			l183:
				{
					position184, tokenIndex184, depth184 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l184
					}
					goto l183
				l184:
					position, tokenIndex, depth = position184, tokenIndex184, depth184
				}
				if !_rules[rulehtmlid]() {
					goto l171
				}
			l185:
				{
					position186, tokenIndex186, depth186 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l186
					}
					goto l185
				l186:
					position, tokenIndex, depth = position186, tokenIndex186, depth186
				}
				if buffer[position] != rune(')') {
					goto l171
				}
				position++
				if !_rules[ruleAction9]() {
					goto l171
				}
				depth--
				add(ruleform, position172)
			}
			return true
		l171:
			position, tokenIndex, depth = position171, tokenIndex171, depth171
			return false
		},
		/* 15 goExpr <- <(('g' / 'G') ('o' / 'O') isp* '(' isp* expr isp* ')' Action10)> */
		func() bool {
			position187, tokenIndex187, depth187 := position, tokenIndex, depth
			{
				position188 := position
				depth++
				{
					position189, tokenIndex189, depth189 := position, tokenIndex, depth
					if buffer[position] != rune('g') {
						goto l190
					}
					position++
					goto l189
				l190:
					position, tokenIndex, depth = position189, tokenIndex189, depth189
					if buffer[position] != rune('G') {
						goto l187
					}
					position++
				}
			l189:
				{
					position191, tokenIndex191, depth191 := position, tokenIndex, depth
					if buffer[position] != rune('o') {
						goto l192
					}
					position++
					goto l191
				l192:
					position, tokenIndex, depth = position191, tokenIndex191, depth191
					if buffer[position] != rune('O') {
						goto l187
					}
					position++
				}
			l191:
			l193:
				{
					position194, tokenIndex194, depth194 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l194
					}
					goto l193
				l194:
					position, tokenIndex, depth = position194, tokenIndex194, depth194
				}
				if buffer[position] != rune('(') {
					goto l187
				}
				position++
			l195:
				{
					position196, tokenIndex196, depth196 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l196
					}
					goto l195
				l196:
					position, tokenIndex, depth = position196, tokenIndex196, depth196
				}
				if !_rules[ruleexpr]() {
					goto l187
				}
			l197:
				{
					position198, tokenIndex198, depth198 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l198
					}
					goto l197
				l198:
					position, tokenIndex, depth = position198, tokenIndex198, depth198
				}
				if buffer[position] != rune(')') {
					goto l187
				}
				position++
				if !_rules[ruleAction10]() {
					goto l187
				}
				depth--
				add(rulegoExpr, position188)
			}
			return true
		l187:
			position, tokenIndex, depth = position187, tokenIndex187, depth187
			return false
		},
		/* 16 event <- <(('e' / 'E') ('v' / 'V') ('e' / 'E') ('n' / 'N') ('t' / 'T') isp* '(' isp* jsid? isp* ')' Action11)> */
		func() bool {
			position199, tokenIndex199, depth199 := position, tokenIndex, depth
			{
				position200 := position
				depth++
				{
					position201, tokenIndex201, depth201 := position, tokenIndex, depth
					if buffer[position] != rune('e') {
						goto l202
					}
					position++
					goto l201
				l202:
					position, tokenIndex, depth = position201, tokenIndex201, depth201
					if buffer[position] != rune('E') {
						goto l199
					}
					position++
				}
			l201:
				{
					position203, tokenIndex203, depth203 := position, tokenIndex, depth
					if buffer[position] != rune('v') {
						goto l204
					}
					position++
					goto l203
				l204:
					position, tokenIndex, depth = position203, tokenIndex203, depth203
					if buffer[position] != rune('V') {
						goto l199
					}
					position++
				}
			l203:
				{
					position205, tokenIndex205, depth205 := position, tokenIndex, depth
					if buffer[position] != rune('e') {
						goto l206
					}
					position++
					goto l205
				l206:
					position, tokenIndex, depth = position205, tokenIndex205, depth205
					if buffer[position] != rune('E') {
						goto l199
					}
					position++
				}
			l205:
				{
					position207, tokenIndex207, depth207 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l208
					}
					position++
					goto l207
				l208:
					position, tokenIndex, depth = position207, tokenIndex207, depth207
					if buffer[position] != rune('N') {
						goto l199
					}
					position++
				}
			l207:
				{
					position209, tokenIndex209, depth209 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l210
					}
					position++
					goto l209
				l210:
					position, tokenIndex, depth = position209, tokenIndex209, depth209
					if buffer[position] != rune('T') {
						goto l199
					}
					position++
				}
			l209:
			l211:
				{
					position212, tokenIndex212, depth212 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l212
					}
					goto l211
				l212:
					position, tokenIndex, depth = position212, tokenIndex212, depth212
				}
				if buffer[position] != rune('(') {
					goto l199
				}
				position++
			l213:
				{
					position214, tokenIndex214, depth214 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l214
					}
					goto l213
				l214:
					position, tokenIndex, depth = position214, tokenIndex214, depth214
				}
				{
					position215, tokenIndex215, depth215 := position, tokenIndex, depth
					if !_rules[rulejsid]() {
						goto l215
					}
					goto l216
				l215:
					position, tokenIndex, depth = position215, tokenIndex215, depth215
				}
			l216:
			l217:
				{
					position218, tokenIndex218, depth218 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l218
					}
					goto l217
				l218:
					position, tokenIndex, depth = position218, tokenIndex218, depth218
				}
				if buffer[position] != rune(')') {
					goto l199
				}
				position++
				if !_rules[ruleAction11]() {
					goto l199
				}
				depth--
				add(ruleevent, position200)
			}
			return true
		l199:
			position, tokenIndex, depth = position199, tokenIndex199, depth199
			return false
		},
		/* 17 htmlid <- <(<((&('-') '-') | (&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))+> Action12)> */
		func() bool {
			position219, tokenIndex219, depth219 := position, tokenIndex, depth
			{
				position220 := position
				depth++
				{
					position221 := position
					depth++
					{
						switch buffer[position] {
						case '-':
							if buffer[position] != rune('-') {
								goto l219
							}
							position++
							break
						case '_':
							if buffer[position] != rune('_') {
								goto l219
							}
							position++
							break
						case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l219
							}
							position++
							break
						case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l219
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l219
							}
							position++
							break
						}
					}

				l222:
					{
						position223, tokenIndex223, depth223 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '-':
								if buffer[position] != rune('-') {
									goto l223
								}
								position++
								break
							case '_':
								if buffer[position] != rune('_') {
									goto l223
								}
								position++
								break
							case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
								if c := buffer[position]; c < rune('A') || c > rune('Z') {
									goto l223
								}
								position++
								break
							case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l223
								}
								position++
								break
							default:
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l223
								}
								position++
								break
							}
						}

						goto l222
					l223:
						position, tokenIndex, depth = position223, tokenIndex223, depth223
					}
					depth--
					add(rulePegText, position221)
				}
				if !_rules[ruleAction12]() {
					goto l219
				}
				depth--
				add(rulehtmlid, position220)
			}
			return true
		l219:
			position, tokenIndex, depth = position219, tokenIndex219, depth219
			return false
		},
		/* 18 jsid <- <(<(((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z])) ((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))*)> Action13)> */
		func() bool {
			position226, tokenIndex226, depth226 := position, tokenIndex, depth
			{
				position227 := position
				depth++
				{
					position228 := position
					depth++
					{
						switch buffer[position] {
						case '_':
							if buffer[position] != rune('_') {
								goto l226
							}
							position++
							break
						case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l226
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l226
							}
							position++
							break
						}
					}

				l230:
					{
						position231, tokenIndex231, depth231 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '_':
								if buffer[position] != rune('_') {
									goto l231
								}
								position++
								break
							case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
								if c := buffer[position]; c < rune('A') || c > rune('Z') {
									goto l231
								}
								position++
								break
							case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l231
								}
								position++
								break
							default:
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l231
								}
								position++
								break
							}
						}

						goto l230
					l231:
						position, tokenIndex, depth = position231, tokenIndex231, depth231
					}
					depth--
					add(rulePegText, position228)
				}
				if !_rules[ruleAction13]() {
					goto l226
				}
				depth--
				add(rulejsid, position227)
			}
			return true
		l226:
			position, tokenIndex, depth = position226, tokenIndex226, depth226
			return false
		},
		/* 19 expr <- <(<((&('\t' | ' ') isp+) | (&('(' | '[' | '{') enclosed) | (&('!' | '"' | '&' | '*' | '+' | '-' | '.' | '/' | '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9' | ':' | '<' | '=' | '>' | 'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '^' | '_' | '`' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z' | '|') commaless))+> Action14)> */
		func() bool {
			position233, tokenIndex233, depth233 := position, tokenIndex, depth
			{
				position234 := position
				depth++
				{
					position235 := position
					depth++
					{
						switch buffer[position] {
						case '\t', ' ':
							if !_rules[ruleisp]() {
								goto l233
							}
						l239:
							{
								position240, tokenIndex240, depth240 := position, tokenIndex, depth
								if !_rules[ruleisp]() {
									goto l240
								}
								goto l239
							l240:
								position, tokenIndex, depth = position240, tokenIndex240, depth240
							}
							break
						case '(', '[', '{':
							if !_rules[ruleenclosed]() {
								goto l233
							}
							break
						default:
							if !_rules[rulecommaless]() {
								goto l233
							}
							break
						}
					}

				l236:
					{
						position237, tokenIndex237, depth237 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '\t', ' ':
								if !_rules[ruleisp]() {
									goto l237
								}
							l242:
								{
									position243, tokenIndex243, depth243 := position, tokenIndex, depth
									if !_rules[ruleisp]() {
										goto l243
									}
									goto l242
								l243:
									position, tokenIndex, depth = position243, tokenIndex243, depth243
								}
								break
							case '(', '[', '{':
								if !_rules[ruleenclosed]() {
									goto l237
								}
								break
							default:
								if !_rules[rulecommaless]() {
									goto l237
								}
								break
							}
						}

						goto l236
					l237:
						position, tokenIndex, depth = position237, tokenIndex237, depth237
					}
					depth--
					add(rulePegText, position235)
				}
				if !_rules[ruleAction14]() {
					goto l233
				}
				depth--
				add(ruleexpr, position234)
			}
			return true
		l233:
			position, tokenIndex, depth = position233, tokenIndex233, depth233
			return false
		},
		/* 20 commaless <- <((((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))+ '.' ((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))+) / ((&('"' | '`') string) | (&('!' | '&' | '*' | '+' | '-' | '.' | '/' | ':' | '<' | '=' | '>' | '^' | '|') operators) | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') number) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '_' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') identifier)))> */
		func() bool {
			position244, tokenIndex244, depth244 := position, tokenIndex, depth
			{
				position245 := position
				depth++
				{
					position246, tokenIndex246, depth246 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '_':
							if buffer[position] != rune('_') {
								goto l247
							}
							position++
							break
						case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l247
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l247
							}
							position++
							break
						}
					}

				l248:
					{
						position249, tokenIndex249, depth249 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '_':
								if buffer[position] != rune('_') {
									goto l249
								}
								position++
								break
							case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
								if c := buffer[position]; c < rune('A') || c > rune('Z') {
									goto l249
								}
								position++
								break
							default:
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l249
								}
								position++
								break
							}
						}

						goto l248
					l249:
						position, tokenIndex, depth = position249, tokenIndex249, depth249
					}
					if buffer[position] != rune('.') {
						goto l247
					}
					position++
					{
						switch buffer[position] {
						case '_':
							if buffer[position] != rune('_') {
								goto l247
							}
							position++
							break
						case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l247
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l247
							}
							position++
							break
						}
					}

				l252:
					{
						position253, tokenIndex253, depth253 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '_':
								if buffer[position] != rune('_') {
									goto l253
								}
								position++
								break
							case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
								if c := buffer[position]; c < rune('A') || c > rune('Z') {
									goto l253
								}
								position++
								break
							default:
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l253
								}
								position++
								break
							}
						}

						goto l252
					l253:
						position, tokenIndex, depth = position253, tokenIndex253, depth253
					}
					goto l246
				l247:
					position, tokenIndex, depth = position246, tokenIndex246, depth246
					{
						switch buffer[position] {
						case '"', '`':
							if !_rules[rulestring]() {
								goto l244
							}
							break
						case '!', '&', '*', '+', '-', '.', '/', ':', '<', '=', '>', '^', '|':
							if !_rules[ruleoperators]() {
								goto l244
							}
							break
						case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
							if !_rules[rulenumber]() {
								goto l244
							}
							break
						default:
							if !_rules[ruleidentifier]() {
								goto l244
							}
							break
						}
					}

				}
			l246:
				depth--
				add(rulecommaless, position245)
			}
			return true
		l244:
			position, tokenIndex, depth = position244, tokenIndex244, depth244
			return false
		},
		/* 21 number <- <[0-9]+> */
		func() bool {
			position257, tokenIndex257, depth257 := position, tokenIndex, depth
			{
				position258 := position
				depth++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l257
				}
				position++
			l259:
				{
					position260, tokenIndex260, depth260 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l260
					}
					position++
					goto l259
				l260:
					position, tokenIndex, depth = position260, tokenIndex260, depth260
				}
				depth--
				add(rulenumber, position258)
			}
			return true
		l257:
			position, tokenIndex, depth = position257, tokenIndex257, depth257
			return false
		},
		/* 22 operators <- <((&('>') '>') | (&('<') '<') | (&('!') '!') | (&('.') '.') | (&('=') '=') | (&(':') ':') | (&('^') '^') | (&('&') '&') | (&('|') '|') | (&('/') '/') | (&('*') '*') | (&('-') '-') | (&('+') '+'))+> */
		func() bool {
			position261, tokenIndex261, depth261 := position, tokenIndex, depth
			{
				position262 := position
				depth++
				{
					switch buffer[position] {
					case '>':
						if buffer[position] != rune('>') {
							goto l261
						}
						position++
						break
					case '<':
						if buffer[position] != rune('<') {
							goto l261
						}
						position++
						break
					case '!':
						if buffer[position] != rune('!') {
							goto l261
						}
						position++
						break
					case '.':
						if buffer[position] != rune('.') {
							goto l261
						}
						position++
						break
					case '=':
						if buffer[position] != rune('=') {
							goto l261
						}
						position++
						break
					case ':':
						if buffer[position] != rune(':') {
							goto l261
						}
						position++
						break
					case '^':
						if buffer[position] != rune('^') {
							goto l261
						}
						position++
						break
					case '&':
						if buffer[position] != rune('&') {
							goto l261
						}
						position++
						break
					case '|':
						if buffer[position] != rune('|') {
							goto l261
						}
						position++
						break
					case '/':
						if buffer[position] != rune('/') {
							goto l261
						}
						position++
						break
					case '*':
						if buffer[position] != rune('*') {
							goto l261
						}
						position++
						break
					case '-':
						if buffer[position] != rune('-') {
							goto l261
						}
						position++
						break
					default:
						if buffer[position] != rune('+') {
							goto l261
						}
						position++
						break
					}
				}

			l263:
				{
					position264, tokenIndex264, depth264 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '>':
							if buffer[position] != rune('>') {
								goto l264
							}
							position++
							break
						case '<':
							if buffer[position] != rune('<') {
								goto l264
							}
							position++
							break
						case '!':
							if buffer[position] != rune('!') {
								goto l264
							}
							position++
							break
						case '.':
							if buffer[position] != rune('.') {
								goto l264
							}
							position++
							break
						case '=':
							if buffer[position] != rune('=') {
								goto l264
							}
							position++
							break
						case ':':
							if buffer[position] != rune(':') {
								goto l264
							}
							position++
							break
						case '^':
							if buffer[position] != rune('^') {
								goto l264
							}
							position++
							break
						case '&':
							if buffer[position] != rune('&') {
								goto l264
							}
							position++
							break
						case '|':
							if buffer[position] != rune('|') {
								goto l264
							}
							position++
							break
						case '/':
							if buffer[position] != rune('/') {
								goto l264
							}
							position++
							break
						case '*':
							if buffer[position] != rune('*') {
								goto l264
							}
							position++
							break
						case '-':
							if buffer[position] != rune('-') {
								goto l264
							}
							position++
							break
						default:
							if buffer[position] != rune('+') {
								goto l264
							}
							position++
							break
						}
					}

					goto l263
				l264:
					position, tokenIndex, depth = position264, tokenIndex264, depth264
				}
				depth--
				add(ruleoperators, position262)
			}
			return true
		l261:
			position, tokenIndex, depth = position261, tokenIndex261, depth261
			return false
		},
		/* 23 string <- <(('`' (!'`' .)* '`') / ('"' ((!'"' .) / ('\\' '"'))* '"'))> */
		func() bool {
			position267, tokenIndex267, depth267 := position, tokenIndex, depth
			{
				position268 := position
				depth++
				{
					position269, tokenIndex269, depth269 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l270
					}
					position++
				l271:
					{
						position272, tokenIndex272, depth272 := position, tokenIndex, depth
						{
							position273, tokenIndex273, depth273 := position, tokenIndex, depth
							if buffer[position] != rune('`') {
								goto l273
							}
							position++
							goto l272
						l273:
							position, tokenIndex, depth = position273, tokenIndex273, depth273
						}
						if !matchDot() {
							goto l272
						}
						goto l271
					l272:
						position, tokenIndex, depth = position272, tokenIndex272, depth272
					}
					if buffer[position] != rune('`') {
						goto l270
					}
					position++
					goto l269
				l270:
					position, tokenIndex, depth = position269, tokenIndex269, depth269
					if buffer[position] != rune('"') {
						goto l267
					}
					position++
				l274:
					{
						position275, tokenIndex275, depth275 := position, tokenIndex, depth
						{
							position276, tokenIndex276, depth276 := position, tokenIndex, depth
							{
								position278, tokenIndex278, depth278 := position, tokenIndex, depth
								if buffer[position] != rune('"') {
									goto l278
								}
								position++
								goto l277
							l278:
								position, tokenIndex, depth = position278, tokenIndex278, depth278
							}
							if !matchDot() {
								goto l277
							}
							goto l276
						l277:
							position, tokenIndex, depth = position276, tokenIndex276, depth276
							if buffer[position] != rune('\\') {
								goto l275
							}
							position++
							if buffer[position] != rune('"') {
								goto l275
							}
							position++
						}
					l276:
						goto l274
					l275:
						position, tokenIndex, depth = position275, tokenIndex275, depth275
					}
					if buffer[position] != rune('"') {
						goto l267
					}
					position++
				}
			l269:
				depth--
				add(rulestring, position268)
			}
			return true
		l267:
			position, tokenIndex, depth = position267, tokenIndex267, depth267
			return false
		},
		/* 24 enclosed <- <((&('[') brackets) | (&('{') braces) | (&('(') parens))> */
		func() bool {
			position279, tokenIndex279, depth279 := position, tokenIndex, depth
			{
				position280 := position
				depth++
				{
					switch buffer[position] {
					case '[':
						if !_rules[rulebrackets]() {
							goto l279
						}
						break
					case '{':
						if !_rules[rulebraces]() {
							goto l279
						}
						break
					default:
						if !_rules[ruleparens]() {
							goto l279
						}
						break
					}
				}

				depth--
				add(ruleenclosed, position280)
			}
			return true
		l279:
			position, tokenIndex, depth = position279, tokenIndex279, depth279
			return false
		},
		/* 25 parens <- <('(' inner ')')> */
		func() bool {
			position282, tokenIndex282, depth282 := position, tokenIndex, depth
			{
				position283 := position
				depth++
				if buffer[position] != rune('(') {
					goto l282
				}
				position++
				if !_rules[ruleinner]() {
					goto l282
				}
				if buffer[position] != rune(')') {
					goto l282
				}
				position++
				depth--
				add(ruleparens, position283)
			}
			return true
		l282:
			position, tokenIndex, depth = position282, tokenIndex282, depth282
			return false
		},
		/* 26 braces <- <('{' inner '}')> */
		func() bool {
			position284, tokenIndex284, depth284 := position, tokenIndex, depth
			{
				position285 := position
				depth++
				if buffer[position] != rune('{') {
					goto l284
				}
				position++
				if !_rules[ruleinner]() {
					goto l284
				}
				if buffer[position] != rune('}') {
					goto l284
				}
				position++
				depth--
				add(rulebraces, position285)
			}
			return true
		l284:
			position, tokenIndex, depth = position284, tokenIndex284, depth284
			return false
		},
		/* 27 brackets <- <('[' inner ']')> */
		func() bool {
			position286, tokenIndex286, depth286 := position, tokenIndex, depth
			{
				position287 := position
				depth++
				if buffer[position] != rune('[') {
					goto l286
				}
				position++
				if !_rules[ruleinner]() {
					goto l286
				}
				if buffer[position] != rune(']') {
					goto l286
				}
				position++
				depth--
				add(rulebrackets, position287)
			}
			return true
		l286:
			position, tokenIndex, depth = position286, tokenIndex286, depth286
			return false
		},
		/* 28 inner <- <((&('\t' | ' ') isp+) | (&(',') ',') | (&('(' | '[' | '{') enclosed) | (&('!' | '"' | '&' | '*' | '+' | '-' | '.' | '/' | '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9' | ':' | '<' | '=' | '>' | 'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '^' | '_' | '`' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z' | '|') commaless))*> */
		func() bool {
			{
				position289 := position
				depth++
			l290:
				{
					position291, tokenIndex291, depth291 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '\t', ' ':
							if !_rules[ruleisp]() {
								goto l291
							}
						l293:
							{
								position294, tokenIndex294, depth294 := position, tokenIndex, depth
								if !_rules[ruleisp]() {
									goto l294
								}
								goto l293
							l294:
								position, tokenIndex, depth = position294, tokenIndex294, depth294
							}
							break
						case ',':
							if buffer[position] != rune(',') {
								goto l291
							}
							position++
							break
						case '(', '[', '{':
							if !_rules[ruleenclosed]() {
								goto l291
							}
							break
						default:
							if !_rules[rulecommaless]() {
								goto l291
							}
							break
						}
					}

					goto l290
				l291:
					position, tokenIndex, depth = position291, tokenIndex291, depth291
				}
				depth--
				add(ruleinner, position289)
			}
			return true
		},
		/* 29 identifier <- <(((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z])) ((&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') ([0-9] / [0-9])) | (&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))*)> */
		func() bool {
			position295, tokenIndex295, depth295 := position, tokenIndex, depth
			{
				position296 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l295
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l295
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l295
						}
						position++
						break
					}
				}

			l298:
				{
					position299, tokenIndex299, depth299 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
							{
								position301, tokenIndex301, depth301 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l302
								}
								position++
								goto l301
							l302:
								position, tokenIndex, depth = position301, tokenIndex301, depth301
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l299
								}
								position++
							}
						l301:
							break
						case '_':
							if buffer[position] != rune('_') {
								goto l299
							}
							position++
							break
						case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l299
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l299
							}
							position++
							break
						}
					}

					goto l298
				l299:
					position, tokenIndex, depth = position299, tokenIndex299, depth299
				}
				depth--
				add(ruleidentifier, position296)
			}
			return true
		l295:
			position, tokenIndex, depth = position295, tokenIndex295, depth295
			return false
		},
		/* 30 fields <- <(((&('\n') '\n') | (&('\t') '\t') | (&(' ') ' ') | (&(';') ';'))* field isp* (fsep isp* (fsep isp*)* field)* ((&('\n') '\n') | (&('\t') '\t') | (&(' ') ' ') | (&(';') ';'))* !.)> */
		func() bool {
			position303, tokenIndex303, depth303 := position, tokenIndex, depth
			{
				position304 := position
				depth++
			l305:
				{
					position306, tokenIndex306, depth306 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '\n':
							if buffer[position] != rune('\n') {
								goto l306
							}
							position++
							break
						case '\t':
							if buffer[position] != rune('\t') {
								goto l306
							}
							position++
							break
						case ' ':
							if buffer[position] != rune(' ') {
								goto l306
							}
							position++
							break
						default:
							if buffer[position] != rune(';') {
								goto l306
							}
							position++
							break
						}
					}

					goto l305
				l306:
					position, tokenIndex, depth = position306, tokenIndex306, depth306
				}
				if !_rules[rulefield]() {
					goto l303
				}
			l308:
				{
					position309, tokenIndex309, depth309 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l309
					}
					goto l308
				l309:
					position, tokenIndex, depth = position309, tokenIndex309, depth309
				}
			l310:
				{
					position311, tokenIndex311, depth311 := position, tokenIndex, depth
					if !_rules[rulefsep]() {
						goto l311
					}
				l312:
					{
						position313, tokenIndex313, depth313 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l313
						}
						goto l312
					l313:
						position, tokenIndex, depth = position313, tokenIndex313, depth313
					}
				l314:
					{
						position315, tokenIndex315, depth315 := position, tokenIndex, depth
						if !_rules[rulefsep]() {
							goto l315
						}
					l316:
						{
							position317, tokenIndex317, depth317 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l317
							}
							goto l316
						l317:
							position, tokenIndex, depth = position317, tokenIndex317, depth317
						}
						goto l314
					l315:
						position, tokenIndex, depth = position315, tokenIndex315, depth315
					}
					if !_rules[rulefield]() {
						goto l311
					}
					goto l310
				l311:
					position, tokenIndex, depth = position311, tokenIndex311, depth311
				}
			l318:
				{
					position319, tokenIndex319, depth319 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '\n':
							if buffer[position] != rune('\n') {
								goto l319
							}
							position++
							break
						case '\t':
							if buffer[position] != rune('\t') {
								goto l319
							}
							position++
							break
						case ' ':
							if buffer[position] != rune(' ') {
								goto l319
							}
							position++
							break
						default:
							if buffer[position] != rune(';') {
								goto l319
							}
							position++
							break
						}
					}

					goto l318
				l319:
					position, tokenIndex, depth = position319, tokenIndex319, depth319
				}
				{
					position321, tokenIndex321, depth321 := position, tokenIndex, depth
					if !matchDot() {
						goto l321
					}
					goto l303
				l321:
					position, tokenIndex, depth = position321, tokenIndex321, depth321
				}
				depth--
				add(rulefields, position304)
			}
			return true
		l303:
			position, tokenIndex, depth = position303, tokenIndex303, depth303
			return false
		},
		/* 31 fsep <- <(';' / '\n')> */
		func() bool {
			position322, tokenIndex322, depth322 := position, tokenIndex, depth
			{
				position323 := position
				depth++
				{
					position324, tokenIndex324, depth324 := position, tokenIndex, depth
					if buffer[position] != rune(';') {
						goto l325
					}
					position++
					goto l324
				l325:
					position, tokenIndex, depth = position324, tokenIndex324, depth324
					if buffer[position] != rune('\n') {
						goto l322
					}
					position++
				}
			l324:
				depth--
				add(rulefsep, position323)
			}
			return true
		l322:
			position, tokenIndex, depth = position322, tokenIndex322, depth322
			return false
		},
		/* 32 field <- <(name (isp* ',' isp* name)* isp+ type isp* ('=' isp* expr)? Action15)> */
		func() bool {
			position326, tokenIndex326, depth326 := position, tokenIndex, depth
			{
				position327 := position
				depth++
				if !_rules[rulename]() {
					goto l326
				}
			l328:
				{
					position329, tokenIndex329, depth329 := position, tokenIndex, depth
				l330:
					{
						position331, tokenIndex331, depth331 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l331
						}
						goto l330
					l331:
						position, tokenIndex, depth = position331, tokenIndex331, depth331
					}
					if buffer[position] != rune(',') {
						goto l329
					}
					position++
				l332:
					{
						position333, tokenIndex333, depth333 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l333
						}
						goto l332
					l333:
						position, tokenIndex, depth = position333, tokenIndex333, depth333
					}
					if !_rules[rulename]() {
						goto l329
					}
					goto l328
				l329:
					position, tokenIndex, depth = position329, tokenIndex329, depth329
				}
				if !_rules[ruleisp]() {
					goto l326
				}
			l334:
				{
					position335, tokenIndex335, depth335 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l335
					}
					goto l334
				l335:
					position, tokenIndex, depth = position335, tokenIndex335, depth335
				}
				if !_rules[ruletype]() {
					goto l326
				}
			l336:
				{
					position337, tokenIndex337, depth337 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l337
					}
					goto l336
				l337:
					position, tokenIndex, depth = position337, tokenIndex337, depth337
				}
				{
					position338, tokenIndex338, depth338 := position, tokenIndex, depth
					if buffer[position] != rune('=') {
						goto l338
					}
					position++
				l340:
					{
						position341, tokenIndex341, depth341 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l341
						}
						goto l340
					l341:
						position, tokenIndex, depth = position341, tokenIndex341, depth341
					}
					if !_rules[ruleexpr]() {
						goto l338
					}
					goto l339
				l338:
					position, tokenIndex, depth = position338, tokenIndex338, depth338
				}
			l339:
				if !_rules[ruleAction15]() {
					goto l326
				}
				depth--
				add(rulefield, position327)
			}
			return true
		l326:
			position, tokenIndex, depth = position326, tokenIndex326, depth326
			return false
		},
		/* 33 name <- <(<((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))+> Action16)> */
		func() bool {
			position342, tokenIndex342, depth342 := position, tokenIndex, depth
			{
				position343 := position
				depth++
				{
					position344 := position
					depth++
					{
						switch buffer[position] {
						case '_':
							if buffer[position] != rune('_') {
								goto l342
							}
							position++
							break
						case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l342
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l342
							}
							position++
							break
						}
					}

				l345:
					{
						position346, tokenIndex346, depth346 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '_':
								if buffer[position] != rune('_') {
									goto l346
								}
								position++
								break
							case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
								if c := buffer[position]; c < rune('A') || c > rune('Z') {
									goto l346
								}
								position++
								break
							default:
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l346
								}
								position++
								break
							}
						}

						goto l345
					l346:
						position, tokenIndex, depth = position346, tokenIndex346, depth346
					}
					depth--
					add(rulePegText, position344)
				}
				if !_rules[ruleAction16]() {
					goto l342
				}
				depth--
				add(rulename, position343)
			}
			return true
		l342:
			position, tokenIndex, depth = position342, tokenIndex342, depth342
			return false
		},
		/* 34 type <- <(chan / func / qname / sname / ((&('*') pointer) | (&('[') array) | (&('M' | 'm') map)))> */
		func() bool {
			position349, tokenIndex349, depth349 := position, tokenIndex, depth
			{
				position350 := position
				depth++
				{
					position351, tokenIndex351, depth351 := position, tokenIndex, depth
					if !_rules[rulechan]() {
						goto l352
					}
					goto l351
				l352:
					position, tokenIndex, depth = position351, tokenIndex351, depth351
					if !_rules[rulefunc]() {
						goto l353
					}
					goto l351
				l353:
					position, tokenIndex, depth = position351, tokenIndex351, depth351
					if !_rules[ruleqname]() {
						goto l354
					}
					goto l351
				l354:
					position, tokenIndex, depth = position351, tokenIndex351, depth351
					if !_rules[rulesname]() {
						goto l355
					}
					goto l351
				l355:
					position, tokenIndex, depth = position351, tokenIndex351, depth351
					{
						switch buffer[position] {
						case '*':
							if !_rules[rulepointer]() {
								goto l349
							}
							break
						case '[':
							if !_rules[rulearray]() {
								goto l349
							}
							break
						default:
							if !_rules[rulemap]() {
								goto l349
							}
							break
						}
					}

				}
			l351:
				depth--
				add(ruletype, position350)
			}
			return true
		l349:
			position, tokenIndex, depth = position349, tokenIndex349, depth349
			return false
		},
		/* 35 sname <- <(<((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))+> Action17)> */
		func() bool {
			position357, tokenIndex357, depth357 := position, tokenIndex, depth
			{
				position358 := position
				depth++
				{
					position359 := position
					depth++
					{
						switch buffer[position] {
						case '_':
							if buffer[position] != rune('_') {
								goto l357
							}
							position++
							break
						case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l357
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l357
							}
							position++
							break
						}
					}

				l360:
					{
						position361, tokenIndex361, depth361 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '_':
								if buffer[position] != rune('_') {
									goto l361
								}
								position++
								break
							case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
								if c := buffer[position]; c < rune('A') || c > rune('Z') {
									goto l361
								}
								position++
								break
							default:
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l361
								}
								position++
								break
							}
						}

						goto l360
					l361:
						position, tokenIndex, depth = position361, tokenIndex361, depth361
					}
					depth--
					add(rulePegText, position359)
				}
				if !_rules[ruleAction17]() {
					goto l357
				}
				depth--
				add(rulesname, position358)
			}
			return true
		l357:
			position, tokenIndex, depth = position357, tokenIndex357, depth357
			return false
		},
		/* 36 qname <- <(<(((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))+ '.' ((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))+)> Action18)> */
		func() bool {
			position364, tokenIndex364, depth364 := position, tokenIndex, depth
			{
				position365 := position
				depth++
				{
					position366 := position
					depth++
					{
						switch buffer[position] {
						case '_':
							if buffer[position] != rune('_') {
								goto l364
							}
							position++
							break
						case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l364
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l364
							}
							position++
							break
						}
					}

				l367:
					{
						position368, tokenIndex368, depth368 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '_':
								if buffer[position] != rune('_') {
									goto l368
								}
								position++
								break
							case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
								if c := buffer[position]; c < rune('A') || c > rune('Z') {
									goto l368
								}
								position++
								break
							default:
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l368
								}
								position++
								break
							}
						}

						goto l367
					l368:
						position, tokenIndex, depth = position368, tokenIndex368, depth368
					}
					if buffer[position] != rune('.') {
						goto l364
					}
					position++
					{
						switch buffer[position] {
						case '_':
							if buffer[position] != rune('_') {
								goto l364
							}
							position++
							break
						case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l364
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l364
							}
							position++
							break
						}
					}

				l371:
					{
						position372, tokenIndex372, depth372 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '_':
								if buffer[position] != rune('_') {
									goto l372
								}
								position++
								break
							case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
								if c := buffer[position]; c < rune('A') || c > rune('Z') {
									goto l372
								}
								position++
								break
							default:
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l372
								}
								position++
								break
							}
						}

						goto l371
					l372:
						position, tokenIndex, depth = position372, tokenIndex372, depth372
					}
					depth--
					add(rulePegText, position366)
				}
				if !_rules[ruleAction18]() {
					goto l364
				}
				depth--
				add(ruleqname, position365)
			}
			return true
		l364:
			position, tokenIndex, depth = position364, tokenIndex364, depth364
			return false
		},
		/* 37 array <- <('[' ']' type Action19)> */
		func() bool {
			position375, tokenIndex375, depth375 := position, tokenIndex, depth
			{
				position376 := position
				depth++
				if buffer[position] != rune('[') {
					goto l375
				}
				position++
				if buffer[position] != rune(']') {
					goto l375
				}
				position++
				if !_rules[ruletype]() {
					goto l375
				}
				if !_rules[ruleAction19]() {
					goto l375
				}
				depth--
				add(rulearray, position376)
			}
			return true
		l375:
			position, tokenIndex, depth = position375, tokenIndex375, depth375
			return false
		},
		/* 38 map <- <(('m' / 'M') ('a' / 'A') ('p' / 'P') '[' isp* keytype isp* ']' type Action20)> */
		func() bool {
			position377, tokenIndex377, depth377 := position, tokenIndex, depth
			{
				position378 := position
				depth++
				{
					position379, tokenIndex379, depth379 := position, tokenIndex, depth
					if buffer[position] != rune('m') {
						goto l380
					}
					position++
					goto l379
				l380:
					position, tokenIndex, depth = position379, tokenIndex379, depth379
					if buffer[position] != rune('M') {
						goto l377
					}
					position++
				}
			l379:
				{
					position381, tokenIndex381, depth381 := position, tokenIndex, depth
					if buffer[position] != rune('a') {
						goto l382
					}
					position++
					goto l381
				l382:
					position, tokenIndex, depth = position381, tokenIndex381, depth381
					if buffer[position] != rune('A') {
						goto l377
					}
					position++
				}
			l381:
				{
					position383, tokenIndex383, depth383 := position, tokenIndex, depth
					if buffer[position] != rune('p') {
						goto l384
					}
					position++
					goto l383
				l384:
					position, tokenIndex, depth = position383, tokenIndex383, depth383
					if buffer[position] != rune('P') {
						goto l377
					}
					position++
				}
			l383:
				if buffer[position] != rune('[') {
					goto l377
				}
				position++
			l385:
				{
					position386, tokenIndex386, depth386 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l386
					}
					goto l385
				l386:
					position, tokenIndex, depth = position386, tokenIndex386, depth386
				}
				if !_rules[rulekeytype]() {
					goto l377
				}
			l387:
				{
					position388, tokenIndex388, depth388 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l388
					}
					goto l387
				l388:
					position, tokenIndex, depth = position388, tokenIndex388, depth388
				}
				if buffer[position] != rune(']') {
					goto l377
				}
				position++
				if !_rules[ruletype]() {
					goto l377
				}
				if !_rules[ruleAction20]() {
					goto l377
				}
				depth--
				add(rulemap, position378)
			}
			return true
		l377:
			position, tokenIndex, depth = position377, tokenIndex377, depth377
			return false
		},
		/* 39 chan <- <(('c' / 'C') ('h' / 'H') ('a' / 'A') ('n' / 'N') isp+ type Action21)> */
		func() bool {
			position389, tokenIndex389, depth389 := position, tokenIndex, depth
			{
				position390 := position
				depth++
				{
					position391, tokenIndex391, depth391 := position, tokenIndex, depth
					if buffer[position] != rune('c') {
						goto l392
					}
					position++
					goto l391
				l392:
					position, tokenIndex, depth = position391, tokenIndex391, depth391
					if buffer[position] != rune('C') {
						goto l389
					}
					position++
				}
			l391:
				{
					position393, tokenIndex393, depth393 := position, tokenIndex, depth
					if buffer[position] != rune('h') {
						goto l394
					}
					position++
					goto l393
				l394:
					position, tokenIndex, depth = position393, tokenIndex393, depth393
					if buffer[position] != rune('H') {
						goto l389
					}
					position++
				}
			l393:
				{
					position395, tokenIndex395, depth395 := position, tokenIndex, depth
					if buffer[position] != rune('a') {
						goto l396
					}
					position++
					goto l395
				l396:
					position, tokenIndex, depth = position395, tokenIndex395, depth395
					if buffer[position] != rune('A') {
						goto l389
					}
					position++
				}
			l395:
				{
					position397, tokenIndex397, depth397 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l398
					}
					position++
					goto l397
				l398:
					position, tokenIndex, depth = position397, tokenIndex397, depth397
					if buffer[position] != rune('N') {
						goto l389
					}
					position++
				}
			l397:
				if !_rules[ruleisp]() {
					goto l389
				}
			l399:
				{
					position400, tokenIndex400, depth400 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l400
					}
					goto l399
				l400:
					position, tokenIndex, depth = position400, tokenIndex400, depth400
				}
				if !_rules[ruletype]() {
					goto l389
				}
				if !_rules[ruleAction21]() {
					goto l389
				}
				depth--
				add(rulechan, position390)
			}
			return true
		l389:
			position, tokenIndex, depth = position389, tokenIndex389, depth389
			return false
		},
		/* 40 func <- <(('f' / 'F') ('u' / 'U') ('n' / 'N') ('c' / 'C') isp* '(' isp* (param isp* (',' isp* param)*)? ')' isp* type? Action22)> */
		func() bool {
			position401, tokenIndex401, depth401 := position, tokenIndex, depth
			{
				position402 := position
				depth++
				{
					position403, tokenIndex403, depth403 := position, tokenIndex, depth
					if buffer[position] != rune('f') {
						goto l404
					}
					position++
					goto l403
				l404:
					position, tokenIndex, depth = position403, tokenIndex403, depth403
					if buffer[position] != rune('F') {
						goto l401
					}
					position++
				}
			l403:
				{
					position405, tokenIndex405, depth405 := position, tokenIndex, depth
					if buffer[position] != rune('u') {
						goto l406
					}
					position++
					goto l405
				l406:
					position, tokenIndex, depth = position405, tokenIndex405, depth405
					if buffer[position] != rune('U') {
						goto l401
					}
					position++
				}
			l405:
				{
					position407, tokenIndex407, depth407 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l408
					}
					position++
					goto l407
				l408:
					position, tokenIndex, depth = position407, tokenIndex407, depth407
					if buffer[position] != rune('N') {
						goto l401
					}
					position++
				}
			l407:
				{
					position409, tokenIndex409, depth409 := position, tokenIndex, depth
					if buffer[position] != rune('c') {
						goto l410
					}
					position++
					goto l409
				l410:
					position, tokenIndex, depth = position409, tokenIndex409, depth409
					if buffer[position] != rune('C') {
						goto l401
					}
					position++
				}
			l409:
			l411:
				{
					position412, tokenIndex412, depth412 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l412
					}
					goto l411
				l412:
					position, tokenIndex, depth = position412, tokenIndex412, depth412
				}
				if buffer[position] != rune('(') {
					goto l401
				}
				position++
			l413:
				{
					position414, tokenIndex414, depth414 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l414
					}
					goto l413
				l414:
					position, tokenIndex, depth = position414, tokenIndex414, depth414
				}
				{
					position415, tokenIndex415, depth415 := position, tokenIndex, depth
					if !_rules[ruleparam]() {
						goto l415
					}
				l417:
					{
						position418, tokenIndex418, depth418 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l418
						}
						goto l417
					l418:
						position, tokenIndex, depth = position418, tokenIndex418, depth418
					}
				l419:
					{
						position420, tokenIndex420, depth420 := position, tokenIndex, depth
						if buffer[position] != rune(',') {
							goto l420
						}
						position++
					l421:
						{
							position422, tokenIndex422, depth422 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l422
							}
							goto l421
						l422:
							position, tokenIndex, depth = position422, tokenIndex422, depth422
						}
						if !_rules[ruleparam]() {
							goto l420
						}
						goto l419
					l420:
						position, tokenIndex, depth = position420, tokenIndex420, depth420
					}
					goto l416
				l415:
					position, tokenIndex, depth = position415, tokenIndex415, depth415
				}
			l416:
				if buffer[position] != rune(')') {
					goto l401
				}
				position++
			l423:
				{
					position424, tokenIndex424, depth424 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l424
					}
					goto l423
				l424:
					position, tokenIndex, depth = position424, tokenIndex424, depth424
				}
				{
					position425, tokenIndex425, depth425 := position, tokenIndex, depth
					if !_rules[ruletype]() {
						goto l425
					}
					goto l426
				l425:
					position, tokenIndex, depth = position425, tokenIndex425, depth425
				}
			l426:
				if !_rules[ruleAction22]() {
					goto l401
				}
				depth--
				add(rulefunc, position402)
			}
			return true
		l401:
			position, tokenIndex, depth = position401, tokenIndex401, depth401
			return false
		},
		/* 41 keytype <- <(type Action23)> */
		func() bool {
			position427, tokenIndex427, depth427 := position, tokenIndex, depth
			{
				position428 := position
				depth++
				if !_rules[ruletype]() {
					goto l427
				}
				if !_rules[ruleAction23]() {
					goto l427
				}
				depth--
				add(rulekeytype, position428)
			}
			return true
		l427:
			position, tokenIndex, depth = position427, tokenIndex427, depth427
			return false
		},
		/* 42 pointer <- <('*' type Action24)> */
		func() bool {
			position429, tokenIndex429, depth429 := position, tokenIndex, depth
			{
				position430 := position
				depth++
				if buffer[position] != rune('*') {
					goto l429
				}
				position++
				if !_rules[ruletype]() {
					goto l429
				}
				if !_rules[ruleAction24]() {
					goto l429
				}
				depth--
				add(rulepointer, position430)
			}
			return true
		l429:
			position, tokenIndex, depth = position429, tokenIndex429, depth429
			return false
		},
		/* 43 captures <- <(isp* capture isp* (',' isp* capture isp*)* !.)> */
		func() bool {
			position431, tokenIndex431, depth431 := position, tokenIndex, depth
			{
				position432 := position
				depth++
			l433:
				{
					position434, tokenIndex434, depth434 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l434
					}
					goto l433
				l434:
					position, tokenIndex, depth = position434, tokenIndex434, depth434
				}
				if !_rules[rulecapture]() {
					goto l431
				}
			l435:
				{
					position436, tokenIndex436, depth436 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l436
					}
					goto l435
				l436:
					position, tokenIndex, depth = position436, tokenIndex436, depth436
				}
			l437:
				{
					position438, tokenIndex438, depth438 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l438
					}
					position++
				l439:
					{
						position440, tokenIndex440, depth440 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l440
						}
						goto l439
					l440:
						position, tokenIndex, depth = position440, tokenIndex440, depth440
					}
					if !_rules[rulecapture]() {
						goto l438
					}
				l441:
					{
						position442, tokenIndex442, depth442 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l442
						}
						goto l441
					l442:
						position, tokenIndex, depth = position442, tokenIndex442, depth442
					}
					goto l437
				l438:
					position, tokenIndex, depth = position438, tokenIndex438, depth438
				}
				{
					position443, tokenIndex443, depth443 := position, tokenIndex, depth
					if !matchDot() {
						goto l443
					}
					goto l431
				l443:
					position, tokenIndex, depth = position443, tokenIndex443, depth443
				}
				depth--
				add(rulecaptures, position432)
			}
			return true
		l431:
			position, tokenIndex, depth = position431, tokenIndex431, depth431
			return false
		},
		/* 44 capture <- <(eventid isp* ':' handlername isp* mappings isp* tags Action25)> */
		func() bool {
			position444, tokenIndex444, depth444 := position, tokenIndex, depth
			{
				position445 := position
				depth++
				if !_rules[ruleeventid]() {
					goto l444
				}
			l446:
				{
					position447, tokenIndex447, depth447 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l447
					}
					goto l446
				l447:
					position, tokenIndex, depth = position447, tokenIndex447, depth447
				}
				if buffer[position] != rune(':') {
					goto l444
				}
				position++
				if !_rules[rulehandlername]() {
					goto l444
				}
			l448:
				{
					position449, tokenIndex449, depth449 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l449
					}
					goto l448
				l449:
					position, tokenIndex, depth = position449, tokenIndex449, depth449
				}
				if !_rules[rulemappings]() {
					goto l444
				}
			l450:
				{
					position451, tokenIndex451, depth451 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l451
					}
					goto l450
				l451:
					position, tokenIndex, depth = position451, tokenIndex451, depth451
				}
				if !_rules[ruletags]() {
					goto l444
				}
				if !_rules[ruleAction25]() {
					goto l444
				}
				depth--
				add(rulecapture, position445)
			}
			return true
		l444:
			position, tokenIndex, depth = position444, tokenIndex444, depth444
			return false
		},
		/* 45 handlername <- <(<identifier> Action26)> */
		func() bool {
			position452, tokenIndex452, depth452 := position, tokenIndex, depth
			{
				position453 := position
				depth++
				{
					position454 := position
					depth++
					if !_rules[ruleidentifier]() {
						goto l452
					}
					depth--
					add(rulePegText, position454)
				}
				if !_rules[ruleAction26]() {
					goto l452
				}
				depth--
				add(rulehandlername, position453)
			}
			return true
		l452:
			position, tokenIndex, depth = position452, tokenIndex452, depth452
			return false
		},
		/* 46 eventid <- <(<[a-z]+> Action27)> */
		func() bool {
			position455, tokenIndex455, depth455 := position, tokenIndex, depth
			{
				position456 := position
				depth++
				{
					position457 := position
					depth++
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l455
					}
					position++
				l458:
					{
						position459, tokenIndex459, depth459 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l459
						}
						position++
						goto l458
					l459:
						position, tokenIndex, depth = position459, tokenIndex459, depth459
					}
					depth--
					add(rulePegText, position457)
				}
				if !_rules[ruleAction27]() {
					goto l455
				}
				depth--
				add(ruleeventid, position456)
			}
			return true
		l455:
			position, tokenIndex, depth = position455, tokenIndex455, depth455
			return false
		},
		/* 47 mappings <- <(mappingstart (isp* mapping isp* (',' isp* mapping isp*)*)? ')')?> */
		func() bool {
			{
				position461 := position
				depth++
				{
					position462, tokenIndex462, depth462 := position, tokenIndex, depth
					if !_rules[rulemappingstart]() {
						goto l462
					}
					{
						position464, tokenIndex464, depth464 := position, tokenIndex, depth
					l466:
						{
							position467, tokenIndex467, depth467 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l467
							}
							goto l466
						l467:
							position, tokenIndex, depth = position467, tokenIndex467, depth467
						}
						if !_rules[rulemapping]() {
							goto l464
						}
					l468:
						{
							position469, tokenIndex469, depth469 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l469
							}
							goto l468
						l469:
							position, tokenIndex, depth = position469, tokenIndex469, depth469
						}
					l470:
						{
							position471, tokenIndex471, depth471 := position, tokenIndex, depth
							if buffer[position] != rune(',') {
								goto l471
							}
							position++
						l472:
							{
								position473, tokenIndex473, depth473 := position, tokenIndex, depth
								if !_rules[ruleisp]() {
									goto l473
								}
								goto l472
							l473:
								position, tokenIndex, depth = position473, tokenIndex473, depth473
							}
							if !_rules[rulemapping]() {
								goto l471
							}
						l474:
							{
								position475, tokenIndex475, depth475 := position, tokenIndex, depth
								if !_rules[ruleisp]() {
									goto l475
								}
								goto l474
							l475:
								position, tokenIndex, depth = position475, tokenIndex475, depth475
							}
							goto l470
						l471:
							position, tokenIndex, depth = position471, tokenIndex471, depth471
						}
						goto l465
					l464:
						position, tokenIndex, depth = position464, tokenIndex464, depth464
					}
				l465:
					if buffer[position] != rune(')') {
						goto l462
					}
					position++
					goto l463
				l462:
					position, tokenIndex, depth = position462, tokenIndex462, depth462
				}
			l463:
				depth--
				add(rulemappings, position461)
			}
			return true
		},
		/* 48 mappingstart <- <('(' Action28)> */
		func() bool {
			position476, tokenIndex476, depth476 := position, tokenIndex, depth
			{
				position477 := position
				depth++
				if buffer[position] != rune('(') {
					goto l476
				}
				position++
				if !_rules[ruleAction28]() {
					goto l476
				}
				depth--
				add(rulemappingstart, position477)
			}
			return true
		l476:
			position, tokenIndex, depth = position476, tokenIndex476, depth476
			return false
		},
		/* 49 mapping <- <((mappingname isp* '=' isp*)? bound Action29)> */
		func() bool {
			position478, tokenIndex478, depth478 := position, tokenIndex, depth
			{
				position479 := position
				depth++
				{
					position480, tokenIndex480, depth480 := position, tokenIndex, depth
					if !_rules[rulemappingname]() {
						goto l480
					}
				l482:
					{
						position483, tokenIndex483, depth483 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l483
						}
						goto l482
					l483:
						position, tokenIndex, depth = position483, tokenIndex483, depth483
					}
					if buffer[position] != rune('=') {
						goto l480
					}
					position++
				l484:
					{
						position485, tokenIndex485, depth485 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l485
						}
						goto l484
					l485:
						position, tokenIndex, depth = position485, tokenIndex485, depth485
					}
					goto l481
				l480:
					position, tokenIndex, depth = position480, tokenIndex480, depth480
				}
			l481:
				if !_rules[rulebound]() {
					goto l478
				}
				if !_rules[ruleAction29]() {
					goto l478
				}
				depth--
				add(rulemapping, position479)
			}
			return true
		l478:
			position, tokenIndex, depth = position478, tokenIndex478, depth478
			return false
		},
		/* 50 mappingname <- <(<identifier> Action30)> */
		func() bool {
			position486, tokenIndex486, depth486 := position, tokenIndex, depth
			{
				position487 := position
				depth++
				{
					position488 := position
					depth++
					if !_rules[ruleidentifier]() {
						goto l486
					}
					depth--
					add(rulePegText, position488)
				}
				if !_rules[ruleAction30]() {
					goto l486
				}
				depth--
				add(rulemappingname, position487)
			}
			return true
		l486:
			position, tokenIndex, depth = position486, tokenIndex486, depth486
			return false
		},
		/* 51 tags <- <('{' isp* tag isp* (',' isp* tag isp*)* '}')?> */
		func() bool {
			{
				position490 := position
				depth++
				{
					position491, tokenIndex491, depth491 := position, tokenIndex, depth
					if buffer[position] != rune('{') {
						goto l491
					}
					position++
				l493:
					{
						position494, tokenIndex494, depth494 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l494
						}
						goto l493
					l494:
						position, tokenIndex, depth = position494, tokenIndex494, depth494
					}
					if !_rules[ruletag]() {
						goto l491
					}
				l495:
					{
						position496, tokenIndex496, depth496 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l496
						}
						goto l495
					l496:
						position, tokenIndex, depth = position496, tokenIndex496, depth496
					}
				l497:
					{
						position498, tokenIndex498, depth498 := position, tokenIndex, depth
						if buffer[position] != rune(',') {
							goto l498
						}
						position++
					l499:
						{
							position500, tokenIndex500, depth500 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l500
							}
							goto l499
						l500:
							position, tokenIndex, depth = position500, tokenIndex500, depth500
						}
						if !_rules[ruletag]() {
							goto l498
						}
					l501:
						{
							position502, tokenIndex502, depth502 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l502
							}
							goto l501
						l502:
							position, tokenIndex, depth = position502, tokenIndex502, depth502
						}
						goto l497
					l498:
						position, tokenIndex, depth = position498, tokenIndex498, depth498
					}
					if buffer[position] != rune('}') {
						goto l491
					}
					position++
					goto l492
				l491:
					position, tokenIndex, depth = position491, tokenIndex491, depth491
				}
			l492:
				depth--
				add(ruletags, position490)
			}
			return true
		},
		/* 52 tag <- <(tagname ('(' (isp* tagarg isp* (',' isp* tagarg isp*)*)? ')')? Action31)> */
		func() bool {
			position503, tokenIndex503, depth503 := position, tokenIndex, depth
			{
				position504 := position
				depth++
				if !_rules[ruletagname]() {
					goto l503
				}
				{
					position505, tokenIndex505, depth505 := position, tokenIndex, depth
					if buffer[position] != rune('(') {
						goto l505
					}
					position++
					{
						position507, tokenIndex507, depth507 := position, tokenIndex, depth
					l509:
						{
							position510, tokenIndex510, depth510 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l510
							}
							goto l509
						l510:
							position, tokenIndex, depth = position510, tokenIndex510, depth510
						}
						if !_rules[ruletagarg]() {
							goto l507
						}
					l511:
						{
							position512, tokenIndex512, depth512 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l512
							}
							goto l511
						l512:
							position, tokenIndex, depth = position512, tokenIndex512, depth512
						}
					l513:
						{
							position514, tokenIndex514, depth514 := position, tokenIndex, depth
							if buffer[position] != rune(',') {
								goto l514
							}
							position++
						l515:
							{
								position516, tokenIndex516, depth516 := position, tokenIndex, depth
								if !_rules[ruleisp]() {
									goto l516
								}
								goto l515
							l516:
								position, tokenIndex, depth = position516, tokenIndex516, depth516
							}
							if !_rules[ruletagarg]() {
								goto l514
							}
						l517:
							{
								position518, tokenIndex518, depth518 := position, tokenIndex, depth
								if !_rules[ruleisp]() {
									goto l518
								}
								goto l517
							l518:
								position, tokenIndex, depth = position518, tokenIndex518, depth518
							}
							goto l513
						l514:
							position, tokenIndex, depth = position514, tokenIndex514, depth514
						}
						goto l508
					l507:
						position, tokenIndex, depth = position507, tokenIndex507, depth507
					}
				l508:
					if buffer[position] != rune(')') {
						goto l505
					}
					position++
					goto l506
				l505:
					position, tokenIndex, depth = position505, tokenIndex505, depth505
				}
			l506:
				if !_rules[ruleAction31]() {
					goto l503
				}
				depth--
				add(ruletag, position504)
			}
			return true
		l503:
			position, tokenIndex, depth = position503, tokenIndex503, depth503
			return false
		},
		/* 53 tagname <- <(<identifier> Action32)> */
		func() bool {
			position519, tokenIndex519, depth519 := position, tokenIndex, depth
			{
				position520 := position
				depth++
				{
					position521 := position
					depth++
					if !_rules[ruleidentifier]() {
						goto l519
					}
					depth--
					add(rulePegText, position521)
				}
				if !_rules[ruleAction32]() {
					goto l519
				}
				depth--
				add(ruletagname, position520)
			}
			return true
		l519:
			position, tokenIndex, depth = position519, tokenIndex519, depth519
			return false
		},
		/* 54 tagarg <- <(<identifier> Action33)> */
		func() bool {
			position522, tokenIndex522, depth522 := position, tokenIndex, depth
			{
				position523 := position
				depth++
				{
					position524 := position
					depth++
					if !_rules[ruleidentifier]() {
						goto l522
					}
					depth--
					add(rulePegText, position524)
				}
				if !_rules[ruleAction33]() {
					goto l522
				}
				depth--
				add(ruletagarg, position523)
			}
			return true
		l522:
			position, tokenIndex, depth = position522, tokenIndex522, depth522
			return false
		},
		/* 55 for <- <(isp* forVar isp* (',' isp* forVar isp*)? (':' '=') isp* (('r' / 'R') ('a' / 'A') ('n' / 'N') ('g' / 'G') ('e' / 'E')) isp+ expr isp* !.)> */
		func() bool {
			position525, tokenIndex525, depth525 := position, tokenIndex, depth
			{
				position526 := position
				depth++
			l527:
				{
					position528, tokenIndex528, depth528 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l528
					}
					goto l527
				l528:
					position, tokenIndex, depth = position528, tokenIndex528, depth528
				}
				if !_rules[ruleforVar]() {
					goto l525
				}
			l529:
				{
					position530, tokenIndex530, depth530 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l530
					}
					goto l529
				l530:
					position, tokenIndex, depth = position530, tokenIndex530, depth530
				}
				{
					position531, tokenIndex531, depth531 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l531
					}
					position++
				l533:
					{
						position534, tokenIndex534, depth534 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l534
						}
						goto l533
					l534:
						position, tokenIndex, depth = position534, tokenIndex534, depth534
					}
					if !_rules[ruleforVar]() {
						goto l531
					}
				l535:
					{
						position536, tokenIndex536, depth536 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l536
						}
						goto l535
					l536:
						position, tokenIndex, depth = position536, tokenIndex536, depth536
					}
					goto l532
				l531:
					position, tokenIndex, depth = position531, tokenIndex531, depth531
				}
			l532:
				if buffer[position] != rune(':') {
					goto l525
				}
				position++
				if buffer[position] != rune('=') {
					goto l525
				}
				position++
			l537:
				{
					position538, tokenIndex538, depth538 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l538
					}
					goto l537
				l538:
					position, tokenIndex, depth = position538, tokenIndex538, depth538
				}
				{
					position539, tokenIndex539, depth539 := position, tokenIndex, depth
					if buffer[position] != rune('r') {
						goto l540
					}
					position++
					goto l539
				l540:
					position, tokenIndex, depth = position539, tokenIndex539, depth539
					if buffer[position] != rune('R') {
						goto l525
					}
					position++
				}
			l539:
				{
					position541, tokenIndex541, depth541 := position, tokenIndex, depth
					if buffer[position] != rune('a') {
						goto l542
					}
					position++
					goto l541
				l542:
					position, tokenIndex, depth = position541, tokenIndex541, depth541
					if buffer[position] != rune('A') {
						goto l525
					}
					position++
				}
			l541:
				{
					position543, tokenIndex543, depth543 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l544
					}
					position++
					goto l543
				l544:
					position, tokenIndex, depth = position543, tokenIndex543, depth543
					if buffer[position] != rune('N') {
						goto l525
					}
					position++
				}
			l543:
				{
					position545, tokenIndex545, depth545 := position, tokenIndex, depth
					if buffer[position] != rune('g') {
						goto l546
					}
					position++
					goto l545
				l546:
					position, tokenIndex, depth = position545, tokenIndex545, depth545
					if buffer[position] != rune('G') {
						goto l525
					}
					position++
				}
			l545:
				{
					position547, tokenIndex547, depth547 := position, tokenIndex, depth
					if buffer[position] != rune('e') {
						goto l548
					}
					position++
					goto l547
				l548:
					position, tokenIndex, depth = position547, tokenIndex547, depth547
					if buffer[position] != rune('E') {
						goto l525
					}
					position++
				}
			l547:
				if !_rules[ruleisp]() {
					goto l525
				}
			l549:
				{
					position550, tokenIndex550, depth550 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l550
					}
					goto l549
				l550:
					position, tokenIndex, depth = position550, tokenIndex550, depth550
				}
				if !_rules[ruleexpr]() {
					goto l525
				}
			l551:
				{
					position552, tokenIndex552, depth552 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l552
					}
					goto l551
				l552:
					position, tokenIndex, depth = position552, tokenIndex552, depth552
				}
				{
					position553, tokenIndex553, depth553 := position, tokenIndex, depth
					if !matchDot() {
						goto l553
					}
					goto l525
				l553:
					position, tokenIndex, depth = position553, tokenIndex553, depth553
				}
				depth--
				add(rulefor, position526)
			}
			return true
		l525:
			position, tokenIndex, depth = position525, tokenIndex525, depth525
			return false
		},
		/* 56 forVar <- <(<identifier> Action34)> */
		func() bool {
			position554, tokenIndex554, depth554 := position, tokenIndex, depth
			{
				position555 := position
				depth++
				{
					position556 := position
					depth++
					if !_rules[ruleidentifier]() {
						goto l554
					}
					depth--
					add(rulePegText, position556)
				}
				if !_rules[ruleAction34]() {
					goto l554
				}
				depth--
				add(ruleforVar, position555)
			}
			return true
		l554:
			position, tokenIndex, depth = position554, tokenIndex554, depth554
			return false
		},
		/* 57 handlers <- <(isp* (fsep isp*)* handler isp* ((fsep isp*)+ handler isp*)* (fsep isp*)* !.)> */
		func() bool {
			position557, tokenIndex557, depth557 := position, tokenIndex, depth
			{
				position558 := position
				depth++
			l559:
				{
					position560, tokenIndex560, depth560 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l560
					}
					goto l559
				l560:
					position, tokenIndex, depth = position560, tokenIndex560, depth560
				}
			l561:
				{
					position562, tokenIndex562, depth562 := position, tokenIndex, depth
					if !_rules[rulefsep]() {
						goto l562
					}
				l563:
					{
						position564, tokenIndex564, depth564 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l564
						}
						goto l563
					l564:
						position, tokenIndex, depth = position564, tokenIndex564, depth564
					}
					goto l561
				l562:
					position, tokenIndex, depth = position562, tokenIndex562, depth562
				}
				if !_rules[rulehandler]() {
					goto l557
				}
			l565:
				{
					position566, tokenIndex566, depth566 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l566
					}
					goto l565
				l566:
					position, tokenIndex, depth = position566, tokenIndex566, depth566
				}
			l567:
				{
					position568, tokenIndex568, depth568 := position, tokenIndex, depth
					if !_rules[rulefsep]() {
						goto l568
					}
				l571:
					{
						position572, tokenIndex572, depth572 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l572
						}
						goto l571
					l572:
						position, tokenIndex, depth = position572, tokenIndex572, depth572
					}
				l569:
					{
						position570, tokenIndex570, depth570 := position, tokenIndex, depth
						if !_rules[rulefsep]() {
							goto l570
						}
					l573:
						{
							position574, tokenIndex574, depth574 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l574
							}
							goto l573
						l574:
							position, tokenIndex, depth = position574, tokenIndex574, depth574
						}
						goto l569
					l570:
						position, tokenIndex, depth = position570, tokenIndex570, depth570
					}
					if !_rules[rulehandler]() {
						goto l568
					}
				l575:
					{
						position576, tokenIndex576, depth576 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l576
						}
						goto l575
					l576:
						position, tokenIndex, depth = position576, tokenIndex576, depth576
					}
					goto l567
				l568:
					position, tokenIndex, depth = position568, tokenIndex568, depth568
				}
			l577:
				{
					position578, tokenIndex578, depth578 := position, tokenIndex, depth
					if !_rules[rulefsep]() {
						goto l578
					}
				l579:
					{
						position580, tokenIndex580, depth580 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l580
						}
						goto l579
					l580:
						position, tokenIndex, depth = position580, tokenIndex580, depth580
					}
					goto l577
				l578:
					position, tokenIndex, depth = position578, tokenIndex578, depth578
				}
				{
					position581, tokenIndex581, depth581 := position, tokenIndex, depth
					if !matchDot() {
						goto l581
					}
					goto l557
				l581:
					position, tokenIndex, depth = position581, tokenIndex581, depth581
				}
				depth--
				add(rulehandlers, position558)
			}
			return true
		l557:
			position, tokenIndex, depth = position557, tokenIndex557, depth557
			return false
		},
		/* 58 handler <- <(handlername '(' isp* (param isp* (',' isp* param isp*)*)? ')' (isp* type)? Action35)> */
		func() bool {
			position582, tokenIndex582, depth582 := position, tokenIndex, depth
			{
				position583 := position
				depth++
				if !_rules[rulehandlername]() {
					goto l582
				}
				if buffer[position] != rune('(') {
					goto l582
				}
				position++
			l584:
				{
					position585, tokenIndex585, depth585 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l585
					}
					goto l584
				l585:
					position, tokenIndex, depth = position585, tokenIndex585, depth585
				}
				{
					position586, tokenIndex586, depth586 := position, tokenIndex, depth
					if !_rules[ruleparam]() {
						goto l586
					}
				l588:
					{
						position589, tokenIndex589, depth589 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l589
						}
						goto l588
					l589:
						position, tokenIndex, depth = position589, tokenIndex589, depth589
					}
				l590:
					{
						position591, tokenIndex591, depth591 := position, tokenIndex, depth
						if buffer[position] != rune(',') {
							goto l591
						}
						position++
					l592:
						{
							position593, tokenIndex593, depth593 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l593
							}
							goto l592
						l593:
							position, tokenIndex, depth = position593, tokenIndex593, depth593
						}
						if !_rules[ruleparam]() {
							goto l591
						}
					l594:
						{
							position595, tokenIndex595, depth595 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l595
							}
							goto l594
						l595:
							position, tokenIndex, depth = position595, tokenIndex595, depth595
						}
						goto l590
					l591:
						position, tokenIndex, depth = position591, tokenIndex591, depth591
					}
					goto l587
				l586:
					position, tokenIndex, depth = position586, tokenIndex586, depth586
				}
			l587:
				if buffer[position] != rune(')') {
					goto l582
				}
				position++
				{
					position596, tokenIndex596, depth596 := position, tokenIndex, depth
				l598:
					{
						position599, tokenIndex599, depth599 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l599
						}
						goto l598
					l599:
						position, tokenIndex, depth = position599, tokenIndex599, depth599
					}
					if !_rules[ruletype]() {
						goto l596
					}
					goto l597
				l596:
					position, tokenIndex, depth = position596, tokenIndex596, depth596
				}
			l597:
				if !_rules[ruleAction35]() {
					goto l582
				}
				depth--
				add(rulehandler, position583)
			}
			return true
		l582:
			position, tokenIndex, depth = position582, tokenIndex582, depth582
			return false
		},
		/* 59 paramname <- <(<identifier> Action36)> */
		func() bool {
			position600, tokenIndex600, depth600 := position, tokenIndex, depth
			{
				position601 := position
				depth++
				{
					position602 := position
					depth++
					if !_rules[ruleidentifier]() {
						goto l600
					}
					depth--
					add(rulePegText, position602)
				}
				if !_rules[ruleAction36]() {
					goto l600
				}
				depth--
				add(ruleparamname, position601)
			}
			return true
		l600:
			position, tokenIndex, depth = position600, tokenIndex600, depth600
			return false
		},
		/* 60 param <- <(paramname isp+ type Action37)> */
		func() bool {
			position603, tokenIndex603, depth603 := position, tokenIndex, depth
			{
				position604 := position
				depth++
				if !_rules[ruleparamname]() {
					goto l603
				}
				if !_rules[ruleisp]() {
					goto l603
				}
			l605:
				{
					position606, tokenIndex606, depth606 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l606
					}
					goto l605
				l606:
					position, tokenIndex, depth = position606, tokenIndex606, depth606
				}
				if !_rules[ruletype]() {
					goto l603
				}
				if !_rules[ruleAction37]() {
					goto l603
				}
				depth--
				add(ruleparam, position604)
			}
			return true
		l603:
			position, tokenIndex, depth = position603, tokenIndex603, depth603
			return false
		},
		/* 61 cparams <- <(isp* (cparam isp* (',' isp* cparam isp*)*)? !.)> */
		func() bool {
			position607, tokenIndex607, depth607 := position, tokenIndex, depth
			{
				position608 := position
				depth++
			l609:
				{
					position610, tokenIndex610, depth610 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l610
					}
					goto l609
				l610:
					position, tokenIndex, depth = position610, tokenIndex610, depth610
				}
				{
					position611, tokenIndex611, depth611 := position, tokenIndex, depth
					if !_rules[rulecparam]() {
						goto l611
					}
				l613:
					{
						position614, tokenIndex614, depth614 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l614
						}
						goto l613
					l614:
						position, tokenIndex, depth = position614, tokenIndex614, depth614
					}
				l615:
					{
						position616, tokenIndex616, depth616 := position, tokenIndex, depth
						if buffer[position] != rune(',') {
							goto l616
						}
						position++
					l617:
						{
							position618, tokenIndex618, depth618 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l618
							}
							goto l617
						l618:
							position, tokenIndex, depth = position618, tokenIndex618, depth618
						}
						if !_rules[rulecparam]() {
							goto l616
						}
					l619:
						{
							position620, tokenIndex620, depth620 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l620
							}
							goto l619
						l620:
							position, tokenIndex, depth = position620, tokenIndex620, depth620
						}
						goto l615
					l616:
						position, tokenIndex, depth = position616, tokenIndex616, depth616
					}
					goto l612
				l611:
					position, tokenIndex, depth = position611, tokenIndex611, depth611
				}
			l612:
				{
					position621, tokenIndex621, depth621 := position, tokenIndex, depth
					if !matchDot() {
						goto l621
					}
					goto l607
				l621:
					position, tokenIndex, depth = position621, tokenIndex621, depth621
				}
				depth--
				add(rulecparams, position608)
			}
			return true
		l607:
			position, tokenIndex, depth = position607, tokenIndex607, depth607
			return false
		},
		/* 62 cparam <- <((var isp+)? tagname isp+ type Action38)> */
		func() bool {
			position622, tokenIndex622, depth622 := position, tokenIndex, depth
			{
				position623 := position
				depth++
				{
					position624, tokenIndex624, depth624 := position, tokenIndex, depth
					if !_rules[rulevar]() {
						goto l624
					}
					if !_rules[ruleisp]() {
						goto l624
					}
				l626:
					{
						position627, tokenIndex627, depth627 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l627
						}
						goto l626
					l627:
						position, tokenIndex, depth = position627, tokenIndex627, depth627
					}
					goto l625
				l624:
					position, tokenIndex, depth = position624, tokenIndex624, depth624
				}
			l625:
				if !_rules[ruletagname]() {
					goto l622
				}
				if !_rules[ruleisp]() {
					goto l622
				}
			l628:
				{
					position629, tokenIndex629, depth629 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l629
					}
					goto l628
				l629:
					position, tokenIndex, depth = position629, tokenIndex629, depth629
				}
				if !_rules[ruletype]() {
					goto l622
				}
				if !_rules[ruleAction38]() {
					goto l622
				}
				depth--
				add(rulecparam, position623)
			}
			return true
		l622:
			position, tokenIndex, depth = position622, tokenIndex622, depth622
			return false
		},
		/* 63 var <- <(('v' / 'V') ('a' / 'A') ('r' / 'R') Action39)> */
		func() bool {
			position630, tokenIndex630, depth630 := position, tokenIndex, depth
			{
				position631 := position
				depth++
				{
					position632, tokenIndex632, depth632 := position, tokenIndex, depth
					if buffer[position] != rune('v') {
						goto l633
					}
					position++
					goto l632
				l633:
					position, tokenIndex, depth = position632, tokenIndex632, depth632
					if buffer[position] != rune('V') {
						goto l630
					}
					position++
				}
			l632:
				{
					position634, tokenIndex634, depth634 := position, tokenIndex, depth
					if buffer[position] != rune('a') {
						goto l635
					}
					position++
					goto l634
				l635:
					position, tokenIndex, depth = position634, tokenIndex634, depth634
					if buffer[position] != rune('A') {
						goto l630
					}
					position++
				}
			l634:
				{
					position636, tokenIndex636, depth636 := position, tokenIndex, depth
					if buffer[position] != rune('r') {
						goto l637
					}
					position++
					goto l636
				l637:
					position, tokenIndex, depth = position636, tokenIndex636, depth636
					if buffer[position] != rune('R') {
						goto l630
					}
					position++
				}
			l636:
				if !_rules[ruleAction39]() {
					goto l630
				}
				depth--
				add(rulevar, position631)
			}
			return true
		l630:
			position, tokenIndex, depth = position630, tokenIndex630, depth630
			return false
		},
		/* 64 args <- <(isp* arg isp* (',' isp* arg isp*)* !.)> */
		func() bool {
			position638, tokenIndex638, depth638 := position, tokenIndex, depth
			{
				position639 := position
				depth++
			l640:
				{
					position641, tokenIndex641, depth641 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l641
					}
					goto l640
				l641:
					position, tokenIndex, depth = position641, tokenIndex641, depth641
				}
				if !_rules[rulearg]() {
					goto l638
				}
			l642:
				{
					position643, tokenIndex643, depth643 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l643
					}
					goto l642
				l643:
					position, tokenIndex, depth = position643, tokenIndex643, depth643
				}
			l644:
				{
					position645, tokenIndex645, depth645 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l645
					}
					position++
				l646:
					{
						position647, tokenIndex647, depth647 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l647
						}
						goto l646
					l647:
						position, tokenIndex, depth = position647, tokenIndex647, depth647
					}
					if !_rules[rulearg]() {
						goto l645
					}
				l648:
					{
						position649, tokenIndex649, depth649 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l649
						}
						goto l648
					l649:
						position, tokenIndex, depth = position649, tokenIndex649, depth649
					}
					goto l644
				l645:
					position, tokenIndex, depth = position645, tokenIndex645, depth645
				}
				{
					position650, tokenIndex650, depth650 := position, tokenIndex, depth
					if !matchDot() {
						goto l650
					}
					goto l638
				l650:
					position, tokenIndex, depth = position650, tokenIndex650, depth650
				}
				depth--
				add(ruleargs, position639)
			}
			return true
		l638:
			position, tokenIndex, depth = position638, tokenIndex638, depth638
			return false
		},
		/* 65 arg <- <(expr Action40)> */
		func() bool {
			position651, tokenIndex651, depth651 := position, tokenIndex, depth
			{
				position652 := position
				depth++
				if !_rules[ruleexpr]() {
					goto l651
				}
				if !_rules[ruleAction40]() {
					goto l651
				}
				depth--
				add(rulearg, position652)
			}
			return true
		l651:
			position, tokenIndex, depth = position651, tokenIndex651, depth651
			return false
		},
		/* 66 imports <- <(isp* (fsep isp*)* import isp* (fsep isp* (fsep isp*)* import isp*)* (fsep isp*)* !.)> */
		func() bool {
			position653, tokenIndex653, depth653 := position, tokenIndex, depth
			{
				position654 := position
				depth++
			l655:
				{
					position656, tokenIndex656, depth656 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l656
					}
					goto l655
				l656:
					position, tokenIndex, depth = position656, tokenIndex656, depth656
				}
			l657:
				{
					position658, tokenIndex658, depth658 := position, tokenIndex, depth
					if !_rules[rulefsep]() {
						goto l658
					}
				l659:
					{
						position660, tokenIndex660, depth660 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l660
						}
						goto l659
					l660:
						position, tokenIndex, depth = position660, tokenIndex660, depth660
					}
					goto l657
				l658:
					position, tokenIndex, depth = position658, tokenIndex658, depth658
				}
				if !_rules[ruleimport]() {
					goto l653
				}
			l661:
				{
					position662, tokenIndex662, depth662 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l662
					}
					goto l661
				l662:
					position, tokenIndex, depth = position662, tokenIndex662, depth662
				}
			l663:
				{
					position664, tokenIndex664, depth664 := position, tokenIndex, depth
					if !_rules[rulefsep]() {
						goto l664
					}
				l665:
					{
						position666, tokenIndex666, depth666 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l666
						}
						goto l665
					l666:
						position, tokenIndex, depth = position666, tokenIndex666, depth666
					}
				l667:
					{
						position668, tokenIndex668, depth668 := position, tokenIndex, depth
						if !_rules[rulefsep]() {
							goto l668
						}
					l669:
						{
							position670, tokenIndex670, depth670 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l670
							}
							goto l669
						l670:
							position, tokenIndex, depth = position670, tokenIndex670, depth670
						}
						goto l667
					l668:
						position, tokenIndex, depth = position668, tokenIndex668, depth668
					}
					if !_rules[ruleimport]() {
						goto l664
					}
				l671:
					{
						position672, tokenIndex672, depth672 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l672
						}
						goto l671
					l672:
						position, tokenIndex, depth = position672, tokenIndex672, depth672
					}
					goto l663
				l664:
					position, tokenIndex, depth = position664, tokenIndex664, depth664
				}
			l673:
				{
					position674, tokenIndex674, depth674 := position, tokenIndex, depth
					if !_rules[rulefsep]() {
						goto l674
					}
				l675:
					{
						position676, tokenIndex676, depth676 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l676
						}
						goto l675
					l676:
						position, tokenIndex, depth = position676, tokenIndex676, depth676
					}
					goto l673
				l674:
					position, tokenIndex, depth = position674, tokenIndex674, depth674
				}
				{
					position677, tokenIndex677, depth677 := position, tokenIndex, depth
					if !matchDot() {
						goto l677
					}
					goto l653
				l677:
					position, tokenIndex, depth = position677, tokenIndex677, depth677
				}
				depth--
				add(ruleimports, position654)
			}
			return true
		l653:
			position, tokenIndex, depth = position653, tokenIndex653, depth653
			return false
		},
		/* 67 import <- <((tagname isp+)? '"' <(!'"' .)*> '"' Action41)> */
		func() bool {
			position678, tokenIndex678, depth678 := position, tokenIndex, depth
			{
				position679 := position
				depth++
				{
					position680, tokenIndex680, depth680 := position, tokenIndex, depth
					if !_rules[ruletagname]() {
						goto l680
					}
					if !_rules[ruleisp]() {
						goto l680
					}
				l682:
					{
						position683, tokenIndex683, depth683 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l683
						}
						goto l682
					l683:
						position, tokenIndex, depth = position683, tokenIndex683, depth683
					}
					goto l681
				l680:
					position, tokenIndex, depth = position680, tokenIndex680, depth680
				}
			l681:
				if buffer[position] != rune('"') {
					goto l678
				}
				position++
				{
					position684 := position
					depth++
				l685:
					{
						position686, tokenIndex686, depth686 := position, tokenIndex, depth
						{
							position687, tokenIndex687, depth687 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l687
							}
							position++
							goto l686
						l687:
							position, tokenIndex, depth = position687, tokenIndex687, depth687
						}
						if !matchDot() {
							goto l686
						}
						goto l685
					l686:
						position, tokenIndex, depth = position686, tokenIndex686, depth686
					}
					depth--
					add(rulePegText, position684)
				}
				if buffer[position] != rune('"') {
					goto l678
				}
				position++
				if !_rules[ruleAction41]() {
					goto l678
				}
				depth--
				add(ruleimport, position679)
			}
			return true
		l678:
			position, tokenIndex, depth = position678, tokenIndex678, depth678
			return false
		},
		/* 69 Action0 <- <{
			p.varMappings = append(p.varMappings,
				data.VariableMapping{Value: p.bv, Variable: p.goVal})
			p.goVal.Type = nil
			p.bv.IDs = nil
		}> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
		nil,
		/* 71 Action1 <- <{
			p.goVal.Name = buffer[begin:end]
		}> */
		func() bool {
			{
				add(ruleAction1, position)
			}
			return true
		},
		/* 72 Action2 <- <{
			p.goVal.Type = p.valuetype
			p.valuetype = nil
		}> */
		func() bool {
			{
				add(ruleAction2, position)
			}
			return true
		},
		/* 73 Action3 <- <{
			p.assignments = append(p.assignments, data.Assignment{Expression: p.expr,
				Target: p.bv})
			p.bv.IDs = nil
		}> */
		func() bool {
			{
				add(ruleAction3, position)
			}
			return true
		},
		/* 74 Action4 <- <{
			p.bv.Kind = data.BoundSelf
		}> */
		func() bool {
			{
				add(ruleAction4, position)
			}
			return true
		},
		/* 75 Action5 <- <{
			p.bv.Kind = data.BoundDataset
		}> */
		func() bool {
			{
				add(ruleAction5, position)
			}
			return true
		},
		/* 76 Action6 <- <{
			p.bv.Kind = data.BoundProperty
		}> */
		func() bool {
			{
				add(ruleAction6, position)
			}
			return true
		},
		/* 77 Action7 <- <{
			p.bv.Kind = data.BoundStyle
		}> */
		func() bool {
			{
				add(ruleAction7, position)
			}
			return true
		},
		/* 78 Action8 <- <{
			p.bv.Kind = data.BoundClass
		}> */
		func() bool {
			{
				add(ruleAction8, position)
			}
			return true
		},
		/* 79 Action9 <- <{
			p.bv.Kind = data.BoundFormValue
		}> */
		func() bool {
			{
				add(ruleAction9, position)
			}
			return true
		},
		/* 80 Action10 <- <{
			p.bv.Kind = data.BoundExpr
			p.bv.IDs = append(p.bv.IDs, p.expr)
		}> */
		func() bool {
			{
				add(ruleAction10, position)
			}
			return true
		},
		/* 81 Action11 <- <{
			p.bv.Kind = data.BoundEventValue
			if len(p.bv.IDs) == 0 {
				p.bv.IDs = append(p.bv.IDs, "")
			}
		}> */
		func() bool {
			{
				add(ruleAction11, position)
			}
			return true
		},
		/* 82 Action12 <- <{
			p.bv.IDs = append(p.bv.IDs, buffer[begin:end])
		}> */
		func() bool {
			{
				add(ruleAction12, position)
			}
			return true
		},
		/* 83 Action13 <- <{
			p.bv.IDs = append(p.bv.IDs, buffer[begin:end])
		}> */
		func() bool {
			{
				add(ruleAction13, position)
			}
			return true
		},
		/* 84 Action14 <- <{
			p.expr = buffer[begin:end]
		}> */
		func() bool {
			{
				add(ruleAction14, position)
			}
			return true
		},
		/* 85 Action15 <- <{
			var expr *string
			if p.expr != "" {
				expr = new(string)
				*expr = p.expr
			}
			for _, name := range p.names {
				p.fields = append(p.fields, &data.Field{Name: name, Type: p.valuetype, DefaultValue: expr})
			}
			p.expr = ""
			p.valuetype = nil
			p.names = nil
		}> */
		func() bool {
			{
				add(ruleAction15, position)
			}
			return true
		},
		/* 86 Action16 <- <{
			p.names = append(p.names, buffer[begin:end])
		}> */
		func() bool {
			{
				add(ruleAction16, position)
			}
			return true
		},
		/* 87 Action17 <- <{
			switch name := buffer[begin:end]; name {
			case "int":
				p.valuetype = &data.ParamType{Kind: data.IntType}
			case "bool":
				p.valuetype = &data.ParamType{Kind: data.BoolType}
			case "string":
				p.valuetype = &data.ParamType{Kind: data.StringType}
			default:
				p.valuetype = &data.ParamType{Kind: data.NamedType, Name: name}
			}
		}> */
		func() bool {
			{
				add(ruleAction17, position)
			}
			return true
		},
		/* 88 Action18 <- <{
			name := buffer[begin:end]
			if name == "js.Value" {
				p.valuetype = &data.ParamType{Kind: data.JSValueType}
			} else {
				p.valuetype = &data.ParamType{Kind: data.NamedType, Name: name}
			}
		}> */
		func() bool {
			{
				add(ruleAction18, position)
			}
			return true
		},
		/* 89 Action19 <- <{
			p.valuetype = &data.ParamType{Kind: data.ArrayType, ValueType: p.valuetype}
		}> */
		func() bool {
			{
				add(ruleAction19, position)
			}
			return true
		},
		/* 90 Action20 <- <{
			p.valuetype = &data.ParamType{Kind: data.MapType, KeyType: p.keytype, ValueType: p.valuetype}
		}> */
		func() bool {
			{
				add(ruleAction20, position)
			}
			return true
		},
		/* 91 Action21 <- <{
			p.valuetype = &data.ParamType{Kind: data.ChanType, ValueType: p.valuetype}
		}> */
		func() bool {
			{
				add(ruleAction21, position)
			}
			return true
		},
		/* 92 Action22 <- <{
			p.valuetype = &data.ParamType{Kind: data.FuncType, ValueType: p.valuetype,
				Params: p.params}
			p.params = nil
		}> */
		func() bool {
			{
				add(ruleAction22, position)
			}
			return true
		},
		/* 93 Action23 <- <{
			p.keytype = p.valuetype
		}> */
		func() bool {
			{
				add(ruleAction23, position)
			}
			return true
		},
		/* 94 Action24 <- <{
			p.valuetype = &data.ParamType{Kind: data.PointerType, ValueType: p.valuetype}
		}> */
		func() bool {
			{
				add(ruleAction24, position)
			}
			return true
		},
		/* 95 Action25 <- <{
			p.eventMappings = append(p.eventMappings, data.UnboundEventMapping{
				Event: p.eventName, Handler: p.handlername, ParamMappings: p.paramMappings,
				Handling: p.eventHandling})
			p.eventHandling = data.AutoPreventDefault
			p.expr = ""
			p.paramMappings = make(map[string]data.BoundValue)
		}> */
		func() bool {
			{
				add(ruleAction25, position)
			}
			return true
		},
		/* 96 Action26 <- <{
			p.handlername = buffer[begin:end]
		}> */
		func() bool {
			{
				add(ruleAction26, position)
			}
			return true
		},
		/* 97 Action27 <- <{
			p.eventName = buffer[begin:end]
		}> */
		func() bool {
			{
				add(ruleAction27, position)
			}
			return true
		},
		/* 98 Action28 <- <{
			p.paramIndex = 0
			p.tagname = ""
		}> */
		func() bool {
			{
				add(ruleAction28, position)
			}
			return true
		},
		/* 99 Action29 <- <{
			if p.tagname == "" {
				if p.paramIndex == -1 {
					p.err = errors.New("unnamed parameter mapping after named one")
					return
				}
				p.tagname = fmt.Sprintf("~%v", p.paramIndex)
				p.paramIndex++
			} else {
				if _, ok := p.paramMappings[p.tagname]; ok {
					p.err = errors.New("duplicate param: " + p.tagname)
					return
				}
				p.paramIndex = -1
			}
			p.paramMappings[p.tagname] = p.bv
			p.tagname = ""
			p.bv.IDs = nil
		}> */
		func() bool {
			{
				add(ruleAction29, position)
			}
			return true
		},
		/* 100 Action30 <- <{
			p.tagname = buffer[begin:end]
		}> */
		func() bool {
			{
				add(ruleAction30, position)
			}
			return true
		},
		/* 101 Action31 <- <{
			switch p.tagname {
			case "preventDefault":
				if p.eventHandling != data.AutoPreventDefault {
					p.err = errors.New("duplicate preventDefault")
					return
				}
				switch len(p.names) {
				case 0:
					p.eventHandling = data.PreventDefault
				case 1:
					switch p.names[0] {
					case "true":
						p.eventHandling = data.PreventDefault
					case "false":
						p.eventHandling = data.DontPreventDefault
					case "ask":
						p.eventHandling = data.AskPreventDefault
					default:
						p.err = fmt.Errorf("unsupported value for preventDefault: %s", p.names[0])
						return
					}
				default:
					p.err = errors.New("too many parameters for preventDefault")
					return
				}
			default:
				p.err = errors.New("unknown tag: " + p.tagname)
				return
			}
			p.names = nil
		}> */
		func() bool {
			{
				add(ruleAction31, position)
			}
			return true
		},
		/* 102 Action32 <- <{
			p.tagname = buffer[begin:end]
		}> */
		func() bool {
			{
				add(ruleAction32, position)
			}
			return true
		},
		/* 103 Action33 <- <{
			p.names = append(p.names, buffer[begin:end])
		}> */
		func() bool {
			{
				add(ruleAction33, position)
			}
			return true
		},
		/* 104 Action34 <- <{
			p.names = append(p.names, buffer[begin:end])
		}> */
		func() bool {
			{
				add(ruleAction34, position)
			}
			return true
		},
		/* 105 Action35 <- <{
			p.handlers = append(p.handlers, HandlerSpec{
				Name: p.handlername, Params: p.params, Returns: p.valuetype})
			p.valuetype = nil
			p.params = nil
		}> */
		func() bool {
			{
				add(ruleAction35, position)
			}
			return true
		},
		/* 106 Action36 <- <{
			p.paramnames = append(p.paramnames, buffer[begin:end])
		}> */
		func() bool {
			{
				add(ruleAction36, position)
			}
			return true
		},
		/* 107 Action37 <- <{
			name := p.paramnames[len(p.paramnames)-1]
			p.paramnames = p.paramnames[:len(p.paramnames)-1]
			for _, para := range p.params {
				if para.Name == name {
					p.err = errors.New("duplicate param name: " + para.Name)
					return
				}
			}

			p.params = append(p.params, data.Param{Name: name, Type: p.valuetype})
			p.valuetype = nil
		}> */
		func() bool {
			{
				add(ruleAction37, position)
			}
			return true
		},
		/* 108 Action38 <- <{
			p.cParams = append(p.cParams, data.ComponentParam{
				Name: p.tagname, Type: *p.valuetype, IsVar: p.isVar})
			p.valuetype = nil
			p.isVar = false
		}> */
		func() bool {
			{
				add(ruleAction38, position)
			}
			return true
		},
		/* 109 Action39 <- <{
			p.isVar = true
		}> */
		func() bool {
			{
				add(ruleAction39, position)
			}
			return true
		},
		/* 110 Action40 <- <{
		  p.names = append(p.names, p.expr)
		}> */
		func() bool {
			{
				add(ruleAction40, position)
			}
			return true
		},
		/* 111 Action41 <- <{
			path := buffer[begin:end]
			if p.tagname == "" {
				lastDot := strings.LastIndexByte(path, '/')
				if lastDot == -1 {
					p.tagname = path
				} else {
					p.tagname = path[lastDot+1:]
				}
			}
			if _, ok := p.imports[p.tagname]; ok {
				p.err = errors.New("duplicate import name: " + p.tagname)
				return
			}
			p.imports[p.tagname] = path
			p.tagname = ""
		}> */
		func() bool {
			{
				add(ruleAction41, position)
			}
			return true
		},
	}
	p.rules = _rules
}
