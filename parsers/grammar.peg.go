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
	ruledata
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
	"data",
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

			p.bv.Kind = data.BoundData

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
		/* 8 bound <- <(self / data / prop / style / class / form / event)> */
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
					if !_rules[ruledata]() {
						goto l76
					}
					goto l74
				l76:
					position, tokenIndex, depth = position74, tokenIndex74, depth74
					if !_rules[ruleprop]() {
						goto l77
					}
					goto l74
				l77:
					position, tokenIndex, depth = position74, tokenIndex74, depth74
					if !_rules[rulestyle]() {
						goto l78
					}
					goto l74
				l78:
					position, tokenIndex, depth = position74, tokenIndex74, depth74
					if !_rules[ruleclass]() {
						goto l79
					}
					goto l74
				l79:
					position, tokenIndex, depth = position74, tokenIndex74, depth74
					if !_rules[ruleform]() {
						goto l80
					}
					goto l74
				l80:
					position, tokenIndex, depth = position74, tokenIndex74, depth74
					if !_rules[ruleevent]() {
						goto l72
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
			position81, tokenIndex81, depth81 := position, tokenIndex, depth
			{
				position82 := position
				depth++
				{
					position83, tokenIndex83, depth83 := position, tokenIndex, depth
					if buffer[position] != rune('s') {
						goto l84
					}
					position++
					goto l83
				l84:
					position, tokenIndex, depth = position83, tokenIndex83, depth83
					if buffer[position] != rune('S') {
						goto l81
					}
					position++
				}
			l83:
				{
					position85, tokenIndex85, depth85 := position, tokenIndex, depth
					if buffer[position] != rune('e') {
						goto l86
					}
					position++
					goto l85
				l86:
					position, tokenIndex, depth = position85, tokenIndex85, depth85
					if buffer[position] != rune('E') {
						goto l81
					}
					position++
				}
			l85:
				{
					position87, tokenIndex87, depth87 := position, tokenIndex, depth
					if buffer[position] != rune('l') {
						goto l88
					}
					position++
					goto l87
				l88:
					position, tokenIndex, depth = position87, tokenIndex87, depth87
					if buffer[position] != rune('L') {
						goto l81
					}
					position++
				}
			l87:
				{
					position89, tokenIndex89, depth89 := position, tokenIndex, depth
					if buffer[position] != rune('f') {
						goto l90
					}
					position++
					goto l89
				l90:
					position, tokenIndex, depth = position89, tokenIndex89, depth89
					if buffer[position] != rune('F') {
						goto l81
					}
					position++
				}
			l89:
			l91:
				{
					position92, tokenIndex92, depth92 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l92
					}
					goto l91
				l92:
					position, tokenIndex, depth = position92, tokenIndex92, depth92
				}
				if buffer[position] != rune('(') {
					goto l81
				}
				position++
			l93:
				{
					position94, tokenIndex94, depth94 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l94
					}
					goto l93
				l94:
					position, tokenIndex, depth = position94, tokenIndex94, depth94
				}
				if buffer[position] != rune(')') {
					goto l81
				}
				position++
				if !_rules[ruleAction4]() {
					goto l81
				}
				depth--
				add(ruleself, position82)
			}
			return true
		l81:
			position, tokenIndex, depth = position81, tokenIndex81, depth81
			return false
		},
		/* 10 data <- <(('d' / 'D') ('a' / 'A') ('t' / 'T') ('a' / 'A') isp* '(' isp* htmlid isp* ')' Action5)> */
		func() bool {
			position95, tokenIndex95, depth95 := position, tokenIndex, depth
			{
				position96 := position
				depth++
				{
					position97, tokenIndex97, depth97 := position, tokenIndex, depth
					if buffer[position] != rune('d') {
						goto l98
					}
					position++
					goto l97
				l98:
					position, tokenIndex, depth = position97, tokenIndex97, depth97
					if buffer[position] != rune('D') {
						goto l95
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
						goto l95
					}
					position++
				}
			l99:
				{
					position101, tokenIndex101, depth101 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l102
					}
					position++
					goto l101
				l102:
					position, tokenIndex, depth = position101, tokenIndex101, depth101
					if buffer[position] != rune('T') {
						goto l95
					}
					position++
				}
			l101:
				{
					position103, tokenIndex103, depth103 := position, tokenIndex, depth
					if buffer[position] != rune('a') {
						goto l104
					}
					position++
					goto l103
				l104:
					position, tokenIndex, depth = position103, tokenIndex103, depth103
					if buffer[position] != rune('A') {
						goto l95
					}
					position++
				}
			l103:
			l105:
				{
					position106, tokenIndex106, depth106 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l106
					}
					goto l105
				l106:
					position, tokenIndex, depth = position106, tokenIndex106, depth106
				}
				if buffer[position] != rune('(') {
					goto l95
				}
				position++
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
				if !_rules[rulehtmlid]() {
					goto l95
				}
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
				if buffer[position] != rune(')') {
					goto l95
				}
				position++
				if !_rules[ruleAction5]() {
					goto l95
				}
				depth--
				add(ruledata, position96)
			}
			return true
		l95:
			position, tokenIndex, depth = position95, tokenIndex95, depth95
			return false
		},
		/* 11 prop <- <(('p' / 'P') ('r' / 'R') ('o' / 'O') ('p' / 'P') isp* '(' isp* htmlid isp* ')' Action6)> */
		func() bool {
			position111, tokenIndex111, depth111 := position, tokenIndex, depth
			{
				position112 := position
				depth++
				{
					position113, tokenIndex113, depth113 := position, tokenIndex, depth
					if buffer[position] != rune('p') {
						goto l114
					}
					position++
					goto l113
				l114:
					position, tokenIndex, depth = position113, tokenIndex113, depth113
					if buffer[position] != rune('P') {
						goto l111
					}
					position++
				}
			l113:
				{
					position115, tokenIndex115, depth115 := position, tokenIndex, depth
					if buffer[position] != rune('r') {
						goto l116
					}
					position++
					goto l115
				l116:
					position, tokenIndex, depth = position115, tokenIndex115, depth115
					if buffer[position] != rune('R') {
						goto l111
					}
					position++
				}
			l115:
				{
					position117, tokenIndex117, depth117 := position, tokenIndex, depth
					if buffer[position] != rune('o') {
						goto l118
					}
					position++
					goto l117
				l118:
					position, tokenIndex, depth = position117, tokenIndex117, depth117
					if buffer[position] != rune('O') {
						goto l111
					}
					position++
				}
			l117:
				{
					position119, tokenIndex119, depth119 := position, tokenIndex, depth
					if buffer[position] != rune('p') {
						goto l120
					}
					position++
					goto l119
				l120:
					position, tokenIndex, depth = position119, tokenIndex119, depth119
					if buffer[position] != rune('P') {
						goto l111
					}
					position++
				}
			l119:
			l121:
				{
					position122, tokenIndex122, depth122 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l122
					}
					goto l121
				l122:
					position, tokenIndex, depth = position122, tokenIndex122, depth122
				}
				if buffer[position] != rune('(') {
					goto l111
				}
				position++
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
				if !_rules[rulehtmlid]() {
					goto l111
				}
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
				if buffer[position] != rune(')') {
					goto l111
				}
				position++
				if !_rules[ruleAction6]() {
					goto l111
				}
				depth--
				add(ruleprop, position112)
			}
			return true
		l111:
			position, tokenIndex, depth = position111, tokenIndex111, depth111
			return false
		},
		/* 12 style <- <(('s' / 'S') ('t' / 'T') ('y' / 'Y') ('l' / 'L') ('e' / 'E') isp* '(' isp* htmlid isp* ')' Action7)> */
		func() bool {
			position127, tokenIndex127, depth127 := position, tokenIndex, depth
			{
				position128 := position
				depth++
				{
					position129, tokenIndex129, depth129 := position, tokenIndex, depth
					if buffer[position] != rune('s') {
						goto l130
					}
					position++
					goto l129
				l130:
					position, tokenIndex, depth = position129, tokenIndex129, depth129
					if buffer[position] != rune('S') {
						goto l127
					}
					position++
				}
			l129:
				{
					position131, tokenIndex131, depth131 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l132
					}
					position++
					goto l131
				l132:
					position, tokenIndex, depth = position131, tokenIndex131, depth131
					if buffer[position] != rune('T') {
						goto l127
					}
					position++
				}
			l131:
				{
					position133, tokenIndex133, depth133 := position, tokenIndex, depth
					if buffer[position] != rune('y') {
						goto l134
					}
					position++
					goto l133
				l134:
					position, tokenIndex, depth = position133, tokenIndex133, depth133
					if buffer[position] != rune('Y') {
						goto l127
					}
					position++
				}
			l133:
				{
					position135, tokenIndex135, depth135 := position, tokenIndex, depth
					if buffer[position] != rune('l') {
						goto l136
					}
					position++
					goto l135
				l136:
					position, tokenIndex, depth = position135, tokenIndex135, depth135
					if buffer[position] != rune('L') {
						goto l127
					}
					position++
				}
			l135:
				{
					position137, tokenIndex137, depth137 := position, tokenIndex, depth
					if buffer[position] != rune('e') {
						goto l138
					}
					position++
					goto l137
				l138:
					position, tokenIndex, depth = position137, tokenIndex137, depth137
					if buffer[position] != rune('E') {
						goto l127
					}
					position++
				}
			l137:
			l139:
				{
					position140, tokenIndex140, depth140 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l140
					}
					goto l139
				l140:
					position, tokenIndex, depth = position140, tokenIndex140, depth140
				}
				if buffer[position] != rune('(') {
					goto l127
				}
				position++
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
				if !_rules[rulehtmlid]() {
					goto l127
				}
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
				if buffer[position] != rune(')') {
					goto l127
				}
				position++
				if !_rules[ruleAction7]() {
					goto l127
				}
				depth--
				add(rulestyle, position128)
			}
			return true
		l127:
			position, tokenIndex, depth = position127, tokenIndex127, depth127
			return false
		},
		/* 13 class <- <(('c' / 'C') ('l' / 'L') ('a' / 'A') ('s' / 'S') ('s' / 'S') isp* '(' isp* htmlid isp* (',' isp* htmlid isp*)* ')' Action8)> */
		func() bool {
			position145, tokenIndex145, depth145 := position, tokenIndex, depth
			{
				position146 := position
				depth++
				{
					position147, tokenIndex147, depth147 := position, tokenIndex, depth
					if buffer[position] != rune('c') {
						goto l148
					}
					position++
					goto l147
				l148:
					position, tokenIndex, depth = position147, tokenIndex147, depth147
					if buffer[position] != rune('C') {
						goto l145
					}
					position++
				}
			l147:
				{
					position149, tokenIndex149, depth149 := position, tokenIndex, depth
					if buffer[position] != rune('l') {
						goto l150
					}
					position++
					goto l149
				l150:
					position, tokenIndex, depth = position149, tokenIndex149, depth149
					if buffer[position] != rune('L') {
						goto l145
					}
					position++
				}
			l149:
				{
					position151, tokenIndex151, depth151 := position, tokenIndex, depth
					if buffer[position] != rune('a') {
						goto l152
					}
					position++
					goto l151
				l152:
					position, tokenIndex, depth = position151, tokenIndex151, depth151
					if buffer[position] != rune('A') {
						goto l145
					}
					position++
				}
			l151:
				{
					position153, tokenIndex153, depth153 := position, tokenIndex, depth
					if buffer[position] != rune('s') {
						goto l154
					}
					position++
					goto l153
				l154:
					position, tokenIndex, depth = position153, tokenIndex153, depth153
					if buffer[position] != rune('S') {
						goto l145
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
						goto l145
					}
					position++
				}
			l155:
			l157:
				{
					position158, tokenIndex158, depth158 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l158
					}
					goto l157
				l158:
					position, tokenIndex, depth = position158, tokenIndex158, depth158
				}
				if buffer[position] != rune('(') {
					goto l145
				}
				position++
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
				if !_rules[rulehtmlid]() {
					goto l145
				}
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
			l163:
				{
					position164, tokenIndex164, depth164 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l164
					}
					position++
				l165:
					{
						position166, tokenIndex166, depth166 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l166
						}
						goto l165
					l166:
						position, tokenIndex, depth = position166, tokenIndex166, depth166
					}
					if !_rules[rulehtmlid]() {
						goto l164
					}
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
					goto l163
				l164:
					position, tokenIndex, depth = position164, tokenIndex164, depth164
				}
				if buffer[position] != rune(')') {
					goto l145
				}
				position++
				if !_rules[ruleAction8]() {
					goto l145
				}
				depth--
				add(ruleclass, position146)
			}
			return true
		l145:
			position, tokenIndex, depth = position145, tokenIndex145, depth145
			return false
		},
		/* 14 form <- <(('f' / 'F') ('o' / 'O') ('r' / 'R') ('m' / 'M') isp* '(' isp* htmlid isp* ')' Action9)> */
		func() bool {
			position169, tokenIndex169, depth169 := position, tokenIndex, depth
			{
				position170 := position
				depth++
				{
					position171, tokenIndex171, depth171 := position, tokenIndex, depth
					if buffer[position] != rune('f') {
						goto l172
					}
					position++
					goto l171
				l172:
					position, tokenIndex, depth = position171, tokenIndex171, depth171
					if buffer[position] != rune('F') {
						goto l169
					}
					position++
				}
			l171:
				{
					position173, tokenIndex173, depth173 := position, tokenIndex, depth
					if buffer[position] != rune('o') {
						goto l174
					}
					position++
					goto l173
				l174:
					position, tokenIndex, depth = position173, tokenIndex173, depth173
					if buffer[position] != rune('O') {
						goto l169
					}
					position++
				}
			l173:
				{
					position175, tokenIndex175, depth175 := position, tokenIndex, depth
					if buffer[position] != rune('r') {
						goto l176
					}
					position++
					goto l175
				l176:
					position, tokenIndex, depth = position175, tokenIndex175, depth175
					if buffer[position] != rune('R') {
						goto l169
					}
					position++
				}
			l175:
				{
					position177, tokenIndex177, depth177 := position, tokenIndex, depth
					if buffer[position] != rune('m') {
						goto l178
					}
					position++
					goto l177
				l178:
					position, tokenIndex, depth = position177, tokenIndex177, depth177
					if buffer[position] != rune('M') {
						goto l169
					}
					position++
				}
			l177:
			l179:
				{
					position180, tokenIndex180, depth180 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l180
					}
					goto l179
				l180:
					position, tokenIndex, depth = position180, tokenIndex180, depth180
				}
				if buffer[position] != rune('(') {
					goto l169
				}
				position++
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
				if !_rules[rulehtmlid]() {
					goto l169
				}
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
				if buffer[position] != rune(')') {
					goto l169
				}
				position++
				if !_rules[ruleAction9]() {
					goto l169
				}
				depth--
				add(ruleform, position170)
			}
			return true
		l169:
			position, tokenIndex, depth = position169, tokenIndex169, depth169
			return false
		},
		/* 15 event <- <(('e' / 'E') ('v' / 'V') ('e' / 'E') ('n' / 'N') ('t' / 'T') isp* '(' isp* jsid? isp* ')' Action10)> */
		func() bool {
			position185, tokenIndex185, depth185 := position, tokenIndex, depth
			{
				position186 := position
				depth++
				{
					position187, tokenIndex187, depth187 := position, tokenIndex, depth
					if buffer[position] != rune('e') {
						goto l188
					}
					position++
					goto l187
				l188:
					position, tokenIndex, depth = position187, tokenIndex187, depth187
					if buffer[position] != rune('E') {
						goto l185
					}
					position++
				}
			l187:
				{
					position189, tokenIndex189, depth189 := position, tokenIndex, depth
					if buffer[position] != rune('v') {
						goto l190
					}
					position++
					goto l189
				l190:
					position, tokenIndex, depth = position189, tokenIndex189, depth189
					if buffer[position] != rune('V') {
						goto l185
					}
					position++
				}
			l189:
				{
					position191, tokenIndex191, depth191 := position, tokenIndex, depth
					if buffer[position] != rune('e') {
						goto l192
					}
					position++
					goto l191
				l192:
					position, tokenIndex, depth = position191, tokenIndex191, depth191
					if buffer[position] != rune('E') {
						goto l185
					}
					position++
				}
			l191:
				{
					position193, tokenIndex193, depth193 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l194
					}
					position++
					goto l193
				l194:
					position, tokenIndex, depth = position193, tokenIndex193, depth193
					if buffer[position] != rune('N') {
						goto l185
					}
					position++
				}
			l193:
				{
					position195, tokenIndex195, depth195 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l196
					}
					position++
					goto l195
				l196:
					position, tokenIndex, depth = position195, tokenIndex195, depth195
					if buffer[position] != rune('T') {
						goto l185
					}
					position++
				}
			l195:
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
				if buffer[position] != rune('(') {
					goto l185
				}
				position++
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
				{
					position201, tokenIndex201, depth201 := position, tokenIndex, depth
					if !_rules[rulejsid]() {
						goto l201
					}
					goto l202
				l201:
					position, tokenIndex, depth = position201, tokenIndex201, depth201
				}
			l202:
			l203:
				{
					position204, tokenIndex204, depth204 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l204
					}
					goto l203
				l204:
					position, tokenIndex, depth = position204, tokenIndex204, depth204
				}
				if buffer[position] != rune(')') {
					goto l185
				}
				position++
				if !_rules[ruleAction10]() {
					goto l185
				}
				depth--
				add(ruleevent, position186)
			}
			return true
		l185:
			position, tokenIndex, depth = position185, tokenIndex185, depth185
			return false
		},
		/* 16 htmlid <- <(<([0-9] / [a-z] / [A-Z] / '_' / '-')+> Action11)> */
		func() bool {
			position205, tokenIndex205, depth205 := position, tokenIndex, depth
			{
				position206 := position
				depth++
				{
					position207 := position
					depth++
					{
						position210, tokenIndex210, depth210 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l211
						}
						position++
						goto l210
					l211:
						position, tokenIndex, depth = position210, tokenIndex210, depth210
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l212
						}
						position++
						goto l210
					l212:
						position, tokenIndex, depth = position210, tokenIndex210, depth210
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l213
						}
						position++
						goto l210
					l213:
						position, tokenIndex, depth = position210, tokenIndex210, depth210
						if buffer[position] != rune('_') {
							goto l214
						}
						position++
						goto l210
					l214:
						position, tokenIndex, depth = position210, tokenIndex210, depth210
						if buffer[position] != rune('-') {
							goto l205
						}
						position++
					}
				l210:
				l208:
					{
						position209, tokenIndex209, depth209 := position, tokenIndex, depth
						{
							position215, tokenIndex215, depth215 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l216
							}
							position++
							goto l215
						l216:
							position, tokenIndex, depth = position215, tokenIndex215, depth215
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l217
							}
							position++
							goto l215
						l217:
							position, tokenIndex, depth = position215, tokenIndex215, depth215
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l218
							}
							position++
							goto l215
						l218:
							position, tokenIndex, depth = position215, tokenIndex215, depth215
							if buffer[position] != rune('_') {
								goto l219
							}
							position++
							goto l215
						l219:
							position, tokenIndex, depth = position215, tokenIndex215, depth215
							if buffer[position] != rune('-') {
								goto l209
							}
							position++
						}
					l215:
						goto l208
					l209:
						position, tokenIndex, depth = position209, tokenIndex209, depth209
					}
					depth--
					add(rulePegText, position207)
				}
				if !_rules[ruleAction11]() {
					goto l205
				}
				depth--
				add(rulehtmlid, position206)
			}
			return true
		l205:
			position, tokenIndex, depth = position205, tokenIndex205, depth205
			return false
		},
		/* 17 jsid <- <(<(([a-z] / [A-Z] / '_') ([0-9] / [a-z] / [A-Z] / '_')*)> Action12)> */
		func() bool {
			position220, tokenIndex220, depth220 := position, tokenIndex, depth
			{
				position221 := position
				depth++
				{
					position222 := position
					depth++
					{
						position223, tokenIndex223, depth223 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l224
						}
						position++
						goto l223
					l224:
						position, tokenIndex, depth = position223, tokenIndex223, depth223
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l225
						}
						position++
						goto l223
					l225:
						position, tokenIndex, depth = position223, tokenIndex223, depth223
						if buffer[position] != rune('_') {
							goto l220
						}
						position++
					}
				l223:
				l226:
					{
						position227, tokenIndex227, depth227 := position, tokenIndex, depth
						{
							position228, tokenIndex228, depth228 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l229
							}
							position++
							goto l228
						l229:
							position, tokenIndex, depth = position228, tokenIndex228, depth228
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l230
							}
							position++
							goto l228
						l230:
							position, tokenIndex, depth = position228, tokenIndex228, depth228
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l231
							}
							position++
							goto l228
						l231:
							position, tokenIndex, depth = position228, tokenIndex228, depth228
							if buffer[position] != rune('_') {
								goto l227
							}
							position++
						}
					l228:
						goto l226
					l227:
						position, tokenIndex, depth = position227, tokenIndex227, depth227
					}
					depth--
					add(rulePegText, position222)
				}
				if !_rules[ruleAction12]() {
					goto l220
				}
				depth--
				add(rulejsid, position221)
			}
			return true
		l220:
			position, tokenIndex, depth = position220, tokenIndex220, depth220
			return false
		},
		/* 18 expr <- <(<(commaless / enclosed / isp+)+> Action13)> */
		func() bool {
			position232, tokenIndex232, depth232 := position, tokenIndex, depth
			{
				position233 := position
				depth++
				{
					position234 := position
					depth++
					{
						position237, tokenIndex237, depth237 := position, tokenIndex, depth
						if !_rules[rulecommaless]() {
							goto l238
						}
						goto l237
					l238:
						position, tokenIndex, depth = position237, tokenIndex237, depth237
						if !_rules[ruleenclosed]() {
							goto l239
						}
						goto l237
					l239:
						position, tokenIndex, depth = position237, tokenIndex237, depth237
						if !_rules[ruleisp]() {
							goto l232
						}
					l240:
						{
							position241, tokenIndex241, depth241 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l241
							}
							goto l240
						l241:
							position, tokenIndex, depth = position241, tokenIndex241, depth241
						}
					}
				l237:
				l235:
					{
						position236, tokenIndex236, depth236 := position, tokenIndex, depth
						{
							position242, tokenIndex242, depth242 := position, tokenIndex, depth
							if !_rules[rulecommaless]() {
								goto l243
							}
							goto l242
						l243:
							position, tokenIndex, depth = position242, tokenIndex242, depth242
							if !_rules[ruleenclosed]() {
								goto l244
							}
							goto l242
						l244:
							position, tokenIndex, depth = position242, tokenIndex242, depth242
							if !_rules[ruleisp]() {
								goto l236
							}
						l245:
							{
								position246, tokenIndex246, depth246 := position, tokenIndex, depth
								if !_rules[ruleisp]() {
									goto l246
								}
								goto l245
							l246:
								position, tokenIndex, depth = position246, tokenIndex246, depth246
							}
						}
					l242:
						goto l235
					l236:
						position, tokenIndex, depth = position236, tokenIndex236, depth236
					}
					depth--
					add(rulePegText, position234)
				}
				if !_rules[ruleAction13]() {
					goto l232
				}
				depth--
				add(ruleexpr, position233)
			}
			return true
		l232:
			position, tokenIndex, depth = position232, tokenIndex232, depth232
			return false
		},
		/* 19 commaless <- <((([a-z] / [A-Z] / '_')+ '.' ([a-z] / [A-Z] / '_')+) / identifier / number / operators / string)> */
		func() bool {
			position247, tokenIndex247, depth247 := position, tokenIndex, depth
			{
				position248 := position
				depth++
				{
					position249, tokenIndex249, depth249 := position, tokenIndex, depth
					{
						position253, tokenIndex253, depth253 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l254
						}
						position++
						goto l253
					l254:
						position, tokenIndex, depth = position253, tokenIndex253, depth253
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l255
						}
						position++
						goto l253
					l255:
						position, tokenIndex, depth = position253, tokenIndex253, depth253
						if buffer[position] != rune('_') {
							goto l250
						}
						position++
					}
				l253:
				l251:
					{
						position252, tokenIndex252, depth252 := position, tokenIndex, depth
						{
							position256, tokenIndex256, depth256 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l257
							}
							position++
							goto l256
						l257:
							position, tokenIndex, depth = position256, tokenIndex256, depth256
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l258
							}
							position++
							goto l256
						l258:
							position, tokenIndex, depth = position256, tokenIndex256, depth256
							if buffer[position] != rune('_') {
								goto l252
							}
							position++
						}
					l256:
						goto l251
					l252:
						position, tokenIndex, depth = position252, tokenIndex252, depth252
					}
					if buffer[position] != rune('.') {
						goto l250
					}
					position++
					{
						position261, tokenIndex261, depth261 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l262
						}
						position++
						goto l261
					l262:
						position, tokenIndex, depth = position261, tokenIndex261, depth261
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l263
						}
						position++
						goto l261
					l263:
						position, tokenIndex, depth = position261, tokenIndex261, depth261
						if buffer[position] != rune('_') {
							goto l250
						}
						position++
					}
				l261:
				l259:
					{
						position260, tokenIndex260, depth260 := position, tokenIndex, depth
						{
							position264, tokenIndex264, depth264 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l265
							}
							position++
							goto l264
						l265:
							position, tokenIndex, depth = position264, tokenIndex264, depth264
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l266
							}
							position++
							goto l264
						l266:
							position, tokenIndex, depth = position264, tokenIndex264, depth264
							if buffer[position] != rune('_') {
								goto l260
							}
							position++
						}
					l264:
						goto l259
					l260:
						position, tokenIndex, depth = position260, tokenIndex260, depth260
					}
					goto l249
				l250:
					position, tokenIndex, depth = position249, tokenIndex249, depth249
					if !_rules[ruleidentifier]() {
						goto l267
					}
					goto l249
				l267:
					position, tokenIndex, depth = position249, tokenIndex249, depth249
					if !_rules[rulenumber]() {
						goto l268
					}
					goto l249
				l268:
					position, tokenIndex, depth = position249, tokenIndex249, depth249
					if !_rules[ruleoperators]() {
						goto l269
					}
					goto l249
				l269:
					position, tokenIndex, depth = position249, tokenIndex249, depth249
					if !_rules[rulestring]() {
						goto l247
					}
				}
			l249:
				depth--
				add(rulecommaless, position248)
			}
			return true
		l247:
			position, tokenIndex, depth = position247, tokenIndex247, depth247
			return false
		},
		/* 20 number <- <[0-9]+> */
		func() bool {
			position270, tokenIndex270, depth270 := position, tokenIndex, depth
			{
				position271 := position
				depth++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l270
				}
				position++
			l272:
				{
					position273, tokenIndex273, depth273 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l273
					}
					position++
					goto l272
				l273:
					position, tokenIndex, depth = position273, tokenIndex273, depth273
				}
				depth--
				add(rulenumber, position271)
			}
			return true
		l270:
			position, tokenIndex, depth = position270, tokenIndex270, depth270
			return false
		},
		/* 21 operators <- <('+' / '-' / '*' / '/' / '|' / '&' / '^' / ':' / '=' / '.' / '!' / '<' / '>')+> */
		func() bool {
			position274, tokenIndex274, depth274 := position, tokenIndex, depth
			{
				position275 := position
				depth++
				{
					position278, tokenIndex278, depth278 := position, tokenIndex, depth
					if buffer[position] != rune('+') {
						goto l279
					}
					position++
					goto l278
				l279:
					position, tokenIndex, depth = position278, tokenIndex278, depth278
					if buffer[position] != rune('-') {
						goto l280
					}
					position++
					goto l278
				l280:
					position, tokenIndex, depth = position278, tokenIndex278, depth278
					if buffer[position] != rune('*') {
						goto l281
					}
					position++
					goto l278
				l281:
					position, tokenIndex, depth = position278, tokenIndex278, depth278
					if buffer[position] != rune('/') {
						goto l282
					}
					position++
					goto l278
				l282:
					position, tokenIndex, depth = position278, tokenIndex278, depth278
					if buffer[position] != rune('|') {
						goto l283
					}
					position++
					goto l278
				l283:
					position, tokenIndex, depth = position278, tokenIndex278, depth278
					if buffer[position] != rune('&') {
						goto l284
					}
					position++
					goto l278
				l284:
					position, tokenIndex, depth = position278, tokenIndex278, depth278
					if buffer[position] != rune('^') {
						goto l285
					}
					position++
					goto l278
				l285:
					position, tokenIndex, depth = position278, tokenIndex278, depth278
					if buffer[position] != rune(':') {
						goto l286
					}
					position++
					goto l278
				l286:
					position, tokenIndex, depth = position278, tokenIndex278, depth278
					if buffer[position] != rune('=') {
						goto l287
					}
					position++
					goto l278
				l287:
					position, tokenIndex, depth = position278, tokenIndex278, depth278
					if buffer[position] != rune('.') {
						goto l288
					}
					position++
					goto l278
				l288:
					position, tokenIndex, depth = position278, tokenIndex278, depth278
					if buffer[position] != rune('!') {
						goto l289
					}
					position++
					goto l278
				l289:
					position, tokenIndex, depth = position278, tokenIndex278, depth278
					if buffer[position] != rune('<') {
						goto l290
					}
					position++
					goto l278
				l290:
					position, tokenIndex, depth = position278, tokenIndex278, depth278
					if buffer[position] != rune('>') {
						goto l274
					}
					position++
				}
			l278:
			l276:
				{
					position277, tokenIndex277, depth277 := position, tokenIndex, depth
					{
						position291, tokenIndex291, depth291 := position, tokenIndex, depth
						if buffer[position] != rune('+') {
							goto l292
						}
						position++
						goto l291
					l292:
						position, tokenIndex, depth = position291, tokenIndex291, depth291
						if buffer[position] != rune('-') {
							goto l293
						}
						position++
						goto l291
					l293:
						position, tokenIndex, depth = position291, tokenIndex291, depth291
						if buffer[position] != rune('*') {
							goto l294
						}
						position++
						goto l291
					l294:
						position, tokenIndex, depth = position291, tokenIndex291, depth291
						if buffer[position] != rune('/') {
							goto l295
						}
						position++
						goto l291
					l295:
						position, tokenIndex, depth = position291, tokenIndex291, depth291
						if buffer[position] != rune('|') {
							goto l296
						}
						position++
						goto l291
					l296:
						position, tokenIndex, depth = position291, tokenIndex291, depth291
						if buffer[position] != rune('&') {
							goto l297
						}
						position++
						goto l291
					l297:
						position, tokenIndex, depth = position291, tokenIndex291, depth291
						if buffer[position] != rune('^') {
							goto l298
						}
						position++
						goto l291
					l298:
						position, tokenIndex, depth = position291, tokenIndex291, depth291
						if buffer[position] != rune(':') {
							goto l299
						}
						position++
						goto l291
					l299:
						position, tokenIndex, depth = position291, tokenIndex291, depth291
						if buffer[position] != rune('=') {
							goto l300
						}
						position++
						goto l291
					l300:
						position, tokenIndex, depth = position291, tokenIndex291, depth291
						if buffer[position] != rune('.') {
							goto l301
						}
						position++
						goto l291
					l301:
						position, tokenIndex, depth = position291, tokenIndex291, depth291
						if buffer[position] != rune('!') {
							goto l302
						}
						position++
						goto l291
					l302:
						position, tokenIndex, depth = position291, tokenIndex291, depth291
						if buffer[position] != rune('<') {
							goto l303
						}
						position++
						goto l291
					l303:
						position, tokenIndex, depth = position291, tokenIndex291, depth291
						if buffer[position] != rune('>') {
							goto l277
						}
						position++
					}
				l291:
					goto l276
				l277:
					position, tokenIndex, depth = position277, tokenIndex277, depth277
				}
				depth--
				add(ruleoperators, position275)
			}
			return true
		l274:
			position, tokenIndex, depth = position274, tokenIndex274, depth274
			return false
		},
		/* 22 string <- <(('`' (!'`' .)* '`') / ('"' ((!'"' .) / ('\\' '"'))* '"'))> */
		func() bool {
			position304, tokenIndex304, depth304 := position, tokenIndex, depth
			{
				position305 := position
				depth++
				{
					position306, tokenIndex306, depth306 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l307
					}
					position++
				l308:
					{
						position309, tokenIndex309, depth309 := position, tokenIndex, depth
						{
							position310, tokenIndex310, depth310 := position, tokenIndex, depth
							if buffer[position] != rune('`') {
								goto l310
							}
							position++
							goto l309
						l310:
							position, tokenIndex, depth = position310, tokenIndex310, depth310
						}
						if !matchDot() {
							goto l309
						}
						goto l308
					l309:
						position, tokenIndex, depth = position309, tokenIndex309, depth309
					}
					if buffer[position] != rune('`') {
						goto l307
					}
					position++
					goto l306
				l307:
					position, tokenIndex, depth = position306, tokenIndex306, depth306
					if buffer[position] != rune('"') {
						goto l304
					}
					position++
				l311:
					{
						position312, tokenIndex312, depth312 := position, tokenIndex, depth
						{
							position313, tokenIndex313, depth313 := position, tokenIndex, depth
							{
								position315, tokenIndex315, depth315 := position, tokenIndex, depth
								if buffer[position] != rune('"') {
									goto l315
								}
								position++
								goto l314
							l315:
								position, tokenIndex, depth = position315, tokenIndex315, depth315
							}
							if !matchDot() {
								goto l314
							}
							goto l313
						l314:
							position, tokenIndex, depth = position313, tokenIndex313, depth313
							if buffer[position] != rune('\\') {
								goto l312
							}
							position++
							if buffer[position] != rune('"') {
								goto l312
							}
							position++
						}
					l313:
						goto l311
					l312:
						position, tokenIndex, depth = position312, tokenIndex312, depth312
					}
					if buffer[position] != rune('"') {
						goto l304
					}
					position++
				}
			l306:
				depth--
				add(rulestring, position305)
			}
			return true
		l304:
			position, tokenIndex, depth = position304, tokenIndex304, depth304
			return false
		},
		/* 23 enclosed <- <(parens / braces / brackets)> */
		func() bool {
			position316, tokenIndex316, depth316 := position, tokenIndex, depth
			{
				position317 := position
				depth++
				{
					position318, tokenIndex318, depth318 := position, tokenIndex, depth
					if !_rules[ruleparens]() {
						goto l319
					}
					goto l318
				l319:
					position, tokenIndex, depth = position318, tokenIndex318, depth318
					if !_rules[rulebraces]() {
						goto l320
					}
					goto l318
				l320:
					position, tokenIndex, depth = position318, tokenIndex318, depth318
					if !_rules[rulebrackets]() {
						goto l316
					}
				}
			l318:
				depth--
				add(ruleenclosed, position317)
			}
			return true
		l316:
			position, tokenIndex, depth = position316, tokenIndex316, depth316
			return false
		},
		/* 24 parens <- <('(' inner ')')> */
		func() bool {
			position321, tokenIndex321, depth321 := position, tokenIndex, depth
			{
				position322 := position
				depth++
				if buffer[position] != rune('(') {
					goto l321
				}
				position++
				if !_rules[ruleinner]() {
					goto l321
				}
				if buffer[position] != rune(')') {
					goto l321
				}
				position++
				depth--
				add(ruleparens, position322)
			}
			return true
		l321:
			position, tokenIndex, depth = position321, tokenIndex321, depth321
			return false
		},
		/* 25 braces <- <('{' inner '}')> */
		func() bool {
			position323, tokenIndex323, depth323 := position, tokenIndex, depth
			{
				position324 := position
				depth++
				if buffer[position] != rune('{') {
					goto l323
				}
				position++
				if !_rules[ruleinner]() {
					goto l323
				}
				if buffer[position] != rune('}') {
					goto l323
				}
				position++
				depth--
				add(rulebraces, position324)
			}
			return true
		l323:
			position, tokenIndex, depth = position323, tokenIndex323, depth323
			return false
		},
		/* 26 brackets <- <('[' inner ']')> */
		func() bool {
			position325, tokenIndex325, depth325 := position, tokenIndex, depth
			{
				position326 := position
				depth++
				if buffer[position] != rune('[') {
					goto l325
				}
				position++
				if !_rules[ruleinner]() {
					goto l325
				}
				if buffer[position] != rune(']') {
					goto l325
				}
				position++
				depth--
				add(rulebrackets, position326)
			}
			return true
		l325:
			position, tokenIndex, depth = position325, tokenIndex325, depth325
			return false
		},
		/* 27 inner <- <(commaless / enclosed / ',' / isp+)*> */
		func() bool {
			{
				position328 := position
				depth++
			l329:
				{
					position330, tokenIndex330, depth330 := position, tokenIndex, depth
					{
						position331, tokenIndex331, depth331 := position, tokenIndex, depth
						if !_rules[rulecommaless]() {
							goto l332
						}
						goto l331
					l332:
						position, tokenIndex, depth = position331, tokenIndex331, depth331
						if !_rules[ruleenclosed]() {
							goto l333
						}
						goto l331
					l333:
						position, tokenIndex, depth = position331, tokenIndex331, depth331
						if buffer[position] != rune(',') {
							goto l334
						}
						position++
						goto l331
					l334:
						position, tokenIndex, depth = position331, tokenIndex331, depth331
						if !_rules[ruleisp]() {
							goto l330
						}
					l335:
						{
							position336, tokenIndex336, depth336 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l336
							}
							goto l335
						l336:
							position, tokenIndex, depth = position336, tokenIndex336, depth336
						}
					}
				l331:
					goto l329
				l330:
					position, tokenIndex, depth = position330, tokenIndex330, depth330
				}
				depth--
				add(ruleinner, position328)
			}
			return true
		},
		/* 28 identifier <- <(([a-z] / [A-Z] / '_') ([a-z] / [A-Z] / '_' / ([0-9] / [0-9]))*)> */
		func() bool {
			position337, tokenIndex337, depth337 := position, tokenIndex, depth
			{
				position338 := position
				depth++
				{
					position339, tokenIndex339, depth339 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l340
					}
					position++
					goto l339
				l340:
					position, tokenIndex, depth = position339, tokenIndex339, depth339
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l341
					}
					position++
					goto l339
				l341:
					position, tokenIndex, depth = position339, tokenIndex339, depth339
					if buffer[position] != rune('_') {
						goto l337
					}
					position++
				}
			l339:
			l342:
				{
					position343, tokenIndex343, depth343 := position, tokenIndex, depth
					{
						position344, tokenIndex344, depth344 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l345
						}
						position++
						goto l344
					l345:
						position, tokenIndex, depth = position344, tokenIndex344, depth344
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l346
						}
						position++
						goto l344
					l346:
						position, tokenIndex, depth = position344, tokenIndex344, depth344
						if buffer[position] != rune('_') {
							goto l347
						}
						position++
						goto l344
					l347:
						position, tokenIndex, depth = position344, tokenIndex344, depth344
						{
							position348, tokenIndex348, depth348 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l349
							}
							position++
							goto l348
						l349:
							position, tokenIndex, depth = position348, tokenIndex348, depth348
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l343
							}
							position++
						}
					l348:
					}
				l344:
					goto l342
				l343:
					position, tokenIndex, depth = position343, tokenIndex343, depth343
				}
				depth--
				add(ruleidentifier, position338)
			}
			return true
		l337:
			position, tokenIndex, depth = position337, tokenIndex337, depth337
			return false
		},
		/* 29 fields <- <((';' / ' ' / '\t' / '\n')* field isp* (fsep isp* (fsep isp*)* field)* (';' / ' ' / '\t' / '\n')* !.)> */
		func() bool {
			position350, tokenIndex350, depth350 := position, tokenIndex, depth
			{
				position351 := position
				depth++
			l352:
				{
					position353, tokenIndex353, depth353 := position, tokenIndex, depth
					{
						position354, tokenIndex354, depth354 := position, tokenIndex, depth
						if buffer[position] != rune(';') {
							goto l355
						}
						position++
						goto l354
					l355:
						position, tokenIndex, depth = position354, tokenIndex354, depth354
						if buffer[position] != rune(' ') {
							goto l356
						}
						position++
						goto l354
					l356:
						position, tokenIndex, depth = position354, tokenIndex354, depth354
						if buffer[position] != rune('\t') {
							goto l357
						}
						position++
						goto l354
					l357:
						position, tokenIndex, depth = position354, tokenIndex354, depth354
						if buffer[position] != rune('\n') {
							goto l353
						}
						position++
					}
				l354:
					goto l352
				l353:
					position, tokenIndex, depth = position353, tokenIndex353, depth353
				}
				if !_rules[rulefield]() {
					goto l350
				}
			l358:
				{
					position359, tokenIndex359, depth359 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l359
					}
					goto l358
				l359:
					position, tokenIndex, depth = position359, tokenIndex359, depth359
				}
			l360:
				{
					position361, tokenIndex361, depth361 := position, tokenIndex, depth
					if !_rules[rulefsep]() {
						goto l361
					}
				l362:
					{
						position363, tokenIndex363, depth363 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l363
						}
						goto l362
					l363:
						position, tokenIndex, depth = position363, tokenIndex363, depth363
					}
				l364:
					{
						position365, tokenIndex365, depth365 := position, tokenIndex, depth
						if !_rules[rulefsep]() {
							goto l365
						}
					l366:
						{
							position367, tokenIndex367, depth367 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l367
							}
							goto l366
						l367:
							position, tokenIndex, depth = position367, tokenIndex367, depth367
						}
						goto l364
					l365:
						position, tokenIndex, depth = position365, tokenIndex365, depth365
					}
					if !_rules[rulefield]() {
						goto l361
					}
					goto l360
				l361:
					position, tokenIndex, depth = position361, tokenIndex361, depth361
				}
			l368:
				{
					position369, tokenIndex369, depth369 := position, tokenIndex, depth
					{
						position370, tokenIndex370, depth370 := position, tokenIndex, depth
						if buffer[position] != rune(';') {
							goto l371
						}
						position++
						goto l370
					l371:
						position, tokenIndex, depth = position370, tokenIndex370, depth370
						if buffer[position] != rune(' ') {
							goto l372
						}
						position++
						goto l370
					l372:
						position, tokenIndex, depth = position370, tokenIndex370, depth370
						if buffer[position] != rune('\t') {
							goto l373
						}
						position++
						goto l370
					l373:
						position, tokenIndex, depth = position370, tokenIndex370, depth370
						if buffer[position] != rune('\n') {
							goto l369
						}
						position++
					}
				l370:
					goto l368
				l369:
					position, tokenIndex, depth = position369, tokenIndex369, depth369
				}
				{
					position374, tokenIndex374, depth374 := position, tokenIndex, depth
					if !matchDot() {
						goto l374
					}
					goto l350
				l374:
					position, tokenIndex, depth = position374, tokenIndex374, depth374
				}
				depth--
				add(rulefields, position351)
			}
			return true
		l350:
			position, tokenIndex, depth = position350, tokenIndex350, depth350
			return false
		},
		/* 30 fsep <- <(';' / '\n')> */
		func() bool {
			position375, tokenIndex375, depth375 := position, tokenIndex, depth
			{
				position376 := position
				depth++
				{
					position377, tokenIndex377, depth377 := position, tokenIndex, depth
					if buffer[position] != rune(';') {
						goto l378
					}
					position++
					goto l377
				l378:
					position, tokenIndex, depth = position377, tokenIndex377, depth377
					if buffer[position] != rune('\n') {
						goto l375
					}
					position++
				}
			l377:
				depth--
				add(rulefsep, position376)
			}
			return true
		l375:
			position, tokenIndex, depth = position375, tokenIndex375, depth375
			return false
		},
		/* 31 field <- <(name (isp* ',' isp* name)* isp+ type isp* ('=' isp* expr)? Action14)> */
		func() bool {
			position379, tokenIndex379, depth379 := position, tokenIndex, depth
			{
				position380 := position
				depth++
				if !_rules[rulename]() {
					goto l379
				}
			l381:
				{
					position382, tokenIndex382, depth382 := position, tokenIndex, depth
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
					if buffer[position] != rune(',') {
						goto l382
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
					if !_rules[rulename]() {
						goto l382
					}
					goto l381
				l382:
					position, tokenIndex, depth = position382, tokenIndex382, depth382
				}
				if !_rules[ruleisp]() {
					goto l379
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
				if !_rules[ruletype]() {
					goto l379
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
				{
					position391, tokenIndex391, depth391 := position, tokenIndex, depth
					if buffer[position] != rune('=') {
						goto l391
					}
					position++
				l393:
					{
						position394, tokenIndex394, depth394 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l394
						}
						goto l393
					l394:
						position, tokenIndex, depth = position394, tokenIndex394, depth394
					}
					if !_rules[ruleexpr]() {
						goto l391
					}
					goto l392
				l391:
					position, tokenIndex, depth = position391, tokenIndex391, depth391
				}
			l392:
				if !_rules[ruleAction14]() {
					goto l379
				}
				depth--
				add(rulefield, position380)
			}
			return true
		l379:
			position, tokenIndex, depth = position379, tokenIndex379, depth379
			return false
		},
		/* 32 name <- <(<([a-z] / [A-Z] / '_')+> Action15)> */
		func() bool {
			position395, tokenIndex395, depth395 := position, tokenIndex, depth
			{
				position396 := position
				depth++
				{
					position397 := position
					depth++
					{
						position400, tokenIndex400, depth400 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l401
						}
						position++
						goto l400
					l401:
						position, tokenIndex, depth = position400, tokenIndex400, depth400
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l402
						}
						position++
						goto l400
					l402:
						position, tokenIndex, depth = position400, tokenIndex400, depth400
						if buffer[position] != rune('_') {
							goto l395
						}
						position++
					}
				l400:
				l398:
					{
						position399, tokenIndex399, depth399 := position, tokenIndex, depth
						{
							position403, tokenIndex403, depth403 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l404
							}
							position++
							goto l403
						l404:
							position, tokenIndex, depth = position403, tokenIndex403, depth403
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l405
							}
							position++
							goto l403
						l405:
							position, tokenIndex, depth = position403, tokenIndex403, depth403
							if buffer[position] != rune('_') {
								goto l399
							}
							position++
						}
					l403:
						goto l398
					l399:
						position, tokenIndex, depth = position399, tokenIndex399, depth399
					}
					depth--
					add(rulePegText, position397)
				}
				if !_rules[ruleAction15]() {
					goto l395
				}
				depth--
				add(rulename, position396)
			}
			return true
		l395:
			position, tokenIndex, depth = position395, tokenIndex395, depth395
			return false
		},
		/* 33 type <- <(qname / sname / array / map / pointer)> */
		func() bool {
			position406, tokenIndex406, depth406 := position, tokenIndex, depth
			{
				position407 := position
				depth++
				{
					position408, tokenIndex408, depth408 := position, tokenIndex, depth
					if !_rules[ruleqname]() {
						goto l409
					}
					goto l408
				l409:
					position, tokenIndex, depth = position408, tokenIndex408, depth408
					if !_rules[rulesname]() {
						goto l410
					}
					goto l408
				l410:
					position, tokenIndex, depth = position408, tokenIndex408, depth408
					if !_rules[rulearray]() {
						goto l411
					}
					goto l408
				l411:
					position, tokenIndex, depth = position408, tokenIndex408, depth408
					if !_rules[rulemap]() {
						goto l412
					}
					goto l408
				l412:
					position, tokenIndex, depth = position408, tokenIndex408, depth408
					if !_rules[rulepointer]() {
						goto l406
					}
				}
			l408:
				depth--
				add(ruletype, position407)
			}
			return true
		l406:
			position, tokenIndex, depth = position406, tokenIndex406, depth406
			return false
		},
		/* 34 sname <- <(<([a-z] / [A-Z] / '_')+> Action16)> */
		func() bool {
			position413, tokenIndex413, depth413 := position, tokenIndex, depth
			{
				position414 := position
				depth++
				{
					position415 := position
					depth++
					{
						position418, tokenIndex418, depth418 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l419
						}
						position++
						goto l418
					l419:
						position, tokenIndex, depth = position418, tokenIndex418, depth418
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l420
						}
						position++
						goto l418
					l420:
						position, tokenIndex, depth = position418, tokenIndex418, depth418
						if buffer[position] != rune('_') {
							goto l413
						}
						position++
					}
				l418:
				l416:
					{
						position417, tokenIndex417, depth417 := position, tokenIndex, depth
						{
							position421, tokenIndex421, depth421 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l422
							}
							position++
							goto l421
						l422:
							position, tokenIndex, depth = position421, tokenIndex421, depth421
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l423
							}
							position++
							goto l421
						l423:
							position, tokenIndex, depth = position421, tokenIndex421, depth421
							if buffer[position] != rune('_') {
								goto l417
							}
							position++
						}
					l421:
						goto l416
					l417:
						position, tokenIndex, depth = position417, tokenIndex417, depth417
					}
					depth--
					add(rulePegText, position415)
				}
				if !_rules[ruleAction16]() {
					goto l413
				}
				depth--
				add(rulesname, position414)
			}
			return true
		l413:
			position, tokenIndex, depth = position413, tokenIndex413, depth413
			return false
		},
		/* 35 qname <- <(<(([a-z] / [A-Z] / '_')+ '.' ([a-z] / [A-Z] / '_')+)> Action17)> */
		func() bool {
			position424, tokenIndex424, depth424 := position, tokenIndex, depth
			{
				position425 := position
				depth++
				{
					position426 := position
					depth++
					{
						position429, tokenIndex429, depth429 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l430
						}
						position++
						goto l429
					l430:
						position, tokenIndex, depth = position429, tokenIndex429, depth429
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l431
						}
						position++
						goto l429
					l431:
						position, tokenIndex, depth = position429, tokenIndex429, depth429
						if buffer[position] != rune('_') {
							goto l424
						}
						position++
					}
				l429:
				l427:
					{
						position428, tokenIndex428, depth428 := position, tokenIndex, depth
						{
							position432, tokenIndex432, depth432 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l433
							}
							position++
							goto l432
						l433:
							position, tokenIndex, depth = position432, tokenIndex432, depth432
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l434
							}
							position++
							goto l432
						l434:
							position, tokenIndex, depth = position432, tokenIndex432, depth432
							if buffer[position] != rune('_') {
								goto l428
							}
							position++
						}
					l432:
						goto l427
					l428:
						position, tokenIndex, depth = position428, tokenIndex428, depth428
					}
					if buffer[position] != rune('.') {
						goto l424
					}
					position++
					{
						position437, tokenIndex437, depth437 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l438
						}
						position++
						goto l437
					l438:
						position, tokenIndex, depth = position437, tokenIndex437, depth437
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l439
						}
						position++
						goto l437
					l439:
						position, tokenIndex, depth = position437, tokenIndex437, depth437
						if buffer[position] != rune('_') {
							goto l424
						}
						position++
					}
				l437:
				l435:
					{
						position436, tokenIndex436, depth436 := position, tokenIndex, depth
						{
							position440, tokenIndex440, depth440 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l441
							}
							position++
							goto l440
						l441:
							position, tokenIndex, depth = position440, tokenIndex440, depth440
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l442
							}
							position++
							goto l440
						l442:
							position, tokenIndex, depth = position440, tokenIndex440, depth440
							if buffer[position] != rune('_') {
								goto l436
							}
							position++
						}
					l440:
						goto l435
					l436:
						position, tokenIndex, depth = position436, tokenIndex436, depth436
					}
					depth--
					add(rulePegText, position426)
				}
				if !_rules[ruleAction17]() {
					goto l424
				}
				depth--
				add(ruleqname, position425)
			}
			return true
		l424:
			position, tokenIndex, depth = position424, tokenIndex424, depth424
			return false
		},
		/* 36 array <- <('[' ']' type Action18)> */
		func() bool {
			position443, tokenIndex443, depth443 := position, tokenIndex, depth
			{
				position444 := position
				depth++
				if buffer[position] != rune('[') {
					goto l443
				}
				position++
				if buffer[position] != rune(']') {
					goto l443
				}
				position++
				if !_rules[ruletype]() {
					goto l443
				}
				if !_rules[ruleAction18]() {
					goto l443
				}
				depth--
				add(rulearray, position444)
			}
			return true
		l443:
			position, tokenIndex, depth = position443, tokenIndex443, depth443
			return false
		},
		/* 37 map <- <(('m' / 'M') ('a' / 'A') ('p' / 'P') '[' isp* keytype isp* ']' type Action19)> */
		func() bool {
			position445, tokenIndex445, depth445 := position, tokenIndex, depth
			{
				position446 := position
				depth++
				{
					position447, tokenIndex447, depth447 := position, tokenIndex, depth
					if buffer[position] != rune('m') {
						goto l448
					}
					position++
					goto l447
				l448:
					position, tokenIndex, depth = position447, tokenIndex447, depth447
					if buffer[position] != rune('M') {
						goto l445
					}
					position++
				}
			l447:
				{
					position449, tokenIndex449, depth449 := position, tokenIndex, depth
					if buffer[position] != rune('a') {
						goto l450
					}
					position++
					goto l449
				l450:
					position, tokenIndex, depth = position449, tokenIndex449, depth449
					if buffer[position] != rune('A') {
						goto l445
					}
					position++
				}
			l449:
				{
					position451, tokenIndex451, depth451 := position, tokenIndex, depth
					if buffer[position] != rune('p') {
						goto l452
					}
					position++
					goto l451
				l452:
					position, tokenIndex, depth = position451, tokenIndex451, depth451
					if buffer[position] != rune('P') {
						goto l445
					}
					position++
				}
			l451:
				if buffer[position] != rune('[') {
					goto l445
				}
				position++
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
				if !_rules[rulekeytype]() {
					goto l445
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
				if buffer[position] != rune(']') {
					goto l445
				}
				position++
				if !_rules[ruletype]() {
					goto l445
				}
				if !_rules[ruleAction19]() {
					goto l445
				}
				depth--
				add(rulemap, position446)
			}
			return true
		l445:
			position, tokenIndex, depth = position445, tokenIndex445, depth445
			return false
		},
		/* 38 keytype <- <(type Action20)> */
		func() bool {
			position457, tokenIndex457, depth457 := position, tokenIndex, depth
			{
				position458 := position
				depth++
				if !_rules[ruletype]() {
					goto l457
				}
				if !_rules[ruleAction20]() {
					goto l457
				}
				depth--
				add(rulekeytype, position458)
			}
			return true
		l457:
			position, tokenIndex, depth = position457, tokenIndex457, depth457
			return false
		},
		/* 39 pointer <- <('*' type Action21)> */
		func() bool {
			position459, tokenIndex459, depth459 := position, tokenIndex, depth
			{
				position460 := position
				depth++
				if buffer[position] != rune('*') {
					goto l459
				}
				position++
				if !_rules[ruletype]() {
					goto l459
				}
				if !_rules[ruleAction21]() {
					goto l459
				}
				depth--
				add(rulepointer, position460)
			}
			return true
		l459:
			position, tokenIndex, depth = position459, tokenIndex459, depth459
			return false
		},
		/* 40 captures <- <(isp* capture isp* (',' isp* capture isp*)* !.)> */
		func() bool {
			position461, tokenIndex461, depth461 := position, tokenIndex, depth
			{
				position462 := position
				depth++
			l463:
				{
					position464, tokenIndex464, depth464 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l464
					}
					goto l463
				l464:
					position, tokenIndex, depth = position464, tokenIndex464, depth464
				}
				if !_rules[rulecapture]() {
					goto l461
				}
			l465:
				{
					position466, tokenIndex466, depth466 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l466
					}
					goto l465
				l466:
					position, tokenIndex, depth = position466, tokenIndex466, depth466
				}
			l467:
				{
					position468, tokenIndex468, depth468 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l468
					}
					position++
				l469:
					{
						position470, tokenIndex470, depth470 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l470
						}
						goto l469
					l470:
						position, tokenIndex, depth = position470, tokenIndex470, depth470
					}
					if !_rules[rulecapture]() {
						goto l468
					}
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
					goto l467
				l468:
					position, tokenIndex, depth = position468, tokenIndex468, depth468
				}
				{
					position473, tokenIndex473, depth473 := position, tokenIndex, depth
					if !matchDot() {
						goto l473
					}
					goto l461
				l473:
					position, tokenIndex, depth = position473, tokenIndex473, depth473
				}
				depth--
				add(rulecaptures, position462)
			}
			return true
		l461:
			position, tokenIndex, depth = position461, tokenIndex461, depth461
			return false
		},
		/* 41 capture <- <(eventid isp* ':' handlername isp* mappings isp* tags Action22)> */
		func() bool {
			position474, tokenIndex474, depth474 := position, tokenIndex, depth
			{
				position475 := position
				depth++
				if !_rules[ruleeventid]() {
					goto l474
				}
			l476:
				{
					position477, tokenIndex477, depth477 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l477
					}
					goto l476
				l477:
					position, tokenIndex, depth = position477, tokenIndex477, depth477
				}
				if buffer[position] != rune(':') {
					goto l474
				}
				position++
				if !_rules[rulehandlername]() {
					goto l474
				}
			l478:
				{
					position479, tokenIndex479, depth479 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l479
					}
					goto l478
				l479:
					position, tokenIndex, depth = position479, tokenIndex479, depth479
				}
				if !_rules[rulemappings]() {
					goto l474
				}
			l480:
				{
					position481, tokenIndex481, depth481 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l481
					}
					goto l480
				l481:
					position, tokenIndex, depth = position481, tokenIndex481, depth481
				}
				if !_rules[ruletags]() {
					goto l474
				}
				if !_rules[ruleAction22]() {
					goto l474
				}
				depth--
				add(rulecapture, position475)
			}
			return true
		l474:
			position, tokenIndex, depth = position474, tokenIndex474, depth474
			return false
		},
		/* 42 handlername <- <(<identifier> Action23)> */
		func() bool {
			position482, tokenIndex482, depth482 := position, tokenIndex, depth
			{
				position483 := position
				depth++
				{
					position484 := position
					depth++
					if !_rules[ruleidentifier]() {
						goto l482
					}
					depth--
					add(rulePegText, position484)
				}
				if !_rules[ruleAction23]() {
					goto l482
				}
				depth--
				add(rulehandlername, position483)
			}
			return true
		l482:
			position, tokenIndex, depth = position482, tokenIndex482, depth482
			return false
		},
		/* 43 eventid <- <(<[a-z]+> Action24)> */
		func() bool {
			position485, tokenIndex485, depth485 := position, tokenIndex, depth
			{
				position486 := position
				depth++
				{
					position487 := position
					depth++
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l485
					}
					position++
				l488:
					{
						position489, tokenIndex489, depth489 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l489
						}
						position++
						goto l488
					l489:
						position, tokenIndex, depth = position489, tokenIndex489, depth489
					}
					depth--
					add(rulePegText, position487)
				}
				if !_rules[ruleAction24]() {
					goto l485
				}
				depth--
				add(ruleeventid, position486)
			}
			return true
		l485:
			position, tokenIndex, depth = position485, tokenIndex485, depth485
			return false
		},
		/* 44 mappings <- <('(' (isp* mapping isp* (',' isp* mapping isp*)*)? ')')?> */
		func() bool {
			{
				position491 := position
				depth++
				{
					position492, tokenIndex492, depth492 := position, tokenIndex, depth
					if buffer[position] != rune('(') {
						goto l492
					}
					position++
					{
						position494, tokenIndex494, depth494 := position, tokenIndex, depth
					l496:
						{
							position497, tokenIndex497, depth497 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l497
							}
							goto l496
						l497:
							position, tokenIndex, depth = position497, tokenIndex497, depth497
						}
						if !_rules[rulemapping]() {
							goto l494
						}
					l498:
						{
							position499, tokenIndex499, depth499 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l499
							}
							goto l498
						l499:
							position, tokenIndex, depth = position499, tokenIndex499, depth499
						}
					l500:
						{
							position501, tokenIndex501, depth501 := position, tokenIndex, depth
							if buffer[position] != rune(',') {
								goto l501
							}
							position++
						l502:
							{
								position503, tokenIndex503, depth503 := position, tokenIndex, depth
								if !_rules[ruleisp]() {
									goto l503
								}
								goto l502
							l503:
								position, tokenIndex, depth = position503, tokenIndex503, depth503
							}
							if !_rules[rulemapping]() {
								goto l501
							}
						l504:
							{
								position505, tokenIndex505, depth505 := position, tokenIndex, depth
								if !_rules[ruleisp]() {
									goto l505
								}
								goto l504
							l505:
								position, tokenIndex, depth = position505, tokenIndex505, depth505
							}
							goto l500
						l501:
							position, tokenIndex, depth = position501, tokenIndex501, depth501
						}
						goto l495
					l494:
						position, tokenIndex, depth = position494, tokenIndex494, depth494
					}
				l495:
					if buffer[position] != rune(')') {
						goto l492
					}
					position++
					goto l493
				l492:
					position, tokenIndex, depth = position492, tokenIndex492, depth492
				}
			l493:
				depth--
				add(rulemappings, position491)
			}
			return true
		},
		/* 45 mapping <- <(mappingname isp* '=' isp* bound Action25)> */
		func() bool {
			position506, tokenIndex506, depth506 := position, tokenIndex, depth
			{
				position507 := position
				depth++
				if !_rules[rulemappingname]() {
					goto l506
				}
			l508:
				{
					position509, tokenIndex509, depth509 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l509
					}
					goto l508
				l509:
					position, tokenIndex, depth = position509, tokenIndex509, depth509
				}
				if buffer[position] != rune('=') {
					goto l506
				}
				position++
			l510:
				{
					position511, tokenIndex511, depth511 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l511
					}
					goto l510
				l511:
					position, tokenIndex, depth = position511, tokenIndex511, depth511
				}
				if !_rules[rulebound]() {
					goto l506
				}
				if !_rules[ruleAction25]() {
					goto l506
				}
				depth--
				add(rulemapping, position507)
			}
			return true
		l506:
			position, tokenIndex, depth = position506, tokenIndex506, depth506
			return false
		},
		/* 46 mappingname <- <(<identifier> Action26)> */
		func() bool {
			position512, tokenIndex512, depth512 := position, tokenIndex, depth
			{
				position513 := position
				depth++
				{
					position514 := position
					depth++
					if !_rules[ruleidentifier]() {
						goto l512
					}
					depth--
					add(rulePegText, position514)
				}
				if !_rules[ruleAction26]() {
					goto l512
				}
				depth--
				add(rulemappingname, position513)
			}
			return true
		l512:
			position, tokenIndex, depth = position512, tokenIndex512, depth512
			return false
		},
		/* 47 tags <- <('{' isp* tag isp* (',' isp* tag isp*)* '}')?> */
		func() bool {
			{
				position516 := position
				depth++
				{
					position517, tokenIndex517, depth517 := position, tokenIndex, depth
					if buffer[position] != rune('{') {
						goto l517
					}
					position++
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
					if !_rules[ruletag]() {
						goto l517
					}
				l521:
					{
						position522, tokenIndex522, depth522 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l522
						}
						goto l521
					l522:
						position, tokenIndex, depth = position522, tokenIndex522, depth522
					}
				l523:
					{
						position524, tokenIndex524, depth524 := position, tokenIndex, depth
						if buffer[position] != rune(',') {
							goto l524
						}
						position++
					l525:
						{
							position526, tokenIndex526, depth526 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l526
							}
							goto l525
						l526:
							position, tokenIndex, depth = position526, tokenIndex526, depth526
						}
						if !_rules[ruletag]() {
							goto l524
						}
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
						goto l523
					l524:
						position, tokenIndex, depth = position524, tokenIndex524, depth524
					}
					if buffer[position] != rune('}') {
						goto l517
					}
					position++
					goto l518
				l517:
					position, tokenIndex, depth = position517, tokenIndex517, depth517
				}
			l518:
				depth--
				add(ruletags, position516)
			}
			return true
		},
		/* 48 tag <- <(tagname ('(' (isp* tagarg isp* (',' isp* tagarg isp*)*)? ')')? Action27)> */
		func() bool {
			position529, tokenIndex529, depth529 := position, tokenIndex, depth
			{
				position530 := position
				depth++
				if !_rules[ruletagname]() {
					goto l529
				}
				{
					position531, tokenIndex531, depth531 := position, tokenIndex, depth
					if buffer[position] != rune('(') {
						goto l531
					}
					position++
					{
						position533, tokenIndex533, depth533 := position, tokenIndex, depth
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
						if !_rules[ruletagarg]() {
							goto l533
						}
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
					l539:
						{
							position540, tokenIndex540, depth540 := position, tokenIndex, depth
							if buffer[position] != rune(',') {
								goto l540
							}
							position++
						l541:
							{
								position542, tokenIndex542, depth542 := position, tokenIndex, depth
								if !_rules[ruleisp]() {
									goto l542
								}
								goto l541
							l542:
								position, tokenIndex, depth = position542, tokenIndex542, depth542
							}
							if !_rules[ruletagarg]() {
								goto l540
							}
						l543:
							{
								position544, tokenIndex544, depth544 := position, tokenIndex, depth
								if !_rules[ruleisp]() {
									goto l544
								}
								goto l543
							l544:
								position, tokenIndex, depth = position544, tokenIndex544, depth544
							}
							goto l539
						l540:
							position, tokenIndex, depth = position540, tokenIndex540, depth540
						}
						goto l534
					l533:
						position, tokenIndex, depth = position533, tokenIndex533, depth533
					}
				l534:
					if buffer[position] != rune(')') {
						goto l531
					}
					position++
					goto l532
				l531:
					position, tokenIndex, depth = position531, tokenIndex531, depth531
				}
			l532:
				if !_rules[ruleAction27]() {
					goto l529
				}
				depth--
				add(ruletag, position530)
			}
			return true
		l529:
			position, tokenIndex, depth = position529, tokenIndex529, depth529
			return false
		},
		/* 49 tagname <- <(<identifier> Action28)> */
		func() bool {
			position545, tokenIndex545, depth545 := position, tokenIndex, depth
			{
				position546 := position
				depth++
				{
					position547 := position
					depth++
					if !_rules[ruleidentifier]() {
						goto l545
					}
					depth--
					add(rulePegText, position547)
				}
				if !_rules[ruleAction28]() {
					goto l545
				}
				depth--
				add(ruletagname, position546)
			}
			return true
		l545:
			position, tokenIndex, depth = position545, tokenIndex545, depth545
			return false
		},
		/* 50 tagarg <- <(<identifier> Action29)> */
		func() bool {
			position548, tokenIndex548, depth548 := position, tokenIndex, depth
			{
				position549 := position
				depth++
				{
					position550 := position
					depth++
					if !_rules[ruleidentifier]() {
						goto l548
					}
					depth--
					add(rulePegText, position550)
				}
				if !_rules[ruleAction29]() {
					goto l548
				}
				depth--
				add(ruletagarg, position549)
			}
			return true
		l548:
			position, tokenIndex, depth = position548, tokenIndex548, depth548
			return false
		},
		/* 51 for <- <(isp* forVar isp* (',' isp* forVar isp*)? (':' '=') isp* (('r' / 'R') ('a' / 'A') ('n' / 'N') ('g' / 'G') ('e' / 'E')) isp+ expr isp* !.)> */
		func() bool {
			position551, tokenIndex551, depth551 := position, tokenIndex, depth
			{
				position552 := position
				depth++
			l553:
				{
					position554, tokenIndex554, depth554 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l554
					}
					goto l553
				l554:
					position, tokenIndex, depth = position554, tokenIndex554, depth554
				}
				if !_rules[ruleforVar]() {
					goto l551
				}
			l555:
				{
					position556, tokenIndex556, depth556 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l556
					}
					goto l555
				l556:
					position, tokenIndex, depth = position556, tokenIndex556, depth556
				}
				{
					position557, tokenIndex557, depth557 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l557
					}
					position++
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
					if !_rules[ruleforVar]() {
						goto l557
					}
				l561:
					{
						position562, tokenIndex562, depth562 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l562
						}
						goto l561
					l562:
						position, tokenIndex, depth = position562, tokenIndex562, depth562
					}
					goto l558
				l557:
					position, tokenIndex, depth = position557, tokenIndex557, depth557
				}
			l558:
				if buffer[position] != rune(':') {
					goto l551
				}
				position++
				if buffer[position] != rune('=') {
					goto l551
				}
				position++
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
				{
					position565, tokenIndex565, depth565 := position, tokenIndex, depth
					if buffer[position] != rune('r') {
						goto l566
					}
					position++
					goto l565
				l566:
					position, tokenIndex, depth = position565, tokenIndex565, depth565
					if buffer[position] != rune('R') {
						goto l551
					}
					position++
				}
			l565:
				{
					position567, tokenIndex567, depth567 := position, tokenIndex, depth
					if buffer[position] != rune('a') {
						goto l568
					}
					position++
					goto l567
				l568:
					position, tokenIndex, depth = position567, tokenIndex567, depth567
					if buffer[position] != rune('A') {
						goto l551
					}
					position++
				}
			l567:
				{
					position569, tokenIndex569, depth569 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l570
					}
					position++
					goto l569
				l570:
					position, tokenIndex, depth = position569, tokenIndex569, depth569
					if buffer[position] != rune('N') {
						goto l551
					}
					position++
				}
			l569:
				{
					position571, tokenIndex571, depth571 := position, tokenIndex, depth
					if buffer[position] != rune('g') {
						goto l572
					}
					position++
					goto l571
				l572:
					position, tokenIndex, depth = position571, tokenIndex571, depth571
					if buffer[position] != rune('G') {
						goto l551
					}
					position++
				}
			l571:
				{
					position573, tokenIndex573, depth573 := position, tokenIndex, depth
					if buffer[position] != rune('e') {
						goto l574
					}
					position++
					goto l573
				l574:
					position, tokenIndex, depth = position573, tokenIndex573, depth573
					if buffer[position] != rune('E') {
						goto l551
					}
					position++
				}
			l573:
				if !_rules[ruleisp]() {
					goto l551
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
				if !_rules[ruleexpr]() {
					goto l551
				}
			l577:
				{
					position578, tokenIndex578, depth578 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l578
					}
					goto l577
				l578:
					position, tokenIndex, depth = position578, tokenIndex578, depth578
				}
				{
					position579, tokenIndex579, depth579 := position, tokenIndex, depth
					if !matchDot() {
						goto l579
					}
					goto l551
				l579:
					position, tokenIndex, depth = position579, tokenIndex579, depth579
				}
				depth--
				add(rulefor, position552)
			}
			return true
		l551:
			position, tokenIndex, depth = position551, tokenIndex551, depth551
			return false
		},
		/* 52 forVar <- <(<identifier> Action30)> */
		func() bool {
			position580, tokenIndex580, depth580 := position, tokenIndex, depth
			{
				position581 := position
				depth++
				{
					position582 := position
					depth++
					if !_rules[ruleidentifier]() {
						goto l580
					}
					depth--
					add(rulePegText, position582)
				}
				if !_rules[ruleAction30]() {
					goto l580
				}
				depth--
				add(ruleforVar, position581)
			}
			return true
		l580:
			position, tokenIndex, depth = position580, tokenIndex580, depth580
			return false
		},
		/* 53 handlers <- <(isp* (fsep isp*)* handler isp* ((fsep isp*)+ handler isp*)* (fsep isp*)* !.)> */
		func() bool {
			position583, tokenIndex583, depth583 := position, tokenIndex, depth
			{
				position584 := position
				depth++
			l585:
				{
					position586, tokenIndex586, depth586 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l586
					}
					goto l585
				l586:
					position, tokenIndex, depth = position586, tokenIndex586, depth586
				}
			l587:
				{
					position588, tokenIndex588, depth588 := position, tokenIndex, depth
					if !_rules[rulefsep]() {
						goto l588
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
					goto l587
				l588:
					position, tokenIndex, depth = position588, tokenIndex588, depth588
				}
				if !_rules[rulehandler]() {
					goto l583
				}
			l591:
				{
					position592, tokenIndex592, depth592 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l592
					}
					goto l591
				l592:
					position, tokenIndex, depth = position592, tokenIndex592, depth592
				}
			l593:
				{
					position594, tokenIndex594, depth594 := position, tokenIndex, depth
					if !_rules[rulefsep]() {
						goto l594
					}
				l597:
					{
						position598, tokenIndex598, depth598 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l598
						}
						goto l597
					l598:
						position, tokenIndex, depth = position598, tokenIndex598, depth598
					}
				l595:
					{
						position596, tokenIndex596, depth596 := position, tokenIndex, depth
						if !_rules[rulefsep]() {
							goto l596
						}
					l599:
						{
							position600, tokenIndex600, depth600 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l600
							}
							goto l599
						l600:
							position, tokenIndex, depth = position600, tokenIndex600, depth600
						}
						goto l595
					l596:
						position, tokenIndex, depth = position596, tokenIndex596, depth596
					}
					if !_rules[rulehandler]() {
						goto l594
					}
				l601:
					{
						position602, tokenIndex602, depth602 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l602
						}
						goto l601
					l602:
						position, tokenIndex, depth = position602, tokenIndex602, depth602
					}
					goto l593
				l594:
					position, tokenIndex, depth = position594, tokenIndex594, depth594
				}
			l603:
				{
					position604, tokenIndex604, depth604 := position, tokenIndex, depth
					if !_rules[rulefsep]() {
						goto l604
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
					goto l603
				l604:
					position, tokenIndex, depth = position604, tokenIndex604, depth604
				}
				{
					position607, tokenIndex607, depth607 := position, tokenIndex, depth
					if !matchDot() {
						goto l607
					}
					goto l583
				l607:
					position, tokenIndex, depth = position607, tokenIndex607, depth607
				}
				depth--
				add(rulehandlers, position584)
			}
			return true
		l583:
			position, tokenIndex, depth = position583, tokenIndex583, depth583
			return false
		},
		/* 54 handler <- <(handlername '(' isp* (param isp* (',' isp* param isp*)*)? ')' (isp* type)? Action31)> */
		func() bool {
			position608, tokenIndex608, depth608 := position, tokenIndex, depth
			{
				position609 := position
				depth++
				if !_rules[rulehandlername]() {
					goto l608
				}
				if buffer[position] != rune('(') {
					goto l608
				}
				position++
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
				{
					position612, tokenIndex612, depth612 := position, tokenIndex, depth
					if !_rules[ruleparam]() {
						goto l612
					}
				l614:
					{
						position615, tokenIndex615, depth615 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l615
						}
						goto l614
					l615:
						position, tokenIndex, depth = position615, tokenIndex615, depth615
					}
				l616:
					{
						position617, tokenIndex617, depth617 := position, tokenIndex, depth
						if buffer[position] != rune(',') {
							goto l617
						}
						position++
					l618:
						{
							position619, tokenIndex619, depth619 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l619
							}
							goto l618
						l619:
							position, tokenIndex, depth = position619, tokenIndex619, depth619
						}
						if !_rules[ruleparam]() {
							goto l617
						}
					l620:
						{
							position621, tokenIndex621, depth621 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l621
							}
							goto l620
						l621:
							position, tokenIndex, depth = position621, tokenIndex621, depth621
						}
						goto l616
					l617:
						position, tokenIndex, depth = position617, tokenIndex617, depth617
					}
					goto l613
				l612:
					position, tokenIndex, depth = position612, tokenIndex612, depth612
				}
			l613:
				if buffer[position] != rune(')') {
					goto l608
				}
				position++
				{
					position622, tokenIndex622, depth622 := position, tokenIndex, depth
				l624:
					{
						position625, tokenIndex625, depth625 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l625
						}
						goto l624
					l625:
						position, tokenIndex, depth = position625, tokenIndex625, depth625
					}
					if !_rules[ruletype]() {
						goto l622
					}
					goto l623
				l622:
					position, tokenIndex, depth = position622, tokenIndex622, depth622
				}
			l623:
				if !_rules[ruleAction31]() {
					goto l608
				}
				depth--
				add(rulehandler, position609)
			}
			return true
		l608:
			position, tokenIndex, depth = position608, tokenIndex608, depth608
			return false
		},
		/* 55 param <- <(tagname isp+ type Action32)> */
		func() bool {
			position626, tokenIndex626, depth626 := position, tokenIndex, depth
			{
				position627 := position
				depth++
				if !_rules[ruletagname]() {
					goto l626
				}
				if !_rules[ruleisp]() {
					goto l626
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
					goto l626
				}
				if !_rules[ruleAction32]() {
					goto l626
				}
				depth--
				add(ruleparam, position627)
			}
			return true
		l626:
			position, tokenIndex, depth = position626, tokenIndex626, depth626
			return false
		},
		/* 56 cparams <- <(isp* (cparam isp* (',' isp* cparam isp*)*)? !.)> */
		func() bool {
			position630, tokenIndex630, depth630 := position, tokenIndex, depth
			{
				position631 := position
				depth++
			l632:
				{
					position633, tokenIndex633, depth633 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l633
					}
					goto l632
				l633:
					position, tokenIndex, depth = position633, tokenIndex633, depth633
				}
				{
					position634, tokenIndex634, depth634 := position, tokenIndex, depth
					if !_rules[rulecparam]() {
						goto l634
					}
				l636:
					{
						position637, tokenIndex637, depth637 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l637
						}
						goto l636
					l637:
						position, tokenIndex, depth = position637, tokenIndex637, depth637
					}
				l638:
					{
						position639, tokenIndex639, depth639 := position, tokenIndex, depth
						if buffer[position] != rune(',') {
							goto l639
						}
						position++
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
						if !_rules[rulecparam]() {
							goto l639
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
						goto l638
					l639:
						position, tokenIndex, depth = position639, tokenIndex639, depth639
					}
					goto l635
				l634:
					position, tokenIndex, depth = position634, tokenIndex634, depth634
				}
			l635:
				{
					position644, tokenIndex644, depth644 := position, tokenIndex, depth
					if !matchDot() {
						goto l644
					}
					goto l630
				l644:
					position, tokenIndex, depth = position644, tokenIndex644, depth644
				}
				depth--
				add(rulecparams, position631)
			}
			return true
		l630:
			position, tokenIndex, depth = position630, tokenIndex630, depth630
			return false
		},
		/* 57 cparam <- <((var isp+)? tagname isp+ type Action33)> */
		func() bool {
			position645, tokenIndex645, depth645 := position, tokenIndex, depth
			{
				position646 := position
				depth++
				{
					position647, tokenIndex647, depth647 := position, tokenIndex, depth
					if !_rules[rulevar]() {
						goto l647
					}
					if !_rules[ruleisp]() {
						goto l647
					}
				l649:
					{
						position650, tokenIndex650, depth650 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l650
						}
						goto l649
					l650:
						position, tokenIndex, depth = position650, tokenIndex650, depth650
					}
					goto l648
				l647:
					position, tokenIndex, depth = position647, tokenIndex647, depth647
				}
			l648:
				if !_rules[ruletagname]() {
					goto l645
				}
				if !_rules[ruleisp]() {
					goto l645
				}
			l651:
				{
					position652, tokenIndex652, depth652 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l652
					}
					goto l651
				l652:
					position, tokenIndex, depth = position652, tokenIndex652, depth652
				}
				if !_rules[ruletype]() {
					goto l645
				}
				if !_rules[ruleAction33]() {
					goto l645
				}
				depth--
				add(rulecparam, position646)
			}
			return true
		l645:
			position, tokenIndex, depth = position645, tokenIndex645, depth645
			return false
		},
		/* 58 var <- <(('v' / 'V') ('a' / 'A') ('r' / 'R') Action34)> */
		func() bool {
			position653, tokenIndex653, depth653 := position, tokenIndex, depth
			{
				position654 := position
				depth++
				{
					position655, tokenIndex655, depth655 := position, tokenIndex, depth
					if buffer[position] != rune('v') {
						goto l656
					}
					position++
					goto l655
				l656:
					position, tokenIndex, depth = position655, tokenIndex655, depth655
					if buffer[position] != rune('V') {
						goto l653
					}
					position++
				}
			l655:
				{
					position657, tokenIndex657, depth657 := position, tokenIndex, depth
					if buffer[position] != rune('a') {
						goto l658
					}
					position++
					goto l657
				l658:
					position, tokenIndex, depth = position657, tokenIndex657, depth657
					if buffer[position] != rune('A') {
						goto l653
					}
					position++
				}
			l657:
				{
					position659, tokenIndex659, depth659 := position, tokenIndex, depth
					if buffer[position] != rune('r') {
						goto l660
					}
					position++
					goto l659
				l660:
					position, tokenIndex, depth = position659, tokenIndex659, depth659
					if buffer[position] != rune('R') {
						goto l653
					}
					position++
				}
			l659:
				if !_rules[ruleAction34]() {
					goto l653
				}
				depth--
				add(rulevar, position654)
			}
			return true
		l653:
			position, tokenIndex, depth = position653, tokenIndex653, depth653
			return false
		},
		/* 59 args <- <(isp* arg isp* (',' isp* arg isp*)* !.)> */
		func() bool {
			position661, tokenIndex661, depth661 := position, tokenIndex, depth
			{
				position662 := position
				depth++
			l663:
				{
					position664, tokenIndex664, depth664 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l664
					}
					goto l663
				l664:
					position, tokenIndex, depth = position664, tokenIndex664, depth664
				}
				if !_rules[rulearg]() {
					goto l661
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
					if buffer[position] != rune(',') {
						goto l668
					}
					position++
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
					if !_rules[rulearg]() {
						goto l668
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
					goto l667
				l668:
					position, tokenIndex, depth = position668, tokenIndex668, depth668
				}
				{
					position673, tokenIndex673, depth673 := position, tokenIndex, depth
					if !matchDot() {
						goto l673
					}
					goto l661
				l673:
					position, tokenIndex, depth = position673, tokenIndex673, depth673
				}
				depth--
				add(ruleargs, position662)
			}
			return true
		l661:
			position, tokenIndex, depth = position661, tokenIndex661, depth661
			return false
		},
		/* 60 arg <- <(expr Action35)> */
		func() bool {
			position674, tokenIndex674, depth674 := position, tokenIndex, depth
			{
				position675 := position
				depth++
				if !_rules[ruleexpr]() {
					goto l674
				}
				if !_rules[ruleAction35]() {
					goto l674
				}
				depth--
				add(rulearg, position675)
			}
			return true
		l674:
			position, tokenIndex, depth = position674, tokenIndex674, depth674
			return false
		},
		/* 61 imports <- <(isp* (fsep isp*)* import isp* (fsep isp* (fsep isp*)* import isp*)* (fsep isp*)* !.)> */
		func() bool {
			position676, tokenIndex676, depth676 := position, tokenIndex, depth
			{
				position677 := position
				depth++
			l678:
				{
					position679, tokenIndex679, depth679 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l679
					}
					goto l678
				l679:
					position, tokenIndex, depth = position679, tokenIndex679, depth679
				}
			l680:
				{
					position681, tokenIndex681, depth681 := position, tokenIndex, depth
					if !_rules[rulefsep]() {
						goto l681
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
					goto l680
				l681:
					position, tokenIndex, depth = position681, tokenIndex681, depth681
				}
				if !_rules[ruleimport]() {
					goto l676
				}
			l684:
				{
					position685, tokenIndex685, depth685 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l685
					}
					goto l684
				l685:
					position, tokenIndex, depth = position685, tokenIndex685, depth685
				}
			l686:
				{
					position687, tokenIndex687, depth687 := position, tokenIndex, depth
					if !_rules[rulefsep]() {
						goto l687
					}
				l688:
					{
						position689, tokenIndex689, depth689 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l689
						}
						goto l688
					l689:
						position, tokenIndex, depth = position689, tokenIndex689, depth689
					}
				l690:
					{
						position691, tokenIndex691, depth691 := position, tokenIndex, depth
						if !_rules[rulefsep]() {
							goto l691
						}
					l692:
						{
							position693, tokenIndex693, depth693 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l693
							}
							goto l692
						l693:
							position, tokenIndex, depth = position693, tokenIndex693, depth693
						}
						goto l690
					l691:
						position, tokenIndex, depth = position691, tokenIndex691, depth691
					}
					if !_rules[ruleimport]() {
						goto l687
					}
				l694:
					{
						position695, tokenIndex695, depth695 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l695
						}
						goto l694
					l695:
						position, tokenIndex, depth = position695, tokenIndex695, depth695
					}
					goto l686
				l687:
					position, tokenIndex, depth = position687, tokenIndex687, depth687
				}
			l696:
				{
					position697, tokenIndex697, depth697 := position, tokenIndex, depth
					if !_rules[rulefsep]() {
						goto l697
					}
				l698:
					{
						position699, tokenIndex699, depth699 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l699
						}
						goto l698
					l699:
						position, tokenIndex, depth = position699, tokenIndex699, depth699
					}
					goto l696
				l697:
					position, tokenIndex, depth = position697, tokenIndex697, depth697
				}
				{
					position700, tokenIndex700, depth700 := position, tokenIndex, depth
					if !matchDot() {
						goto l700
					}
					goto l676
				l700:
					position, tokenIndex, depth = position700, tokenIndex700, depth700
				}
				depth--
				add(ruleimports, position677)
			}
			return true
		l676:
			position, tokenIndex, depth = position676, tokenIndex676, depth676
			return false
		},
		/* 62 import <- <((tagname isp+)? '"' <(!'"' .)*> '"' Action36)> */
		func() bool {
			position701, tokenIndex701, depth701 := position, tokenIndex, depth
			{
				position702 := position
				depth++
				{
					position703, tokenIndex703, depth703 := position, tokenIndex, depth
					if !_rules[ruletagname]() {
						goto l703
					}
					if !_rules[ruleisp]() {
						goto l703
					}
				l705:
					{
						position706, tokenIndex706, depth706 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l706
						}
						goto l705
					l706:
						position, tokenIndex, depth = position706, tokenIndex706, depth706
					}
					goto l704
				l703:
					position, tokenIndex, depth = position703, tokenIndex703, depth703
				}
			l704:
				if buffer[position] != rune('"') {
					goto l701
				}
				position++
				{
					position707 := position
					depth++
				l708:
					{
						position709, tokenIndex709, depth709 := position, tokenIndex, depth
						{
							position710, tokenIndex710, depth710 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l710
							}
							position++
							goto l709
						l710:
							position, tokenIndex, depth = position710, tokenIndex710, depth710
						}
						if !matchDot() {
							goto l709
						}
						goto l708
					l709:
						position, tokenIndex, depth = position709, tokenIndex709, depth709
					}
					depth--
					add(rulePegText, position707)
				}
				if buffer[position] != rune('"') {
					goto l701
				}
				position++
				if !_rules[ruleAction36]() {
					goto l701
				}
				depth--
				add(ruleimport, position702)
			}
			return true
		l701:
			position, tokenIndex, depth = position701, tokenIndex701, depth701
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
			p.bv.Kind = data.BoundData
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
