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
	rulekeytype
	rulepointer
	rulecaptures
	rulecapture
	rulehandlername
	ruleeventid
	rulemappings
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
	"keytype",
	"pointer",
	"captures",
	"capture",
	"handlername",
	"eventid",
	"mappings",
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
	eventHandling              data.EventHandling
	expr, tagname, handlername string
	names                      []string
	keytype, valuetype         *data.ParamType
	fields                     []*data.Field
	bv                         data.BoundValue
	goVal                      data.GoValue
	paramMappings              map[string]data.BoundValue
	params                     []data.Param
	isVar                      bool
	err                        error

	assignments   []data.Assignment
	varMappings   []data.VariableMapping
	eventMappings []data.UnboundEventMapping
	handlers      []HandlerSpec
	cParams       []data.ComponentParam
	imports       map[string]string

	Buffer string
	buffer []rune
	rules  [102]func() bool
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

			p.bv.Kind = data.BoundEventValue
			if len(p.bv.IDs) == 0 {
				p.bv.IDs = append(p.bv.IDs, "")
			}

		case ruleAction11:

			p.bv.IDs = append(p.bv.IDs, buffer[begin:end])

		case ruleAction12:

			p.bv.IDs = append(p.bv.IDs, buffer[begin:end])

		case ruleAction13:

			p.expr = buffer[begin:end]

		case ruleAction14:

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

		case ruleAction15:

			p.names = append(p.names, buffer[begin:end])

		case ruleAction16:

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

		case ruleAction17:

			name := buffer[begin:end]
			if name == "js.Value" {
				p.valuetype = &data.ParamType{Kind: data.JSValueType}
			} else {
				p.valuetype = &data.ParamType{Kind: data.NamedType, Name: name}
			}

		case ruleAction18:

			p.valuetype = &data.ParamType{Kind: data.ArrayType, ValueType: p.valuetype}

		case ruleAction19:

			p.valuetype = &data.ParamType{Kind: data.MapType, KeyType: p.keytype, ValueType: p.valuetype}

		case ruleAction20:

			p.keytype = p.valuetype

		case ruleAction21:

			p.valuetype = &data.ParamType{Kind: data.PointerType, ValueType: p.valuetype}

		case ruleAction22:

			p.eventMappings = append(p.eventMappings, data.UnboundEventMapping{
				Event: p.expr, Handler: p.handlername, ParamMappings: p.paramMappings,
				Handling: p.eventHandling})
			p.eventHandling = data.AutoPreventDefault
			p.expr = ""
			p.paramMappings = make(map[string]data.BoundValue)

		case ruleAction23:

			p.handlername = buffer[begin:end]

		case ruleAction24:

			p.expr = buffer[begin:end]

		case ruleAction25:

			if _, ok := p.paramMappings[p.tagname]; ok {
				p.err = errors.New("duplicate param: " + p.tagname)
				return
			}
			p.paramMappings[p.tagname] = p.bv
			p.bv.IDs = nil

		case ruleAction26:

			p.tagname = buffer[begin:end]

		case ruleAction27:

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

		case ruleAction28:

			p.tagname = buffer[begin:end]

		case ruleAction29:

			p.names = append(p.names, buffer[begin:end])

		case ruleAction30:

			p.names = append(p.names, buffer[begin:end])

		case ruleAction31:

			p.handlers = append(p.handlers, HandlerSpec{
				Name: p.handlername, Params: p.params, Returns: p.valuetype})
			p.valuetype = nil
			p.params = nil

		case ruleAction32:

			for _, para := range p.params {
				if para.Name == p.tagname {
					p.err = errors.New("duplicate param name: " + para.Name)
					return
				}
			}

			p.params = append(p.params, data.Param{Name: p.tagname, Type: p.valuetype})
			p.valuetype = nil

		case ruleAction33:

			p.cParams = append(p.cParams, data.ComponentParam{
				Name: p.tagname, Type: *p.valuetype, IsVar: p.isVar})
			p.valuetype = nil
			p.isVar = false

		case ruleAction34:

			p.isVar = true

		case ruleAction35:

			p.names = append(p.names, p.expr)

		case ruleAction36:

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
		/* 8 bound <- <(self / ((&('E' | 'e') event) | (&('F' | 'f') form) | (&('C' | 'c') class) | (&('S' | 's') style) | (&('P' | 'p') prop) | (&('D' | 'd') dataset)))> */
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
		/* 15 event <- <(('e' / 'E') ('v' / 'V') ('e' / 'E') ('n' / 'N') ('t' / 'T') isp* '(' isp* jsid? isp* ')' Action10)> */
		func() bool {
			position187, tokenIndex187, depth187 := position, tokenIndex, depth
			{
				position188 := position
				depth++
				{
					position189, tokenIndex189, depth189 := position, tokenIndex, depth
					if buffer[position] != rune('e') {
						goto l190
					}
					position++
					goto l189
				l190:
					position, tokenIndex, depth = position189, tokenIndex189, depth189
					if buffer[position] != rune('E') {
						goto l187
					}
					position++
				}
			l189:
				{
					position191, tokenIndex191, depth191 := position, tokenIndex, depth
					if buffer[position] != rune('v') {
						goto l192
					}
					position++
					goto l191
				l192:
					position, tokenIndex, depth = position191, tokenIndex191, depth191
					if buffer[position] != rune('V') {
						goto l187
					}
					position++
				}
			l191:
				{
					position193, tokenIndex193, depth193 := position, tokenIndex, depth
					if buffer[position] != rune('e') {
						goto l194
					}
					position++
					goto l193
				l194:
					position, tokenIndex, depth = position193, tokenIndex193, depth193
					if buffer[position] != rune('E') {
						goto l187
					}
					position++
				}
			l193:
				{
					position195, tokenIndex195, depth195 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l196
					}
					position++
					goto l195
				l196:
					position, tokenIndex, depth = position195, tokenIndex195, depth195
					if buffer[position] != rune('N') {
						goto l187
					}
					position++
				}
			l195:
				{
					position197, tokenIndex197, depth197 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l198
					}
					position++
					goto l197
				l198:
					position, tokenIndex, depth = position197, tokenIndex197, depth197
					if buffer[position] != rune('T') {
						goto l187
					}
					position++
				}
			l197:
			l199:
				{
					position200, tokenIndex200, depth200 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l200
					}
					goto l199
				l200:
					position, tokenIndex, depth = position200, tokenIndex200, depth200
				}
				if buffer[position] != rune('(') {
					goto l187
				}
				position++
			l201:
				{
					position202, tokenIndex202, depth202 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l202
					}
					goto l201
				l202:
					position, tokenIndex, depth = position202, tokenIndex202, depth202
				}
				{
					position203, tokenIndex203, depth203 := position, tokenIndex, depth
					if !_rules[rulejsid]() {
						goto l203
					}
					goto l204
				l203:
					position, tokenIndex, depth = position203, tokenIndex203, depth203
				}
			l204:
			l205:
				{
					position206, tokenIndex206, depth206 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l206
					}
					goto l205
				l206:
					position, tokenIndex, depth = position206, tokenIndex206, depth206
				}
				if buffer[position] != rune(')') {
					goto l187
				}
				position++
				if !_rules[ruleAction10]() {
					goto l187
				}
				depth--
				add(ruleevent, position188)
			}
			return true
		l187:
			position, tokenIndex, depth = position187, tokenIndex187, depth187
			return false
		},
		/* 16 htmlid <- <(<((&('-') '-') | (&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))+> Action11)> */
		func() bool {
			position207, tokenIndex207, depth207 := position, tokenIndex, depth
			{
				position208 := position
				depth++
				{
					position209 := position
					depth++
					{
						switch buffer[position] {
						case '-':
							if buffer[position] != rune('-') {
								goto l207
							}
							position++
							break
						case '_':
							if buffer[position] != rune('_') {
								goto l207
							}
							position++
							break
						case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l207
							}
							position++
							break
						case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l207
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l207
							}
							position++
							break
						}
					}

				l210:
					{
						position211, tokenIndex211, depth211 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '-':
								if buffer[position] != rune('-') {
									goto l211
								}
								position++
								break
							case '_':
								if buffer[position] != rune('_') {
									goto l211
								}
								position++
								break
							case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
								if c := buffer[position]; c < rune('A') || c > rune('Z') {
									goto l211
								}
								position++
								break
							case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l211
								}
								position++
								break
							default:
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l211
								}
								position++
								break
							}
						}

						goto l210
					l211:
						position, tokenIndex, depth = position211, tokenIndex211, depth211
					}
					depth--
					add(rulePegText, position209)
				}
				if !_rules[ruleAction11]() {
					goto l207
				}
				depth--
				add(rulehtmlid, position208)
			}
			return true
		l207:
			position, tokenIndex, depth = position207, tokenIndex207, depth207
			return false
		},
		/* 17 jsid <- <(<(((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z])) ((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))*)> Action12)> */
		func() bool {
			position214, tokenIndex214, depth214 := position, tokenIndex, depth
			{
				position215 := position
				depth++
				{
					position216 := position
					depth++
					{
						switch buffer[position] {
						case '_':
							if buffer[position] != rune('_') {
								goto l214
							}
							position++
							break
						case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l214
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l214
							}
							position++
							break
						}
					}

				l218:
					{
						position219, tokenIndex219, depth219 := position, tokenIndex, depth
						{
							switch buffer[position] {
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

						goto l218
					l219:
						position, tokenIndex, depth = position219, tokenIndex219, depth219
					}
					depth--
					add(rulePegText, position216)
				}
				if !_rules[ruleAction12]() {
					goto l214
				}
				depth--
				add(rulejsid, position215)
			}
			return true
		l214:
			position, tokenIndex, depth = position214, tokenIndex214, depth214
			return false
		},
		/* 18 expr <- <(<((&('\t' | ' ') isp+) | (&('(' | '[' | '{') enclosed) | (&('!' | '"' | '&' | '*' | '+' | '-' | '.' | '/' | '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9' | ':' | '<' | '=' | '>' | 'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '^' | '_' | '`' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z' | '|') commaless))+> Action13)> */
		func() bool {
			position221, tokenIndex221, depth221 := position, tokenIndex, depth
			{
				position222 := position
				depth++
				{
					position223 := position
					depth++
					{
						switch buffer[position] {
						case '\t', ' ':
							if !_rules[ruleisp]() {
								goto l221
							}
						l227:
							{
								position228, tokenIndex228, depth228 := position, tokenIndex, depth
								if !_rules[ruleisp]() {
									goto l228
								}
								goto l227
							l228:
								position, tokenIndex, depth = position228, tokenIndex228, depth228
							}
							break
						case '(', '[', '{':
							if !_rules[ruleenclosed]() {
								goto l221
							}
							break
						default:
							if !_rules[rulecommaless]() {
								goto l221
							}
							break
						}
					}

				l224:
					{
						position225, tokenIndex225, depth225 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '\t', ' ':
								if !_rules[ruleisp]() {
									goto l225
								}
							l230:
								{
									position231, tokenIndex231, depth231 := position, tokenIndex, depth
									if !_rules[ruleisp]() {
										goto l231
									}
									goto l230
								l231:
									position, tokenIndex, depth = position231, tokenIndex231, depth231
								}
								break
							case '(', '[', '{':
								if !_rules[ruleenclosed]() {
									goto l225
								}
								break
							default:
								if !_rules[rulecommaless]() {
									goto l225
								}
								break
							}
						}

						goto l224
					l225:
						position, tokenIndex, depth = position225, tokenIndex225, depth225
					}
					depth--
					add(rulePegText, position223)
				}
				if !_rules[ruleAction13]() {
					goto l221
				}
				depth--
				add(ruleexpr, position222)
			}
			return true
		l221:
			position, tokenIndex, depth = position221, tokenIndex221, depth221
			return false
		},
		/* 19 commaless <- <((((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))+ '.' ((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))+) / ((&('"' | '`') string) | (&('!' | '&' | '*' | '+' | '-' | '.' | '/' | ':' | '<' | '=' | '>' | '^' | '|') operators) | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') number) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '_' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') identifier)))> */
		func() bool {
			position232, tokenIndex232, depth232 := position, tokenIndex, depth
			{
				position233 := position
				depth++
				{
					position234, tokenIndex234, depth234 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '_':
							if buffer[position] != rune('_') {
								goto l235
							}
							position++
							break
						case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l235
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l235
							}
							position++
							break
						}
					}

				l236:
					{
						position237, tokenIndex237, depth237 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '_':
								if buffer[position] != rune('_') {
									goto l237
								}
								position++
								break
							case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
								if c := buffer[position]; c < rune('A') || c > rune('Z') {
									goto l237
								}
								position++
								break
							default:
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l237
								}
								position++
								break
							}
						}

						goto l236
					l237:
						position, tokenIndex, depth = position237, tokenIndex237, depth237
					}
					if buffer[position] != rune('.') {
						goto l235
					}
					position++
					{
						switch buffer[position] {
						case '_':
							if buffer[position] != rune('_') {
								goto l235
							}
							position++
							break
						case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l235
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l235
							}
							position++
							break
						}
					}

				l240:
					{
						position241, tokenIndex241, depth241 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '_':
								if buffer[position] != rune('_') {
									goto l241
								}
								position++
								break
							case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
								if c := buffer[position]; c < rune('A') || c > rune('Z') {
									goto l241
								}
								position++
								break
							default:
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l241
								}
								position++
								break
							}
						}

						goto l240
					l241:
						position, tokenIndex, depth = position241, tokenIndex241, depth241
					}
					goto l234
				l235:
					position, tokenIndex, depth = position234, tokenIndex234, depth234
					{
						switch buffer[position] {
						case '"', '`':
							if !_rules[rulestring]() {
								goto l232
							}
							break
						case '!', '&', '*', '+', '-', '.', '/', ':', '<', '=', '>', '^', '|':
							if !_rules[ruleoperators]() {
								goto l232
							}
							break
						case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
							if !_rules[rulenumber]() {
								goto l232
							}
							break
						default:
							if !_rules[ruleidentifier]() {
								goto l232
							}
							break
						}
					}

				}
			l234:
				depth--
				add(rulecommaless, position233)
			}
			return true
		l232:
			position, tokenIndex, depth = position232, tokenIndex232, depth232
			return false
		},
		/* 20 number <- <[0-9]+> */
		func() bool {
			position245, tokenIndex245, depth245 := position, tokenIndex, depth
			{
				position246 := position
				depth++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l245
				}
				position++
			l247:
				{
					position248, tokenIndex248, depth248 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l248
					}
					position++
					goto l247
				l248:
					position, tokenIndex, depth = position248, tokenIndex248, depth248
				}
				depth--
				add(rulenumber, position246)
			}
			return true
		l245:
			position, tokenIndex, depth = position245, tokenIndex245, depth245
			return false
		},
		/* 21 operators <- <((&('>') '>') | (&('<') '<') | (&('!') '!') | (&('.') '.') | (&('=') '=') | (&(':') ':') | (&('^') '^') | (&('&') '&') | (&('|') '|') | (&('/') '/') | (&('*') '*') | (&('-') '-') | (&('+') '+'))+> */
		func() bool {
			position249, tokenIndex249, depth249 := position, tokenIndex, depth
			{
				position250 := position
				depth++
				{
					switch buffer[position] {
					case '>':
						if buffer[position] != rune('>') {
							goto l249
						}
						position++
						break
					case '<':
						if buffer[position] != rune('<') {
							goto l249
						}
						position++
						break
					case '!':
						if buffer[position] != rune('!') {
							goto l249
						}
						position++
						break
					case '.':
						if buffer[position] != rune('.') {
							goto l249
						}
						position++
						break
					case '=':
						if buffer[position] != rune('=') {
							goto l249
						}
						position++
						break
					case ':':
						if buffer[position] != rune(':') {
							goto l249
						}
						position++
						break
					case '^':
						if buffer[position] != rune('^') {
							goto l249
						}
						position++
						break
					case '&':
						if buffer[position] != rune('&') {
							goto l249
						}
						position++
						break
					case '|':
						if buffer[position] != rune('|') {
							goto l249
						}
						position++
						break
					case '/':
						if buffer[position] != rune('/') {
							goto l249
						}
						position++
						break
					case '*':
						if buffer[position] != rune('*') {
							goto l249
						}
						position++
						break
					case '-':
						if buffer[position] != rune('-') {
							goto l249
						}
						position++
						break
					default:
						if buffer[position] != rune('+') {
							goto l249
						}
						position++
						break
					}
				}

			l251:
				{
					position252, tokenIndex252, depth252 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '>':
							if buffer[position] != rune('>') {
								goto l252
							}
							position++
							break
						case '<':
							if buffer[position] != rune('<') {
								goto l252
							}
							position++
							break
						case '!':
							if buffer[position] != rune('!') {
								goto l252
							}
							position++
							break
						case '.':
							if buffer[position] != rune('.') {
								goto l252
							}
							position++
							break
						case '=':
							if buffer[position] != rune('=') {
								goto l252
							}
							position++
							break
						case ':':
							if buffer[position] != rune(':') {
								goto l252
							}
							position++
							break
						case '^':
							if buffer[position] != rune('^') {
								goto l252
							}
							position++
							break
						case '&':
							if buffer[position] != rune('&') {
								goto l252
							}
							position++
							break
						case '|':
							if buffer[position] != rune('|') {
								goto l252
							}
							position++
							break
						case '/':
							if buffer[position] != rune('/') {
								goto l252
							}
							position++
							break
						case '*':
							if buffer[position] != rune('*') {
								goto l252
							}
							position++
							break
						case '-':
							if buffer[position] != rune('-') {
								goto l252
							}
							position++
							break
						default:
							if buffer[position] != rune('+') {
								goto l252
							}
							position++
							break
						}
					}

					goto l251
				l252:
					position, tokenIndex, depth = position252, tokenIndex252, depth252
				}
				depth--
				add(ruleoperators, position250)
			}
			return true
		l249:
			position, tokenIndex, depth = position249, tokenIndex249, depth249
			return false
		},
		/* 22 string <- <(('`' (!'`' .)* '`') / ('"' ((!'"' .) / ('\\' '"'))* '"'))> */
		func() bool {
			position255, tokenIndex255, depth255 := position, tokenIndex, depth
			{
				position256 := position
				depth++
				{
					position257, tokenIndex257, depth257 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l258
					}
					position++
				l259:
					{
						position260, tokenIndex260, depth260 := position, tokenIndex, depth
						{
							position261, tokenIndex261, depth261 := position, tokenIndex, depth
							if buffer[position] != rune('`') {
								goto l261
							}
							position++
							goto l260
						l261:
							position, tokenIndex, depth = position261, tokenIndex261, depth261
						}
						if !matchDot() {
							goto l260
						}
						goto l259
					l260:
						position, tokenIndex, depth = position260, tokenIndex260, depth260
					}
					if buffer[position] != rune('`') {
						goto l258
					}
					position++
					goto l257
				l258:
					position, tokenIndex, depth = position257, tokenIndex257, depth257
					if buffer[position] != rune('"') {
						goto l255
					}
					position++
				l262:
					{
						position263, tokenIndex263, depth263 := position, tokenIndex, depth
						{
							position264, tokenIndex264, depth264 := position, tokenIndex, depth
							{
								position266, tokenIndex266, depth266 := position, tokenIndex, depth
								if buffer[position] != rune('"') {
									goto l266
								}
								position++
								goto l265
							l266:
								position, tokenIndex, depth = position266, tokenIndex266, depth266
							}
							if !matchDot() {
								goto l265
							}
							goto l264
						l265:
							position, tokenIndex, depth = position264, tokenIndex264, depth264
							if buffer[position] != rune('\\') {
								goto l263
							}
							position++
							if buffer[position] != rune('"') {
								goto l263
							}
							position++
						}
					l264:
						goto l262
					l263:
						position, tokenIndex, depth = position263, tokenIndex263, depth263
					}
					if buffer[position] != rune('"') {
						goto l255
					}
					position++
				}
			l257:
				depth--
				add(rulestring, position256)
			}
			return true
		l255:
			position, tokenIndex, depth = position255, tokenIndex255, depth255
			return false
		},
		/* 23 enclosed <- <((&('[') brackets) | (&('{') braces) | (&('(') parens))> */
		func() bool {
			position267, tokenIndex267, depth267 := position, tokenIndex, depth
			{
				position268 := position
				depth++
				{
					switch buffer[position] {
					case '[':
						if !_rules[rulebrackets]() {
							goto l267
						}
						break
					case '{':
						if !_rules[rulebraces]() {
							goto l267
						}
						break
					default:
						if !_rules[ruleparens]() {
							goto l267
						}
						break
					}
				}

				depth--
				add(ruleenclosed, position268)
			}
			return true
		l267:
			position, tokenIndex, depth = position267, tokenIndex267, depth267
			return false
		},
		/* 24 parens <- <('(' inner ')')> */
		func() bool {
			position270, tokenIndex270, depth270 := position, tokenIndex, depth
			{
				position271 := position
				depth++
				if buffer[position] != rune('(') {
					goto l270
				}
				position++
				if !_rules[ruleinner]() {
					goto l270
				}
				if buffer[position] != rune(')') {
					goto l270
				}
				position++
				depth--
				add(ruleparens, position271)
			}
			return true
		l270:
			position, tokenIndex, depth = position270, tokenIndex270, depth270
			return false
		},
		/* 25 braces <- <('{' inner '}')> */
		func() bool {
			position272, tokenIndex272, depth272 := position, tokenIndex, depth
			{
				position273 := position
				depth++
				if buffer[position] != rune('{') {
					goto l272
				}
				position++
				if !_rules[ruleinner]() {
					goto l272
				}
				if buffer[position] != rune('}') {
					goto l272
				}
				position++
				depth--
				add(rulebraces, position273)
			}
			return true
		l272:
			position, tokenIndex, depth = position272, tokenIndex272, depth272
			return false
		},
		/* 26 brackets <- <('[' inner ']')> */
		func() bool {
			position274, tokenIndex274, depth274 := position, tokenIndex, depth
			{
				position275 := position
				depth++
				if buffer[position] != rune('[') {
					goto l274
				}
				position++
				if !_rules[ruleinner]() {
					goto l274
				}
				if buffer[position] != rune(']') {
					goto l274
				}
				position++
				depth--
				add(rulebrackets, position275)
			}
			return true
		l274:
			position, tokenIndex, depth = position274, tokenIndex274, depth274
			return false
		},
		/* 27 inner <- <((&('\t' | ' ') isp+) | (&(',') ',') | (&('(' | '[' | '{') enclosed) | (&('!' | '"' | '&' | '*' | '+' | '-' | '.' | '/' | '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9' | ':' | '<' | '=' | '>' | 'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '^' | '_' | '`' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z' | '|') commaless))*> */
		func() bool {
			{
				position277 := position
				depth++
			l278:
				{
					position279, tokenIndex279, depth279 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '\t', ' ':
							if !_rules[ruleisp]() {
								goto l279
							}
						l281:
							{
								position282, tokenIndex282, depth282 := position, tokenIndex, depth
								if !_rules[ruleisp]() {
									goto l282
								}
								goto l281
							l282:
								position, tokenIndex, depth = position282, tokenIndex282, depth282
							}
							break
						case ',':
							if buffer[position] != rune(',') {
								goto l279
							}
							position++
							break
						case '(', '[', '{':
							if !_rules[ruleenclosed]() {
								goto l279
							}
							break
						default:
							if !_rules[rulecommaless]() {
								goto l279
							}
							break
						}
					}

					goto l278
				l279:
					position, tokenIndex, depth = position279, tokenIndex279, depth279
				}
				depth--
				add(ruleinner, position277)
			}
			return true
		},
		/* 28 identifier <- <(((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z])) ((&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') ([0-9] / [0-9])) | (&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))*)> */
		func() bool {
			position283, tokenIndex283, depth283 := position, tokenIndex, depth
			{
				position284 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l283
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l283
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l283
						}
						position++
						break
					}
				}

			l286:
				{
					position287, tokenIndex287, depth287 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
							{
								position289, tokenIndex289, depth289 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l290
								}
								position++
								goto l289
							l290:
								position, tokenIndex, depth = position289, tokenIndex289, depth289
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l287
								}
								position++
							}
						l289:
							break
						case '_':
							if buffer[position] != rune('_') {
								goto l287
							}
							position++
							break
						case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l287
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l287
							}
							position++
							break
						}
					}

					goto l286
				l287:
					position, tokenIndex, depth = position287, tokenIndex287, depth287
				}
				depth--
				add(ruleidentifier, position284)
			}
			return true
		l283:
			position, tokenIndex, depth = position283, tokenIndex283, depth283
			return false
		},
		/* 29 fields <- <(((&('\n') '\n') | (&('\t') '\t') | (&(' ') ' ') | (&(';') ';'))* field isp* (fsep isp* (fsep isp*)* field)* ((&('\n') '\n') | (&('\t') '\t') | (&(' ') ' ') | (&(';') ';'))* !.)> */
		func() bool {
			position291, tokenIndex291, depth291 := position, tokenIndex, depth
			{
				position292 := position
				depth++
			l293:
				{
					position294, tokenIndex294, depth294 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '\n':
							if buffer[position] != rune('\n') {
								goto l294
							}
							position++
							break
						case '\t':
							if buffer[position] != rune('\t') {
								goto l294
							}
							position++
							break
						case ' ':
							if buffer[position] != rune(' ') {
								goto l294
							}
							position++
							break
						default:
							if buffer[position] != rune(';') {
								goto l294
							}
							position++
							break
						}
					}

					goto l293
				l294:
					position, tokenIndex, depth = position294, tokenIndex294, depth294
				}
				if !_rules[rulefield]() {
					goto l291
				}
			l296:
				{
					position297, tokenIndex297, depth297 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l297
					}
					goto l296
				l297:
					position, tokenIndex, depth = position297, tokenIndex297, depth297
				}
			l298:
				{
					position299, tokenIndex299, depth299 := position, tokenIndex, depth
					if !_rules[rulefsep]() {
						goto l299
					}
				l300:
					{
						position301, tokenIndex301, depth301 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l301
						}
						goto l300
					l301:
						position, tokenIndex, depth = position301, tokenIndex301, depth301
					}
				l302:
					{
						position303, tokenIndex303, depth303 := position, tokenIndex, depth
						if !_rules[rulefsep]() {
							goto l303
						}
					l304:
						{
							position305, tokenIndex305, depth305 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l305
							}
							goto l304
						l305:
							position, tokenIndex, depth = position305, tokenIndex305, depth305
						}
						goto l302
					l303:
						position, tokenIndex, depth = position303, tokenIndex303, depth303
					}
					if !_rules[rulefield]() {
						goto l299
					}
					goto l298
				l299:
					position, tokenIndex, depth = position299, tokenIndex299, depth299
				}
			l306:
				{
					position307, tokenIndex307, depth307 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '\n':
							if buffer[position] != rune('\n') {
								goto l307
							}
							position++
							break
						case '\t':
							if buffer[position] != rune('\t') {
								goto l307
							}
							position++
							break
						case ' ':
							if buffer[position] != rune(' ') {
								goto l307
							}
							position++
							break
						default:
							if buffer[position] != rune(';') {
								goto l307
							}
							position++
							break
						}
					}

					goto l306
				l307:
					position, tokenIndex, depth = position307, tokenIndex307, depth307
				}
				{
					position309, tokenIndex309, depth309 := position, tokenIndex, depth
					if !matchDot() {
						goto l309
					}
					goto l291
				l309:
					position, tokenIndex, depth = position309, tokenIndex309, depth309
				}
				depth--
				add(rulefields, position292)
			}
			return true
		l291:
			position, tokenIndex, depth = position291, tokenIndex291, depth291
			return false
		},
		/* 30 fsep <- <(';' / '\n')> */
		func() bool {
			position310, tokenIndex310, depth310 := position, tokenIndex, depth
			{
				position311 := position
				depth++
				{
					position312, tokenIndex312, depth312 := position, tokenIndex, depth
					if buffer[position] != rune(';') {
						goto l313
					}
					position++
					goto l312
				l313:
					position, tokenIndex, depth = position312, tokenIndex312, depth312
					if buffer[position] != rune('\n') {
						goto l310
					}
					position++
				}
			l312:
				depth--
				add(rulefsep, position311)
			}
			return true
		l310:
			position, tokenIndex, depth = position310, tokenIndex310, depth310
			return false
		},
		/* 31 field <- <(name (isp* ',' isp* name)* isp+ type isp* ('=' isp* expr)? Action14)> */
		func() bool {
			position314, tokenIndex314, depth314 := position, tokenIndex, depth
			{
				position315 := position
				depth++
				if !_rules[rulename]() {
					goto l314
				}
			l316:
				{
					position317, tokenIndex317, depth317 := position, tokenIndex, depth
				l318:
					{
						position319, tokenIndex319, depth319 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l319
						}
						goto l318
					l319:
						position, tokenIndex, depth = position319, tokenIndex319, depth319
					}
					if buffer[position] != rune(',') {
						goto l317
					}
					position++
				l320:
					{
						position321, tokenIndex321, depth321 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l321
						}
						goto l320
					l321:
						position, tokenIndex, depth = position321, tokenIndex321, depth321
					}
					if !_rules[rulename]() {
						goto l317
					}
					goto l316
				l317:
					position, tokenIndex, depth = position317, tokenIndex317, depth317
				}
				if !_rules[ruleisp]() {
					goto l314
				}
			l322:
				{
					position323, tokenIndex323, depth323 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l323
					}
					goto l322
				l323:
					position, tokenIndex, depth = position323, tokenIndex323, depth323
				}
				if !_rules[ruletype]() {
					goto l314
				}
			l324:
				{
					position325, tokenIndex325, depth325 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l325
					}
					goto l324
				l325:
					position, tokenIndex, depth = position325, tokenIndex325, depth325
				}
				{
					position326, tokenIndex326, depth326 := position, tokenIndex, depth
					if buffer[position] != rune('=') {
						goto l326
					}
					position++
				l328:
					{
						position329, tokenIndex329, depth329 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l329
						}
						goto l328
					l329:
						position, tokenIndex, depth = position329, tokenIndex329, depth329
					}
					if !_rules[ruleexpr]() {
						goto l326
					}
					goto l327
				l326:
					position, tokenIndex, depth = position326, tokenIndex326, depth326
				}
			l327:
				if !_rules[ruleAction14]() {
					goto l314
				}
				depth--
				add(rulefield, position315)
			}
			return true
		l314:
			position, tokenIndex, depth = position314, tokenIndex314, depth314
			return false
		},
		/* 32 name <- <(<((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))+> Action15)> */
		func() bool {
			position330, tokenIndex330, depth330 := position, tokenIndex, depth
			{
				position331 := position
				depth++
				{
					position332 := position
					depth++
					{
						switch buffer[position] {
						case '_':
							if buffer[position] != rune('_') {
								goto l330
							}
							position++
							break
						case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l330
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l330
							}
							position++
							break
						}
					}

				l333:
					{
						position334, tokenIndex334, depth334 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '_':
								if buffer[position] != rune('_') {
									goto l334
								}
								position++
								break
							case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
								if c := buffer[position]; c < rune('A') || c > rune('Z') {
									goto l334
								}
								position++
								break
							default:
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l334
								}
								position++
								break
							}
						}

						goto l333
					l334:
						position, tokenIndex, depth = position334, tokenIndex334, depth334
					}
					depth--
					add(rulePegText, position332)
				}
				if !_rules[ruleAction15]() {
					goto l330
				}
				depth--
				add(rulename, position331)
			}
			return true
		l330:
			position, tokenIndex, depth = position330, tokenIndex330, depth330
			return false
		},
		/* 33 type <- <(qname / sname / ((&('*') pointer) | (&('[') array) | (&('M' | 'm') map)))> */
		func() bool {
			position337, tokenIndex337, depth337 := position, tokenIndex, depth
			{
				position338 := position
				depth++
				{
					position339, tokenIndex339, depth339 := position, tokenIndex, depth
					if !_rules[ruleqname]() {
						goto l340
					}
					goto l339
				l340:
					position, tokenIndex, depth = position339, tokenIndex339, depth339
					if !_rules[rulesname]() {
						goto l341
					}
					goto l339
				l341:
					position, tokenIndex, depth = position339, tokenIndex339, depth339
					{
						switch buffer[position] {
						case '*':
							if !_rules[rulepointer]() {
								goto l337
							}
							break
						case '[':
							if !_rules[rulearray]() {
								goto l337
							}
							break
						default:
							if !_rules[rulemap]() {
								goto l337
							}
							break
						}
					}

				}
			l339:
				depth--
				add(ruletype, position338)
			}
			return true
		l337:
			position, tokenIndex, depth = position337, tokenIndex337, depth337
			return false
		},
		/* 34 sname <- <(<((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))+> Action16)> */
		func() bool {
			position343, tokenIndex343, depth343 := position, tokenIndex, depth
			{
				position344 := position
				depth++
				{
					position345 := position
					depth++
					{
						switch buffer[position] {
						case '_':
							if buffer[position] != rune('_') {
								goto l343
							}
							position++
							break
						case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l343
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l343
							}
							position++
							break
						}
					}

				l346:
					{
						position347, tokenIndex347, depth347 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '_':
								if buffer[position] != rune('_') {
									goto l347
								}
								position++
								break
							case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
								if c := buffer[position]; c < rune('A') || c > rune('Z') {
									goto l347
								}
								position++
								break
							default:
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l347
								}
								position++
								break
							}
						}

						goto l346
					l347:
						position, tokenIndex, depth = position347, tokenIndex347, depth347
					}
					depth--
					add(rulePegText, position345)
				}
				if !_rules[ruleAction16]() {
					goto l343
				}
				depth--
				add(rulesname, position344)
			}
			return true
		l343:
			position, tokenIndex, depth = position343, tokenIndex343, depth343
			return false
		},
		/* 35 qname <- <(<(((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))+ '.' ((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))+)> Action17)> */
		func() bool {
			position350, tokenIndex350, depth350 := position, tokenIndex, depth
			{
				position351 := position
				depth++
				{
					position352 := position
					depth++
					{
						switch buffer[position] {
						case '_':
							if buffer[position] != rune('_') {
								goto l350
							}
							position++
							break
						case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l350
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l350
							}
							position++
							break
						}
					}

				l353:
					{
						position354, tokenIndex354, depth354 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '_':
								if buffer[position] != rune('_') {
									goto l354
								}
								position++
								break
							case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
								if c := buffer[position]; c < rune('A') || c > rune('Z') {
									goto l354
								}
								position++
								break
							default:
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l354
								}
								position++
								break
							}
						}

						goto l353
					l354:
						position, tokenIndex, depth = position354, tokenIndex354, depth354
					}
					if buffer[position] != rune('.') {
						goto l350
					}
					position++
					{
						switch buffer[position] {
						case '_':
							if buffer[position] != rune('_') {
								goto l350
							}
							position++
							break
						case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l350
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l350
							}
							position++
							break
						}
					}

				l357:
					{
						position358, tokenIndex358, depth358 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '_':
								if buffer[position] != rune('_') {
									goto l358
								}
								position++
								break
							case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
								if c := buffer[position]; c < rune('A') || c > rune('Z') {
									goto l358
								}
								position++
								break
							default:
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l358
								}
								position++
								break
							}
						}

						goto l357
					l358:
						position, tokenIndex, depth = position358, tokenIndex358, depth358
					}
					depth--
					add(rulePegText, position352)
				}
				if !_rules[ruleAction17]() {
					goto l350
				}
				depth--
				add(ruleqname, position351)
			}
			return true
		l350:
			position, tokenIndex, depth = position350, tokenIndex350, depth350
			return false
		},
		/* 36 array <- <('[' ']' type Action18)> */
		func() bool {
			position361, tokenIndex361, depth361 := position, tokenIndex, depth
			{
				position362 := position
				depth++
				if buffer[position] != rune('[') {
					goto l361
				}
				position++
				if buffer[position] != rune(']') {
					goto l361
				}
				position++
				if !_rules[ruletype]() {
					goto l361
				}
				if !_rules[ruleAction18]() {
					goto l361
				}
				depth--
				add(rulearray, position362)
			}
			return true
		l361:
			position, tokenIndex, depth = position361, tokenIndex361, depth361
			return false
		},
		/* 37 map <- <(('m' / 'M') ('a' / 'A') ('p' / 'P') '[' isp* keytype isp* ']' type Action19)> */
		func() bool {
			position363, tokenIndex363, depth363 := position, tokenIndex, depth
			{
				position364 := position
				depth++
				{
					position365, tokenIndex365, depth365 := position, tokenIndex, depth
					if buffer[position] != rune('m') {
						goto l366
					}
					position++
					goto l365
				l366:
					position, tokenIndex, depth = position365, tokenIndex365, depth365
					if buffer[position] != rune('M') {
						goto l363
					}
					position++
				}
			l365:
				{
					position367, tokenIndex367, depth367 := position, tokenIndex, depth
					if buffer[position] != rune('a') {
						goto l368
					}
					position++
					goto l367
				l368:
					position, tokenIndex, depth = position367, tokenIndex367, depth367
					if buffer[position] != rune('A') {
						goto l363
					}
					position++
				}
			l367:
				{
					position369, tokenIndex369, depth369 := position, tokenIndex, depth
					if buffer[position] != rune('p') {
						goto l370
					}
					position++
					goto l369
				l370:
					position, tokenIndex, depth = position369, tokenIndex369, depth369
					if buffer[position] != rune('P') {
						goto l363
					}
					position++
				}
			l369:
				if buffer[position] != rune('[') {
					goto l363
				}
				position++
			l371:
				{
					position372, tokenIndex372, depth372 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l372
					}
					goto l371
				l372:
					position, tokenIndex, depth = position372, tokenIndex372, depth372
				}
				if !_rules[rulekeytype]() {
					goto l363
				}
			l373:
				{
					position374, tokenIndex374, depth374 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l374
					}
					goto l373
				l374:
					position, tokenIndex, depth = position374, tokenIndex374, depth374
				}
				if buffer[position] != rune(']') {
					goto l363
				}
				position++
				if !_rules[ruletype]() {
					goto l363
				}
				if !_rules[ruleAction19]() {
					goto l363
				}
				depth--
				add(rulemap, position364)
			}
			return true
		l363:
			position, tokenIndex, depth = position363, tokenIndex363, depth363
			return false
		},
		/* 38 keytype <- <(type Action20)> */
		func() bool {
			position375, tokenIndex375, depth375 := position, tokenIndex, depth
			{
				position376 := position
				depth++
				if !_rules[ruletype]() {
					goto l375
				}
				if !_rules[ruleAction20]() {
					goto l375
				}
				depth--
				add(rulekeytype, position376)
			}
			return true
		l375:
			position, tokenIndex, depth = position375, tokenIndex375, depth375
			return false
		},
		/* 39 pointer <- <('*' type Action21)> */
		func() bool {
			position377, tokenIndex377, depth377 := position, tokenIndex, depth
			{
				position378 := position
				depth++
				if buffer[position] != rune('*') {
					goto l377
				}
				position++
				if !_rules[ruletype]() {
					goto l377
				}
				if !_rules[ruleAction21]() {
					goto l377
				}
				depth--
				add(rulepointer, position378)
			}
			return true
		l377:
			position, tokenIndex, depth = position377, tokenIndex377, depth377
			return false
		},
		/* 40 captures <- <(isp* capture isp* (',' isp* capture isp*)* !.)> */
		func() bool {
			position379, tokenIndex379, depth379 := position, tokenIndex, depth
			{
				position380 := position
				depth++
			l381:
				{
					position382, tokenIndex382, depth382 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l382
					}
					goto l381
				l382:
					position, tokenIndex, depth = position382, tokenIndex382, depth382
				}
				if !_rules[rulecapture]() {
					goto l379
				}
			l383:
				{
					position384, tokenIndex384, depth384 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l384
					}
					goto l383
				l384:
					position, tokenIndex, depth = position384, tokenIndex384, depth384
				}
			l385:
				{
					position386, tokenIndex386, depth386 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l386
					}
					position++
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
					if !_rules[rulecapture]() {
						goto l386
					}
				l389:
					{
						position390, tokenIndex390, depth390 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l390
						}
						goto l389
					l390:
						position, tokenIndex, depth = position390, tokenIndex390, depth390
					}
					goto l385
				l386:
					position, tokenIndex, depth = position386, tokenIndex386, depth386
				}
				{
					position391, tokenIndex391, depth391 := position, tokenIndex, depth
					if !matchDot() {
						goto l391
					}
					goto l379
				l391:
					position, tokenIndex, depth = position391, tokenIndex391, depth391
				}
				depth--
				add(rulecaptures, position380)
			}
			return true
		l379:
			position, tokenIndex, depth = position379, tokenIndex379, depth379
			return false
		},
		/* 41 capture <- <(eventid isp* ':' handlername isp* mappings isp* tags Action22)> */
		func() bool {
			position392, tokenIndex392, depth392 := position, tokenIndex, depth
			{
				position393 := position
				depth++
				if !_rules[ruleeventid]() {
					goto l392
				}
			l394:
				{
					position395, tokenIndex395, depth395 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l395
					}
					goto l394
				l395:
					position, tokenIndex, depth = position395, tokenIndex395, depth395
				}
				if buffer[position] != rune(':') {
					goto l392
				}
				position++
				if !_rules[rulehandlername]() {
					goto l392
				}
			l396:
				{
					position397, tokenIndex397, depth397 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l397
					}
					goto l396
				l397:
					position, tokenIndex, depth = position397, tokenIndex397, depth397
				}
				if !_rules[rulemappings]() {
					goto l392
				}
			l398:
				{
					position399, tokenIndex399, depth399 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l399
					}
					goto l398
				l399:
					position, tokenIndex, depth = position399, tokenIndex399, depth399
				}
				if !_rules[ruletags]() {
					goto l392
				}
				if !_rules[ruleAction22]() {
					goto l392
				}
				depth--
				add(rulecapture, position393)
			}
			return true
		l392:
			position, tokenIndex, depth = position392, tokenIndex392, depth392
			return false
		},
		/* 42 handlername <- <(<identifier> Action23)> */
		func() bool {
			position400, tokenIndex400, depth400 := position, tokenIndex, depth
			{
				position401 := position
				depth++
				{
					position402 := position
					depth++
					if !_rules[ruleidentifier]() {
						goto l400
					}
					depth--
					add(rulePegText, position402)
				}
				if !_rules[ruleAction23]() {
					goto l400
				}
				depth--
				add(rulehandlername, position401)
			}
			return true
		l400:
			position, tokenIndex, depth = position400, tokenIndex400, depth400
			return false
		},
		/* 43 eventid <- <(<[a-z]+> Action24)> */
		func() bool {
			position403, tokenIndex403, depth403 := position, tokenIndex, depth
			{
				position404 := position
				depth++
				{
					position405 := position
					depth++
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l403
					}
					position++
				l406:
					{
						position407, tokenIndex407, depth407 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l407
						}
						position++
						goto l406
					l407:
						position, tokenIndex, depth = position407, tokenIndex407, depth407
					}
					depth--
					add(rulePegText, position405)
				}
				if !_rules[ruleAction24]() {
					goto l403
				}
				depth--
				add(ruleeventid, position404)
			}
			return true
		l403:
			position, tokenIndex, depth = position403, tokenIndex403, depth403
			return false
		},
		/* 44 mappings <- <('(' (isp* mapping isp* (',' isp* mapping isp*)*)? ')')?> */
		func() bool {
			{
				position409 := position
				depth++
				{
					position410, tokenIndex410, depth410 := position, tokenIndex, depth
					if buffer[position] != rune('(') {
						goto l410
					}
					position++
					{
						position412, tokenIndex412, depth412 := position, tokenIndex, depth
					l414:
						{
							position415, tokenIndex415, depth415 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l415
							}
							goto l414
						l415:
							position, tokenIndex, depth = position415, tokenIndex415, depth415
						}
						if !_rules[rulemapping]() {
							goto l412
						}
					l416:
						{
							position417, tokenIndex417, depth417 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l417
							}
							goto l416
						l417:
							position, tokenIndex, depth = position417, tokenIndex417, depth417
						}
					l418:
						{
							position419, tokenIndex419, depth419 := position, tokenIndex, depth
							if buffer[position] != rune(',') {
								goto l419
							}
							position++
						l420:
							{
								position421, tokenIndex421, depth421 := position, tokenIndex, depth
								if !_rules[ruleisp]() {
									goto l421
								}
								goto l420
							l421:
								position, tokenIndex, depth = position421, tokenIndex421, depth421
							}
							if !_rules[rulemapping]() {
								goto l419
							}
						l422:
							{
								position423, tokenIndex423, depth423 := position, tokenIndex, depth
								if !_rules[ruleisp]() {
									goto l423
								}
								goto l422
							l423:
								position, tokenIndex, depth = position423, tokenIndex423, depth423
							}
							goto l418
						l419:
							position, tokenIndex, depth = position419, tokenIndex419, depth419
						}
						goto l413
					l412:
						position, tokenIndex, depth = position412, tokenIndex412, depth412
					}
				l413:
					if buffer[position] != rune(')') {
						goto l410
					}
					position++
					goto l411
				l410:
					position, tokenIndex, depth = position410, tokenIndex410, depth410
				}
			l411:
				depth--
				add(rulemappings, position409)
			}
			return true
		},
		/* 45 mapping <- <(mappingname isp* '=' isp* bound Action25)> */
		func() bool {
			position424, tokenIndex424, depth424 := position, tokenIndex, depth
			{
				position425 := position
				depth++
				if !_rules[rulemappingname]() {
					goto l424
				}
			l426:
				{
					position427, tokenIndex427, depth427 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l427
					}
					goto l426
				l427:
					position, tokenIndex, depth = position427, tokenIndex427, depth427
				}
				if buffer[position] != rune('=') {
					goto l424
				}
				position++
			l428:
				{
					position429, tokenIndex429, depth429 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l429
					}
					goto l428
				l429:
					position, tokenIndex, depth = position429, tokenIndex429, depth429
				}
				if !_rules[rulebound]() {
					goto l424
				}
				if !_rules[ruleAction25]() {
					goto l424
				}
				depth--
				add(rulemapping, position425)
			}
			return true
		l424:
			position, tokenIndex, depth = position424, tokenIndex424, depth424
			return false
		},
		/* 46 mappingname <- <(<identifier> Action26)> */
		func() bool {
			position430, tokenIndex430, depth430 := position, tokenIndex, depth
			{
				position431 := position
				depth++
				{
					position432 := position
					depth++
					if !_rules[ruleidentifier]() {
						goto l430
					}
					depth--
					add(rulePegText, position432)
				}
				if !_rules[ruleAction26]() {
					goto l430
				}
				depth--
				add(rulemappingname, position431)
			}
			return true
		l430:
			position, tokenIndex, depth = position430, tokenIndex430, depth430
			return false
		},
		/* 47 tags <- <('{' isp* tag isp* (',' isp* tag isp*)* '}')?> */
		func() bool {
			{
				position434 := position
				depth++
				{
					position435, tokenIndex435, depth435 := position, tokenIndex, depth
					if buffer[position] != rune('{') {
						goto l435
					}
					position++
				l437:
					{
						position438, tokenIndex438, depth438 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l438
						}
						goto l437
					l438:
						position, tokenIndex, depth = position438, tokenIndex438, depth438
					}
					if !_rules[ruletag]() {
						goto l435
					}
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
				l441:
					{
						position442, tokenIndex442, depth442 := position, tokenIndex, depth
						if buffer[position] != rune(',') {
							goto l442
						}
						position++
					l443:
						{
							position444, tokenIndex444, depth444 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l444
							}
							goto l443
						l444:
							position, tokenIndex, depth = position444, tokenIndex444, depth444
						}
						if !_rules[ruletag]() {
							goto l442
						}
					l445:
						{
							position446, tokenIndex446, depth446 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l446
							}
							goto l445
						l446:
							position, tokenIndex, depth = position446, tokenIndex446, depth446
						}
						goto l441
					l442:
						position, tokenIndex, depth = position442, tokenIndex442, depth442
					}
					if buffer[position] != rune('}') {
						goto l435
					}
					position++
					goto l436
				l435:
					position, tokenIndex, depth = position435, tokenIndex435, depth435
				}
			l436:
				depth--
				add(ruletags, position434)
			}
			return true
		},
		/* 48 tag <- <(tagname ('(' (isp* tagarg isp* (',' isp* tagarg isp*)*)? ')')? Action27)> */
		func() bool {
			position447, tokenIndex447, depth447 := position, tokenIndex, depth
			{
				position448 := position
				depth++
				if !_rules[ruletagname]() {
					goto l447
				}
				{
					position449, tokenIndex449, depth449 := position, tokenIndex, depth
					if buffer[position] != rune('(') {
						goto l449
					}
					position++
					{
						position451, tokenIndex451, depth451 := position, tokenIndex, depth
					l453:
						{
							position454, tokenIndex454, depth454 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l454
							}
							goto l453
						l454:
							position, tokenIndex, depth = position454, tokenIndex454, depth454
						}
						if !_rules[ruletagarg]() {
							goto l451
						}
					l455:
						{
							position456, tokenIndex456, depth456 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l456
							}
							goto l455
						l456:
							position, tokenIndex, depth = position456, tokenIndex456, depth456
						}
					l457:
						{
							position458, tokenIndex458, depth458 := position, tokenIndex, depth
							if buffer[position] != rune(',') {
								goto l458
							}
							position++
						l459:
							{
								position460, tokenIndex460, depth460 := position, tokenIndex, depth
								if !_rules[ruleisp]() {
									goto l460
								}
								goto l459
							l460:
								position, tokenIndex, depth = position460, tokenIndex460, depth460
							}
							if !_rules[ruletagarg]() {
								goto l458
							}
						l461:
							{
								position462, tokenIndex462, depth462 := position, tokenIndex, depth
								if !_rules[ruleisp]() {
									goto l462
								}
								goto l461
							l462:
								position, tokenIndex, depth = position462, tokenIndex462, depth462
							}
							goto l457
						l458:
							position, tokenIndex, depth = position458, tokenIndex458, depth458
						}
						goto l452
					l451:
						position, tokenIndex, depth = position451, tokenIndex451, depth451
					}
				l452:
					if buffer[position] != rune(')') {
						goto l449
					}
					position++
					goto l450
				l449:
					position, tokenIndex, depth = position449, tokenIndex449, depth449
				}
			l450:
				if !_rules[ruleAction27]() {
					goto l447
				}
				depth--
				add(ruletag, position448)
			}
			return true
		l447:
			position, tokenIndex, depth = position447, tokenIndex447, depth447
			return false
		},
		/* 49 tagname <- <(<identifier> Action28)> */
		func() bool {
			position463, tokenIndex463, depth463 := position, tokenIndex, depth
			{
				position464 := position
				depth++
				{
					position465 := position
					depth++
					if !_rules[ruleidentifier]() {
						goto l463
					}
					depth--
					add(rulePegText, position465)
				}
				if !_rules[ruleAction28]() {
					goto l463
				}
				depth--
				add(ruletagname, position464)
			}
			return true
		l463:
			position, tokenIndex, depth = position463, tokenIndex463, depth463
			return false
		},
		/* 50 tagarg <- <(<identifier> Action29)> */
		func() bool {
			position466, tokenIndex466, depth466 := position, tokenIndex, depth
			{
				position467 := position
				depth++
				{
					position468 := position
					depth++
					if !_rules[ruleidentifier]() {
						goto l466
					}
					depth--
					add(rulePegText, position468)
				}
				if !_rules[ruleAction29]() {
					goto l466
				}
				depth--
				add(ruletagarg, position467)
			}
			return true
		l466:
			position, tokenIndex, depth = position466, tokenIndex466, depth466
			return false
		},
		/* 51 for <- <(isp* forVar isp* (',' isp* forVar isp*)? (':' '=') isp* (('r' / 'R') ('a' / 'A') ('n' / 'N') ('g' / 'G') ('e' / 'E')) isp+ expr isp* !.)> */
		func() bool {
			position469, tokenIndex469, depth469 := position, tokenIndex, depth
			{
				position470 := position
				depth++
			l471:
				{
					position472, tokenIndex472, depth472 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l472
					}
					goto l471
				l472:
					position, tokenIndex, depth = position472, tokenIndex472, depth472
				}
				if !_rules[ruleforVar]() {
					goto l469
				}
			l473:
				{
					position474, tokenIndex474, depth474 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l474
					}
					goto l473
				l474:
					position, tokenIndex, depth = position474, tokenIndex474, depth474
				}
				{
					position475, tokenIndex475, depth475 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l475
					}
					position++
				l477:
					{
						position478, tokenIndex478, depth478 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l478
						}
						goto l477
					l478:
						position, tokenIndex, depth = position478, tokenIndex478, depth478
					}
					if !_rules[ruleforVar]() {
						goto l475
					}
				l479:
					{
						position480, tokenIndex480, depth480 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l480
						}
						goto l479
					l480:
						position, tokenIndex, depth = position480, tokenIndex480, depth480
					}
					goto l476
				l475:
					position, tokenIndex, depth = position475, tokenIndex475, depth475
				}
			l476:
				if buffer[position] != rune(':') {
					goto l469
				}
				position++
				if buffer[position] != rune('=') {
					goto l469
				}
				position++
			l481:
				{
					position482, tokenIndex482, depth482 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l482
					}
					goto l481
				l482:
					position, tokenIndex, depth = position482, tokenIndex482, depth482
				}
				{
					position483, tokenIndex483, depth483 := position, tokenIndex, depth
					if buffer[position] != rune('r') {
						goto l484
					}
					position++
					goto l483
				l484:
					position, tokenIndex, depth = position483, tokenIndex483, depth483
					if buffer[position] != rune('R') {
						goto l469
					}
					position++
				}
			l483:
				{
					position485, tokenIndex485, depth485 := position, tokenIndex, depth
					if buffer[position] != rune('a') {
						goto l486
					}
					position++
					goto l485
				l486:
					position, tokenIndex, depth = position485, tokenIndex485, depth485
					if buffer[position] != rune('A') {
						goto l469
					}
					position++
				}
			l485:
				{
					position487, tokenIndex487, depth487 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l488
					}
					position++
					goto l487
				l488:
					position, tokenIndex, depth = position487, tokenIndex487, depth487
					if buffer[position] != rune('N') {
						goto l469
					}
					position++
				}
			l487:
				{
					position489, tokenIndex489, depth489 := position, tokenIndex, depth
					if buffer[position] != rune('g') {
						goto l490
					}
					position++
					goto l489
				l490:
					position, tokenIndex, depth = position489, tokenIndex489, depth489
					if buffer[position] != rune('G') {
						goto l469
					}
					position++
				}
			l489:
				{
					position491, tokenIndex491, depth491 := position, tokenIndex, depth
					if buffer[position] != rune('e') {
						goto l492
					}
					position++
					goto l491
				l492:
					position, tokenIndex, depth = position491, tokenIndex491, depth491
					if buffer[position] != rune('E') {
						goto l469
					}
					position++
				}
			l491:
				if !_rules[ruleisp]() {
					goto l469
				}
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
				if !_rules[ruleexpr]() {
					goto l469
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
				{
					position497, tokenIndex497, depth497 := position, tokenIndex, depth
					if !matchDot() {
						goto l497
					}
					goto l469
				l497:
					position, tokenIndex, depth = position497, tokenIndex497, depth497
				}
				depth--
				add(rulefor, position470)
			}
			return true
		l469:
			position, tokenIndex, depth = position469, tokenIndex469, depth469
			return false
		},
		/* 52 forVar <- <(<identifier> Action30)> */
		func() bool {
			position498, tokenIndex498, depth498 := position, tokenIndex, depth
			{
				position499 := position
				depth++
				{
					position500 := position
					depth++
					if !_rules[ruleidentifier]() {
						goto l498
					}
					depth--
					add(rulePegText, position500)
				}
				if !_rules[ruleAction30]() {
					goto l498
				}
				depth--
				add(ruleforVar, position499)
			}
			return true
		l498:
			position, tokenIndex, depth = position498, tokenIndex498, depth498
			return false
		},
		/* 53 handlers <- <(isp* (fsep isp*)* handler isp* ((fsep isp*)+ handler isp*)* (fsep isp*)* !.)> */
		func() bool {
			position501, tokenIndex501, depth501 := position, tokenIndex, depth
			{
				position502 := position
				depth++
			l503:
				{
					position504, tokenIndex504, depth504 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l504
					}
					goto l503
				l504:
					position, tokenIndex, depth = position504, tokenIndex504, depth504
				}
			l505:
				{
					position506, tokenIndex506, depth506 := position, tokenIndex, depth
					if !_rules[rulefsep]() {
						goto l506
					}
				l507:
					{
						position508, tokenIndex508, depth508 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l508
						}
						goto l507
					l508:
						position, tokenIndex, depth = position508, tokenIndex508, depth508
					}
					goto l505
				l506:
					position, tokenIndex, depth = position506, tokenIndex506, depth506
				}
				if !_rules[rulehandler]() {
					goto l501
				}
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
			l511:
				{
					position512, tokenIndex512, depth512 := position, tokenIndex, depth
					if !_rules[rulefsep]() {
						goto l512
					}
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
				l513:
					{
						position514, tokenIndex514, depth514 := position, tokenIndex, depth
						if !_rules[rulefsep]() {
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
					if !_rules[rulehandler]() {
						goto l512
					}
				l519:
					{
						position520, tokenIndex520, depth520 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l520
						}
						goto l519
					l520:
						position, tokenIndex, depth = position520, tokenIndex520, depth520
					}
					goto l511
				l512:
					position, tokenIndex, depth = position512, tokenIndex512, depth512
				}
			l521:
				{
					position522, tokenIndex522, depth522 := position, tokenIndex, depth
					if !_rules[rulefsep]() {
						goto l522
					}
				l523:
					{
						position524, tokenIndex524, depth524 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l524
						}
						goto l523
					l524:
						position, tokenIndex, depth = position524, tokenIndex524, depth524
					}
					goto l521
				l522:
					position, tokenIndex, depth = position522, tokenIndex522, depth522
				}
				{
					position525, tokenIndex525, depth525 := position, tokenIndex, depth
					if !matchDot() {
						goto l525
					}
					goto l501
				l525:
					position, tokenIndex, depth = position525, tokenIndex525, depth525
				}
				depth--
				add(rulehandlers, position502)
			}
			return true
		l501:
			position, tokenIndex, depth = position501, tokenIndex501, depth501
			return false
		},
		/* 54 handler <- <(handlername '(' isp* (param isp* (',' isp* param isp*)*)? ')' (isp* type)? Action31)> */
		func() bool {
			position526, tokenIndex526, depth526 := position, tokenIndex, depth
			{
				position527 := position
				depth++
				if !_rules[rulehandlername]() {
					goto l526
				}
				if buffer[position] != rune('(') {
					goto l526
				}
				position++
			l528:
				{
					position529, tokenIndex529, depth529 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l529
					}
					goto l528
				l529:
					position, tokenIndex, depth = position529, tokenIndex529, depth529
				}
				{
					position530, tokenIndex530, depth530 := position, tokenIndex, depth
					if !_rules[ruleparam]() {
						goto l530
					}
				l532:
					{
						position533, tokenIndex533, depth533 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l533
						}
						goto l532
					l533:
						position, tokenIndex, depth = position533, tokenIndex533, depth533
					}
				l534:
					{
						position535, tokenIndex535, depth535 := position, tokenIndex, depth
						if buffer[position] != rune(',') {
							goto l535
						}
						position++
					l536:
						{
							position537, tokenIndex537, depth537 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l537
							}
							goto l536
						l537:
							position, tokenIndex, depth = position537, tokenIndex537, depth537
						}
						if !_rules[ruleparam]() {
							goto l535
						}
					l538:
						{
							position539, tokenIndex539, depth539 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l539
							}
							goto l538
						l539:
							position, tokenIndex, depth = position539, tokenIndex539, depth539
						}
						goto l534
					l535:
						position, tokenIndex, depth = position535, tokenIndex535, depth535
					}
					goto l531
				l530:
					position, tokenIndex, depth = position530, tokenIndex530, depth530
				}
			l531:
				if buffer[position] != rune(')') {
					goto l526
				}
				position++
				{
					position540, tokenIndex540, depth540 := position, tokenIndex, depth
				l542:
					{
						position543, tokenIndex543, depth543 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l543
						}
						goto l542
					l543:
						position, tokenIndex, depth = position543, tokenIndex543, depth543
					}
					if !_rules[ruletype]() {
						goto l540
					}
					goto l541
				l540:
					position, tokenIndex, depth = position540, tokenIndex540, depth540
				}
			l541:
				if !_rules[ruleAction31]() {
					goto l526
				}
				depth--
				add(rulehandler, position527)
			}
			return true
		l526:
			position, tokenIndex, depth = position526, tokenIndex526, depth526
			return false
		},
		/* 55 param <- <(tagname isp+ type Action32)> */
		func() bool {
			position544, tokenIndex544, depth544 := position, tokenIndex, depth
			{
				position545 := position
				depth++
				if !_rules[ruletagname]() {
					goto l544
				}
				if !_rules[ruleisp]() {
					goto l544
				}
			l546:
				{
					position547, tokenIndex547, depth547 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l547
					}
					goto l546
				l547:
					position, tokenIndex, depth = position547, tokenIndex547, depth547
				}
				if !_rules[ruletype]() {
					goto l544
				}
				if !_rules[ruleAction32]() {
					goto l544
				}
				depth--
				add(ruleparam, position545)
			}
			return true
		l544:
			position, tokenIndex, depth = position544, tokenIndex544, depth544
			return false
		},
		/* 56 cparams <- <(isp* (cparam isp* (',' isp* cparam isp*)*)? !.)> */
		func() bool {
			position548, tokenIndex548, depth548 := position, tokenIndex, depth
			{
				position549 := position
				depth++
			l550:
				{
					position551, tokenIndex551, depth551 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l551
					}
					goto l550
				l551:
					position, tokenIndex, depth = position551, tokenIndex551, depth551
				}
				{
					position552, tokenIndex552, depth552 := position, tokenIndex, depth
					if !_rules[rulecparam]() {
						goto l552
					}
				l554:
					{
						position555, tokenIndex555, depth555 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l555
						}
						goto l554
					l555:
						position, tokenIndex, depth = position555, tokenIndex555, depth555
					}
				l556:
					{
						position557, tokenIndex557, depth557 := position, tokenIndex, depth
						if buffer[position] != rune(',') {
							goto l557
						}
						position++
					l558:
						{
							position559, tokenIndex559, depth559 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l559
							}
							goto l558
						l559:
							position, tokenIndex, depth = position559, tokenIndex559, depth559
						}
						if !_rules[rulecparam]() {
							goto l557
						}
					l560:
						{
							position561, tokenIndex561, depth561 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l561
							}
							goto l560
						l561:
							position, tokenIndex, depth = position561, tokenIndex561, depth561
						}
						goto l556
					l557:
						position, tokenIndex, depth = position557, tokenIndex557, depth557
					}
					goto l553
				l552:
					position, tokenIndex, depth = position552, tokenIndex552, depth552
				}
			l553:
				{
					position562, tokenIndex562, depth562 := position, tokenIndex, depth
					if !matchDot() {
						goto l562
					}
					goto l548
				l562:
					position, tokenIndex, depth = position562, tokenIndex562, depth562
				}
				depth--
				add(rulecparams, position549)
			}
			return true
		l548:
			position, tokenIndex, depth = position548, tokenIndex548, depth548
			return false
		},
		/* 57 cparam <- <((var isp+)? tagname isp+ type Action33)> */
		func() bool {
			position563, tokenIndex563, depth563 := position, tokenIndex, depth
			{
				position564 := position
				depth++
				{
					position565, tokenIndex565, depth565 := position, tokenIndex, depth
					if !_rules[rulevar]() {
						goto l565
					}
					if !_rules[ruleisp]() {
						goto l565
					}
				l567:
					{
						position568, tokenIndex568, depth568 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l568
						}
						goto l567
					l568:
						position, tokenIndex, depth = position568, tokenIndex568, depth568
					}
					goto l566
				l565:
					position, tokenIndex, depth = position565, tokenIndex565, depth565
				}
			l566:
				if !_rules[ruletagname]() {
					goto l563
				}
				if !_rules[ruleisp]() {
					goto l563
				}
			l569:
				{
					position570, tokenIndex570, depth570 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l570
					}
					goto l569
				l570:
					position, tokenIndex, depth = position570, tokenIndex570, depth570
				}
				if !_rules[ruletype]() {
					goto l563
				}
				if !_rules[ruleAction33]() {
					goto l563
				}
				depth--
				add(rulecparam, position564)
			}
			return true
		l563:
			position, tokenIndex, depth = position563, tokenIndex563, depth563
			return false
		},
		/* 58 var <- <(('v' / 'V') ('a' / 'A') ('r' / 'R') Action34)> */
		func() bool {
			position571, tokenIndex571, depth571 := position, tokenIndex, depth
			{
				position572 := position
				depth++
				{
					position573, tokenIndex573, depth573 := position, tokenIndex, depth
					if buffer[position] != rune('v') {
						goto l574
					}
					position++
					goto l573
				l574:
					position, tokenIndex, depth = position573, tokenIndex573, depth573
					if buffer[position] != rune('V') {
						goto l571
					}
					position++
				}
			l573:
				{
					position575, tokenIndex575, depth575 := position, tokenIndex, depth
					if buffer[position] != rune('a') {
						goto l576
					}
					position++
					goto l575
				l576:
					position, tokenIndex, depth = position575, tokenIndex575, depth575
					if buffer[position] != rune('A') {
						goto l571
					}
					position++
				}
			l575:
				{
					position577, tokenIndex577, depth577 := position, tokenIndex, depth
					if buffer[position] != rune('r') {
						goto l578
					}
					position++
					goto l577
				l578:
					position, tokenIndex, depth = position577, tokenIndex577, depth577
					if buffer[position] != rune('R') {
						goto l571
					}
					position++
				}
			l577:
				if !_rules[ruleAction34]() {
					goto l571
				}
				depth--
				add(rulevar, position572)
			}
			return true
		l571:
			position, tokenIndex, depth = position571, tokenIndex571, depth571
			return false
		},
		/* 59 args <- <(isp* arg isp* (',' isp* arg isp*)* !.)> */
		func() bool {
			position579, tokenIndex579, depth579 := position, tokenIndex, depth
			{
				position580 := position
				depth++
			l581:
				{
					position582, tokenIndex582, depth582 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l582
					}
					goto l581
				l582:
					position, tokenIndex, depth = position582, tokenIndex582, depth582
				}
				if !_rules[rulearg]() {
					goto l579
				}
			l583:
				{
					position584, tokenIndex584, depth584 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l584
					}
					goto l583
				l584:
					position, tokenIndex, depth = position584, tokenIndex584, depth584
				}
			l585:
				{
					position586, tokenIndex586, depth586 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l586
					}
					position++
				l587:
					{
						position588, tokenIndex588, depth588 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l588
						}
						goto l587
					l588:
						position, tokenIndex, depth = position588, tokenIndex588, depth588
					}
					if !_rules[rulearg]() {
						goto l586
					}
				l589:
					{
						position590, tokenIndex590, depth590 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l590
						}
						goto l589
					l590:
						position, tokenIndex, depth = position590, tokenIndex590, depth590
					}
					goto l585
				l586:
					position, tokenIndex, depth = position586, tokenIndex586, depth586
				}
				{
					position591, tokenIndex591, depth591 := position, tokenIndex, depth
					if !matchDot() {
						goto l591
					}
					goto l579
				l591:
					position, tokenIndex, depth = position591, tokenIndex591, depth591
				}
				depth--
				add(ruleargs, position580)
			}
			return true
		l579:
			position, tokenIndex, depth = position579, tokenIndex579, depth579
			return false
		},
		/* 60 arg <- <(expr Action35)> */
		func() bool {
			position592, tokenIndex592, depth592 := position, tokenIndex, depth
			{
				position593 := position
				depth++
				if !_rules[ruleexpr]() {
					goto l592
				}
				if !_rules[ruleAction35]() {
					goto l592
				}
				depth--
				add(rulearg, position593)
			}
			return true
		l592:
			position, tokenIndex, depth = position592, tokenIndex592, depth592
			return false
		},
		/* 61 imports <- <(isp* (fsep isp*)* import isp* (fsep isp* (fsep isp*)* import isp*)* (fsep isp*)* !.)> */
		func() bool {
			position594, tokenIndex594, depth594 := position, tokenIndex, depth
			{
				position595 := position
				depth++
			l596:
				{
					position597, tokenIndex597, depth597 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l597
					}
					goto l596
				l597:
					position, tokenIndex, depth = position597, tokenIndex597, depth597
				}
			l598:
				{
					position599, tokenIndex599, depth599 := position, tokenIndex, depth
					if !_rules[rulefsep]() {
						goto l599
					}
				l600:
					{
						position601, tokenIndex601, depth601 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l601
						}
						goto l600
					l601:
						position, tokenIndex, depth = position601, tokenIndex601, depth601
					}
					goto l598
				l599:
					position, tokenIndex, depth = position599, tokenIndex599, depth599
				}
				if !_rules[ruleimport]() {
					goto l594
				}
			l602:
				{
					position603, tokenIndex603, depth603 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l603
					}
					goto l602
				l603:
					position, tokenIndex, depth = position603, tokenIndex603, depth603
				}
			l604:
				{
					position605, tokenIndex605, depth605 := position, tokenIndex, depth
					if !_rules[rulefsep]() {
						goto l605
					}
				l606:
					{
						position607, tokenIndex607, depth607 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l607
						}
						goto l606
					l607:
						position, tokenIndex, depth = position607, tokenIndex607, depth607
					}
				l608:
					{
						position609, tokenIndex609, depth609 := position, tokenIndex, depth
						if !_rules[rulefsep]() {
							goto l609
						}
					l610:
						{
							position611, tokenIndex611, depth611 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l611
							}
							goto l610
						l611:
							position, tokenIndex, depth = position611, tokenIndex611, depth611
						}
						goto l608
					l609:
						position, tokenIndex, depth = position609, tokenIndex609, depth609
					}
					if !_rules[ruleimport]() {
						goto l605
					}
				l612:
					{
						position613, tokenIndex613, depth613 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l613
						}
						goto l612
					l613:
						position, tokenIndex, depth = position613, tokenIndex613, depth613
					}
					goto l604
				l605:
					position, tokenIndex, depth = position605, tokenIndex605, depth605
				}
			l614:
				{
					position615, tokenIndex615, depth615 := position, tokenIndex, depth
					if !_rules[rulefsep]() {
						goto l615
					}
				l616:
					{
						position617, tokenIndex617, depth617 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l617
						}
						goto l616
					l617:
						position, tokenIndex, depth = position617, tokenIndex617, depth617
					}
					goto l614
				l615:
					position, tokenIndex, depth = position615, tokenIndex615, depth615
				}
				{
					position618, tokenIndex618, depth618 := position, tokenIndex, depth
					if !matchDot() {
						goto l618
					}
					goto l594
				l618:
					position, tokenIndex, depth = position618, tokenIndex618, depth618
				}
				depth--
				add(ruleimports, position595)
			}
			return true
		l594:
			position, tokenIndex, depth = position594, tokenIndex594, depth594
			return false
		},
		/* 62 import <- <((tagname isp+)? '"' <(!'"' .)*> '"' Action36)> */
		func() bool {
			position619, tokenIndex619, depth619 := position, tokenIndex, depth
			{
				position620 := position
				depth++
				{
					position621, tokenIndex621, depth621 := position, tokenIndex, depth
					if !_rules[ruletagname]() {
						goto l621
					}
					if !_rules[ruleisp]() {
						goto l621
					}
				l623:
					{
						position624, tokenIndex624, depth624 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l624
						}
						goto l623
					l624:
						position, tokenIndex, depth = position624, tokenIndex624, depth624
					}
					goto l622
				l621:
					position, tokenIndex, depth = position621, tokenIndex621, depth621
				}
			l622:
				if buffer[position] != rune('"') {
					goto l619
				}
				position++
				{
					position625 := position
					depth++
				l626:
					{
						position627, tokenIndex627, depth627 := position, tokenIndex, depth
						{
							position628, tokenIndex628, depth628 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l628
							}
							position++
							goto l627
						l628:
							position, tokenIndex, depth = position628, tokenIndex628, depth628
						}
						if !matchDot() {
							goto l627
						}
						goto l626
					l627:
						position, tokenIndex, depth = position627, tokenIndex627, depth627
					}
					depth--
					add(rulePegText, position625)
				}
				if buffer[position] != rune('"') {
					goto l619
				}
				position++
				if !_rules[ruleAction36]() {
					goto l619
				}
				depth--
				add(ruleimport, position620)
			}
			return true
		l619:
			position, tokenIndex, depth = position619, tokenIndex619, depth619
			return false
		},
		/* 64 Action0 <- <{
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
		/* 66 Action1 <- <{
			p.goVal.Name = buffer[begin:end]
		}> */
		func() bool {
			{
				add(ruleAction1, position)
			}
			return true
		},
		/* 67 Action2 <- <{
			p.goVal.Type = p.valuetype
			p.valuetype = nil
		}> */
		func() bool {
			{
				add(ruleAction2, position)
			}
			return true
		},
		/* 68 Action3 <- <{
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
		/* 69 Action4 <- <{
			p.bv.Kind = data.BoundSelf
		}> */
		func() bool {
			{
				add(ruleAction4, position)
			}
			return true
		},
		/* 70 Action5 <- <{
			p.bv.Kind = data.BoundDataset
		}> */
		func() bool {
			{
				add(ruleAction5, position)
			}
			return true
		},
		/* 71 Action6 <- <{
			p.bv.Kind = data.BoundProperty
		}> */
		func() bool {
			{
				add(ruleAction6, position)
			}
			return true
		},
		/* 72 Action7 <- <{
			p.bv.Kind = data.BoundStyle
		}> */
		func() bool {
			{
				add(ruleAction7, position)
			}
			return true
		},
		/* 73 Action8 <- <{
			p.bv.Kind = data.BoundClass
		}> */
		func() bool {
			{
				add(ruleAction8, position)
			}
			return true
		},
		/* 74 Action9 <- <{
			p.bv.Kind = data.BoundFormValue
		}> */
		func() bool {
			{
				add(ruleAction9, position)
			}
			return true
		},
		/* 75 Action10 <- <{
			p.bv.Kind = data.BoundEventValue
			if len(p.bv.IDs) == 0 {
				p.bv.IDs = append(p.bv.IDs, "")
			}
		}> */
		func() bool {
			{
				add(ruleAction10, position)
			}
			return true
		},
		/* 76 Action11 <- <{
			p.bv.IDs = append(p.bv.IDs, buffer[begin:end])
		}> */
		func() bool {
			{
				add(ruleAction11, position)
			}
			return true
		},
		/* 77 Action12 <- <{
			p.bv.IDs = append(p.bv.IDs, buffer[begin:end])
		}> */
		func() bool {
			{
				add(ruleAction12, position)
			}
			return true
		},
		/* 78 Action13 <- <{
			p.expr = buffer[begin:end]
		}> */
		func() bool {
			{
				add(ruleAction13, position)
			}
			return true
		},
		/* 79 Action14 <- <{
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
				add(ruleAction14, position)
			}
			return true
		},
		/* 80 Action15 <- <{
			p.names = append(p.names, buffer[begin:end])
		}> */
		func() bool {
			{
				add(ruleAction15, position)
			}
			return true
		},
		/* 81 Action16 <- <{
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
				add(ruleAction16, position)
			}
			return true
		},
		/* 82 Action17 <- <{
			name := buffer[begin:end]
			if name == "js.Value" {
				p.valuetype = &data.ParamType{Kind: data.JSValueType}
			} else {
				p.valuetype = &data.ParamType{Kind: data.NamedType, Name: name}
			}
		}> */
		func() bool {
			{
				add(ruleAction17, position)
			}
			return true
		},
		/* 83 Action18 <- <{
			p.valuetype = &data.ParamType{Kind: data.ArrayType, ValueType: p.valuetype}
		}> */
		func() bool {
			{
				add(ruleAction18, position)
			}
			return true
		},
		/* 84 Action19 <- <{
			p.valuetype = &data.ParamType{Kind: data.MapType, KeyType: p.keytype, ValueType: p.valuetype}
		}> */
		func() bool {
			{
				add(ruleAction19, position)
			}
			return true
		},
		/* 85 Action20 <- <{
			p.keytype = p.valuetype
		}> */
		func() bool {
			{
				add(ruleAction20, position)
			}
			return true
		},
		/* 86 Action21 <- <{
			p.valuetype = &data.ParamType{Kind: data.PointerType, ValueType: p.valuetype}
		}> */
		func() bool {
			{
				add(ruleAction21, position)
			}
			return true
		},
		/* 87 Action22 <- <{
			p.eventMappings = append(p.eventMappings, data.UnboundEventMapping{
				Event: p.expr, Handler: p.handlername, ParamMappings: p.paramMappings,
				Handling: p.eventHandling})
			p.eventHandling = data.AutoPreventDefault
			p.expr = ""
			p.paramMappings = make(map[string]data.BoundValue)
		}> */
		func() bool {
			{
				add(ruleAction22, position)
			}
			return true
		},
		/* 88 Action23 <- <{
			p.handlername = buffer[begin:end]
		}> */
		func() bool {
			{
				add(ruleAction23, position)
			}
			return true
		},
		/* 89 Action24 <- <{
			p.expr = buffer[begin:end]
		}> */
		func() bool {
			{
				add(ruleAction24, position)
			}
			return true
		},
		/* 90 Action25 <- <{
			if _, ok := p.paramMappings[p.tagname]; ok {
				p.err = errors.New("duplicate param: " + p.tagname)
				return
			}
			p.paramMappings[p.tagname] = p.bv
			p.bv.IDs = nil
		}> */
		func() bool {
			{
				add(ruleAction25, position)
			}
			return true
		},
		/* 91 Action26 <- <{
			p.tagname = buffer[begin:end]
		}> */
		func() bool {
			{
				add(ruleAction26, position)
			}
			return true
		},
		/* 92 Action27 <- <{
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
				add(ruleAction27, position)
			}
			return true
		},
		/* 93 Action28 <- <{
			p.tagname = buffer[begin:end]
		}> */
		func() bool {
			{
				add(ruleAction28, position)
			}
			return true
		},
		/* 94 Action29 <- <{
			p.names = append(p.names, buffer[begin:end])
		}> */
		func() bool {
			{
				add(ruleAction29, position)
			}
			return true
		},
		/* 95 Action30 <- <{
			p.names = append(p.names, buffer[begin:end])
		}> */
		func() bool {
			{
				add(ruleAction30, position)
			}
			return true
		},
		/* 96 Action31 <- <{
			p.handlers = append(p.handlers, HandlerSpec{
				Name: p.handlername, Params: p.params, Returns: p.valuetype})
			p.valuetype = nil
			p.params = nil
		}> */
		func() bool {
			{
				add(ruleAction31, position)
			}
			return true
		},
		/* 97 Action32 <- <{
			for _, para := range p.params {
				if para.Name == p.tagname {
					p.err = errors.New("duplicate param name: " + para.Name)
					return
				}
			}

			p.params = append(p.params, data.Param{Name: p.tagname, Type: p.valuetype})
			p.valuetype = nil
		}> */
		func() bool {
			{
				add(ruleAction32, position)
			}
			return true
		},
		/* 98 Action33 <- <{
			p.cParams = append(p.cParams, data.ComponentParam{
				Name: p.tagname, Type: *p.valuetype, IsVar: p.isVar})
			p.valuetype = nil
			p.isVar = false
		}> */
		func() bool {
			{
				add(ruleAction33, position)
			}
			return true
		},
		/* 99 Action34 <- <{
			p.isVar = true
		}> */
		func() bool {
			{
				add(ruleAction34, position)
			}
			return true
		},
		/* 100 Action35 <- <{
		  p.names = append(p.names, p.expr)
		}> */
		func() bool {
			{
				add(ruleAction35, position)
			}
			return true
		},
		/* 101 Action36 <- <{
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
				add(ruleAction36, position)
			}
			return true
		},
	}
	p.rules = _rules
}
