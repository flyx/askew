package binding

import (
	"errors"
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

type BindingParser struct {
	eventHandling              data.EventHandling
	expr, tagname, handlername string
	names                      []string
	keytype, valuetype         *data.ParamType
	fields                     []*data.Field
	bv                         data.BoundValue
	goVal                      data.GoValue
	paramMappings              map[string]data.BoundValue
	err                        error

	assignments   []data.Assignment
	varMappings   []data.VariableMapping
	eventMappings []data.UnboundEventMapping

	Buffer string
	buffer []rune
	rules  [83]func() bool
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
	p   *BindingParser
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

func (p *BindingParser) PrintSyntaxTree() {
	p.tokens32.PrintSyntaxTree(p.Buffer)
}

func (p *BindingParser) Highlighter() {
	p.PrintSyntax()
}

func (p *BindingParser) Execute() {
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

		}
	}
	_, _, _, _, _ = buffer, _buffer, text, begin, end
}

func (p *BindingParser) Init() {
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
		/* 0 e <- <(assignments / bindings / captures / fields)> */
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
			position6, tokenIndex6, depth6 := position, tokenIndex, depth
			{
				position7 := position
				depth++
			l8:
				{
					position9, tokenIndex9, depth9 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l9
					}
					goto l8
				l9:
					position, tokenIndex, depth = position9, tokenIndex9, depth9
				}
				if !_rules[ruleassignment]() {
					goto l6
				}
			l10:
				{
					position11, tokenIndex11, depth11 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l11
					}
					goto l10
				l11:
					position, tokenIndex, depth = position11, tokenIndex11, depth11
				}
			l12:
				{
					position13, tokenIndex13, depth13 := position, tokenIndex, depth
					{
						position14, tokenIndex14, depth14 := position, tokenIndex, depth
						if buffer[position] != rune(',') {
							goto l15
						}
						position++
						goto l14
					l15:
						position, tokenIndex, depth = position14, tokenIndex14, depth14
						if buffer[position] != rune(';') {
							goto l13
						}
						position++
					}
				l14:
				l16:
					{
						position17, tokenIndex17, depth17 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l17
						}
						goto l16
					l17:
						position, tokenIndex, depth = position17, tokenIndex17, depth17
					}
					if !_rules[ruleassignment]() {
						goto l13
					}
				l18:
					{
						position19, tokenIndex19, depth19 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l19
						}
						goto l18
					l19:
						position, tokenIndex, depth = position19, tokenIndex19, depth19
					}
					goto l12
				l13:
					position, tokenIndex, depth = position13, tokenIndex13, depth13
				}
				{
					position20, tokenIndex20, depth20 := position, tokenIndex, depth
					if !matchDot() {
						goto l20
					}
					goto l6
				l20:
					position, tokenIndex, depth = position20, tokenIndex20, depth20
				}
				depth--
				add(ruleassignments, position7)
			}
			return true
		l6:
			position, tokenIndex, depth = position6, tokenIndex6, depth6
			return false
		},
		/* 2 bindings <- <(isp* binding isp* ((',' / ';') isp* binding isp*)* !.)> */
		func() bool {
			position21, tokenIndex21, depth21 := position, tokenIndex, depth
			{
				position22 := position
				depth++
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
				if !_rules[rulebinding]() {
					goto l21
				}
			l25:
				{
					position26, tokenIndex26, depth26 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l26
					}
					goto l25
				l26:
					position, tokenIndex, depth = position26, tokenIndex26, depth26
				}
			l27:
				{
					position28, tokenIndex28, depth28 := position, tokenIndex, depth
					{
						position29, tokenIndex29, depth29 := position, tokenIndex, depth
						if buffer[position] != rune(',') {
							goto l30
						}
						position++
						goto l29
					l30:
						position, tokenIndex, depth = position29, tokenIndex29, depth29
						if buffer[position] != rune(';') {
							goto l28
						}
						position++
					}
				l29:
				l31:
					{
						position32, tokenIndex32, depth32 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l32
						}
						goto l31
					l32:
						position, tokenIndex, depth = position32, tokenIndex32, depth32
					}
					if !_rules[rulebinding]() {
						goto l28
					}
				l33:
					{
						position34, tokenIndex34, depth34 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l34
						}
						goto l33
					l34:
						position, tokenIndex, depth = position34, tokenIndex34, depth34
					}
					goto l27
				l28:
					position, tokenIndex, depth = position28, tokenIndex28, depth28
				}
				{
					position35, tokenIndex35, depth35 := position, tokenIndex, depth
					if !matchDot() {
						goto l35
					}
					goto l21
				l35:
					position, tokenIndex, depth = position35, tokenIndex35, depth35
				}
				depth--
				add(rulebindings, position22)
			}
			return true
		l21:
			position, tokenIndex, depth = position21, tokenIndex21, depth21
			return false
		},
		/* 3 binding <- <(bound isp* ':' isp* (autovar / typedvar) Action0)> */
		func() bool {
			position36, tokenIndex36, depth36 := position, tokenIndex, depth
			{
				position37 := position
				depth++
				if !_rules[rulebound]() {
					goto l36
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
				if buffer[position] != rune(':') {
					goto l36
				}
				position++
			l40:
				{
					position41, tokenIndex41, depth41 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l41
					}
					goto l40
				l41:
					position, tokenIndex, depth = position41, tokenIndex41, depth41
				}
				{
					position42, tokenIndex42, depth42 := position, tokenIndex, depth
					if !_rules[ruleautovar]() {
						goto l43
					}
					goto l42
				l43:
					position, tokenIndex, depth = position42, tokenIndex42, depth42
					if !_rules[ruletypedvar]() {
						goto l36
					}
				}
			l42:
				if !_rules[ruleAction0]() {
					goto l36
				}
				depth--
				add(rulebinding, position37)
			}
			return true
		l36:
			position, tokenIndex, depth = position36, tokenIndex36, depth36
			return false
		},
		/* 4 autovar <- <(<identifier> Action1)> */
		func() bool {
			position44, tokenIndex44, depth44 := position, tokenIndex, depth
			{
				position45 := position
				depth++
				{
					position46 := position
					depth++
					if !_rules[ruleidentifier]() {
						goto l44
					}
					depth--
					add(rulePegText, position46)
				}
				if !_rules[ruleAction1]() {
					goto l44
				}
				depth--
				add(ruleautovar, position45)
			}
			return true
		l44:
			position, tokenIndex, depth = position44, tokenIndex44, depth44
			return false
		},
		/* 5 typedvar <- <('(' isp* autovar isp+ type isp* ')' Action2)> */
		func() bool {
			position47, tokenIndex47, depth47 := position, tokenIndex, depth
			{
				position48 := position
				depth++
				if buffer[position] != rune('(') {
					goto l47
				}
				position++
			l49:
				{
					position50, tokenIndex50, depth50 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l50
					}
					goto l49
				l50:
					position, tokenIndex, depth = position50, tokenIndex50, depth50
				}
				if !_rules[ruleautovar]() {
					goto l47
				}
				if !_rules[ruleisp]() {
					goto l47
				}
			l51:
				{
					position52, tokenIndex52, depth52 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l52
					}
					goto l51
				l52:
					position, tokenIndex, depth = position52, tokenIndex52, depth52
				}
				if !_rules[ruletype]() {
					goto l47
				}
			l53:
				{
					position54, tokenIndex54, depth54 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l54
					}
					goto l53
				l54:
					position, tokenIndex, depth = position54, tokenIndex54, depth54
				}
				if buffer[position] != rune(')') {
					goto l47
				}
				position++
				if !_rules[ruleAction2]() {
					goto l47
				}
				depth--
				add(ruletypedvar, position48)
			}
			return true
		l47:
			position, tokenIndex, depth = position47, tokenIndex47, depth47
			return false
		},
		/* 6 isp <- <(' ' / '\t')> */
		func() bool {
			position55, tokenIndex55, depth55 := position, tokenIndex, depth
			{
				position56 := position
				depth++
				{
					position57, tokenIndex57, depth57 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l58
					}
					position++
					goto l57
				l58:
					position, tokenIndex, depth = position57, tokenIndex57, depth57
					if buffer[position] != rune('\t') {
						goto l55
					}
					position++
				}
			l57:
				depth--
				add(ruleisp, position56)
			}
			return true
		l55:
			position, tokenIndex, depth = position55, tokenIndex55, depth55
			return false
		},
		/* 7 assignment <- <(isp* bound isp* '=' isp* expr Action3)> */
		func() bool {
			position59, tokenIndex59, depth59 := position, tokenIndex, depth
			{
				position60 := position
				depth++
			l61:
				{
					position62, tokenIndex62, depth62 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l62
					}
					goto l61
				l62:
					position, tokenIndex, depth = position62, tokenIndex62, depth62
				}
				if !_rules[rulebound]() {
					goto l59
				}
			l63:
				{
					position64, tokenIndex64, depth64 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l64
					}
					goto l63
				l64:
					position, tokenIndex, depth = position64, tokenIndex64, depth64
				}
				if buffer[position] != rune('=') {
					goto l59
				}
				position++
			l65:
				{
					position66, tokenIndex66, depth66 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l66
					}
					goto l65
				l66:
					position, tokenIndex, depth = position66, tokenIndex66, depth66
				}
				if !_rules[ruleexpr]() {
					goto l59
				}
				if !_rules[ruleAction3]() {
					goto l59
				}
				depth--
				add(ruleassignment, position60)
			}
			return true
		l59:
			position, tokenIndex, depth = position59, tokenIndex59, depth59
			return false
		},
		/* 8 bound <- <(self / ((&('E' | 'e') event) | (&('F' | 'f') form) | (&('C' | 'c') class) | (&('S' | 's') style) | (&('P' | 'p') prop) | (&('D' | 'd') data)))> */
		func() bool {
			position67, tokenIndex67, depth67 := position, tokenIndex, depth
			{
				position68 := position
				depth++
				{
					position69, tokenIndex69, depth69 := position, tokenIndex, depth
					if !_rules[ruleself]() {
						goto l70
					}
					goto l69
				l70:
					position, tokenIndex, depth = position69, tokenIndex69, depth69
					{
						switch buffer[position] {
						case 'E', 'e':
							if !_rules[ruleevent]() {
								goto l67
							}
							break
						case 'F', 'f':
							if !_rules[ruleform]() {
								goto l67
							}
							break
						case 'C', 'c':
							if !_rules[ruleclass]() {
								goto l67
							}
							break
						case 'S', 's':
							if !_rules[rulestyle]() {
								goto l67
							}
							break
						case 'P', 'p':
							if !_rules[ruleprop]() {
								goto l67
							}
							break
						default:
							if !_rules[ruledata]() {
								goto l67
							}
							break
						}
					}

				}
			l69:
				depth--
				add(rulebound, position68)
			}
			return true
		l67:
			position, tokenIndex, depth = position67, tokenIndex67, depth67
			return false
		},
		/* 9 self <- <(('s' / 'S') ('e' / 'E') ('l' / 'L') ('f' / 'F') isp* '(' isp* ')' Action4)> */
		func() bool {
			position72, tokenIndex72, depth72 := position, tokenIndex, depth
			{
				position73 := position
				depth++
				{
					position74, tokenIndex74, depth74 := position, tokenIndex, depth
					if buffer[position] != rune('s') {
						goto l75
					}
					position++
					goto l74
				l75:
					position, tokenIndex, depth = position74, tokenIndex74, depth74
					if buffer[position] != rune('S') {
						goto l72
					}
					position++
				}
			l74:
				{
					position76, tokenIndex76, depth76 := position, tokenIndex, depth
					if buffer[position] != rune('e') {
						goto l77
					}
					position++
					goto l76
				l77:
					position, tokenIndex, depth = position76, tokenIndex76, depth76
					if buffer[position] != rune('E') {
						goto l72
					}
					position++
				}
			l76:
				{
					position78, tokenIndex78, depth78 := position, tokenIndex, depth
					if buffer[position] != rune('l') {
						goto l79
					}
					position++
					goto l78
				l79:
					position, tokenIndex, depth = position78, tokenIndex78, depth78
					if buffer[position] != rune('L') {
						goto l72
					}
					position++
				}
			l78:
				{
					position80, tokenIndex80, depth80 := position, tokenIndex, depth
					if buffer[position] != rune('f') {
						goto l81
					}
					position++
					goto l80
				l81:
					position, tokenIndex, depth = position80, tokenIndex80, depth80
					if buffer[position] != rune('F') {
						goto l72
					}
					position++
				}
			l80:
			l82:
				{
					position83, tokenIndex83, depth83 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l83
					}
					goto l82
				l83:
					position, tokenIndex, depth = position83, tokenIndex83, depth83
				}
				if buffer[position] != rune('(') {
					goto l72
				}
				position++
			l84:
				{
					position85, tokenIndex85, depth85 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l85
					}
					goto l84
				l85:
					position, tokenIndex, depth = position85, tokenIndex85, depth85
				}
				if buffer[position] != rune(')') {
					goto l72
				}
				position++
				if !_rules[ruleAction4]() {
					goto l72
				}
				depth--
				add(ruleself, position73)
			}
			return true
		l72:
			position, tokenIndex, depth = position72, tokenIndex72, depth72
			return false
		},
		/* 10 data <- <(('d' / 'D') ('a' / 'A') ('t' / 'T') ('a' / 'A') isp* '(' isp* htmlid isp* ')' Action5)> */
		func() bool {
			position86, tokenIndex86, depth86 := position, tokenIndex, depth
			{
				position87 := position
				depth++
				{
					position88, tokenIndex88, depth88 := position, tokenIndex, depth
					if buffer[position] != rune('d') {
						goto l89
					}
					position++
					goto l88
				l89:
					position, tokenIndex, depth = position88, tokenIndex88, depth88
					if buffer[position] != rune('D') {
						goto l86
					}
					position++
				}
			l88:
				{
					position90, tokenIndex90, depth90 := position, tokenIndex, depth
					if buffer[position] != rune('a') {
						goto l91
					}
					position++
					goto l90
				l91:
					position, tokenIndex, depth = position90, tokenIndex90, depth90
					if buffer[position] != rune('A') {
						goto l86
					}
					position++
				}
			l90:
				{
					position92, tokenIndex92, depth92 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l93
					}
					position++
					goto l92
				l93:
					position, tokenIndex, depth = position92, tokenIndex92, depth92
					if buffer[position] != rune('T') {
						goto l86
					}
					position++
				}
			l92:
				{
					position94, tokenIndex94, depth94 := position, tokenIndex, depth
					if buffer[position] != rune('a') {
						goto l95
					}
					position++
					goto l94
				l95:
					position, tokenIndex, depth = position94, tokenIndex94, depth94
					if buffer[position] != rune('A') {
						goto l86
					}
					position++
				}
			l94:
			l96:
				{
					position97, tokenIndex97, depth97 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l97
					}
					goto l96
				l97:
					position, tokenIndex, depth = position97, tokenIndex97, depth97
				}
				if buffer[position] != rune('(') {
					goto l86
				}
				position++
			l98:
				{
					position99, tokenIndex99, depth99 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l99
					}
					goto l98
				l99:
					position, tokenIndex, depth = position99, tokenIndex99, depth99
				}
				if !_rules[rulehtmlid]() {
					goto l86
				}
			l100:
				{
					position101, tokenIndex101, depth101 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l101
					}
					goto l100
				l101:
					position, tokenIndex, depth = position101, tokenIndex101, depth101
				}
				if buffer[position] != rune(')') {
					goto l86
				}
				position++
				if !_rules[ruleAction5]() {
					goto l86
				}
				depth--
				add(ruledata, position87)
			}
			return true
		l86:
			position, tokenIndex, depth = position86, tokenIndex86, depth86
			return false
		},
		/* 11 prop <- <(('p' / 'P') ('r' / 'R') ('o' / 'O') ('p' / 'P') isp* '(' isp* htmlid isp* ')' Action6)> */
		func() bool {
			position102, tokenIndex102, depth102 := position, tokenIndex, depth
			{
				position103 := position
				depth++
				{
					position104, tokenIndex104, depth104 := position, tokenIndex, depth
					if buffer[position] != rune('p') {
						goto l105
					}
					position++
					goto l104
				l105:
					position, tokenIndex, depth = position104, tokenIndex104, depth104
					if buffer[position] != rune('P') {
						goto l102
					}
					position++
				}
			l104:
				{
					position106, tokenIndex106, depth106 := position, tokenIndex, depth
					if buffer[position] != rune('r') {
						goto l107
					}
					position++
					goto l106
				l107:
					position, tokenIndex, depth = position106, tokenIndex106, depth106
					if buffer[position] != rune('R') {
						goto l102
					}
					position++
				}
			l106:
				{
					position108, tokenIndex108, depth108 := position, tokenIndex, depth
					if buffer[position] != rune('o') {
						goto l109
					}
					position++
					goto l108
				l109:
					position, tokenIndex, depth = position108, tokenIndex108, depth108
					if buffer[position] != rune('O') {
						goto l102
					}
					position++
				}
			l108:
				{
					position110, tokenIndex110, depth110 := position, tokenIndex, depth
					if buffer[position] != rune('p') {
						goto l111
					}
					position++
					goto l110
				l111:
					position, tokenIndex, depth = position110, tokenIndex110, depth110
					if buffer[position] != rune('P') {
						goto l102
					}
					position++
				}
			l110:
			l112:
				{
					position113, tokenIndex113, depth113 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l113
					}
					goto l112
				l113:
					position, tokenIndex, depth = position113, tokenIndex113, depth113
				}
				if buffer[position] != rune('(') {
					goto l102
				}
				position++
			l114:
				{
					position115, tokenIndex115, depth115 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l115
					}
					goto l114
				l115:
					position, tokenIndex, depth = position115, tokenIndex115, depth115
				}
				if !_rules[rulehtmlid]() {
					goto l102
				}
			l116:
				{
					position117, tokenIndex117, depth117 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l117
					}
					goto l116
				l117:
					position, tokenIndex, depth = position117, tokenIndex117, depth117
				}
				if buffer[position] != rune(')') {
					goto l102
				}
				position++
				if !_rules[ruleAction6]() {
					goto l102
				}
				depth--
				add(ruleprop, position103)
			}
			return true
		l102:
			position, tokenIndex, depth = position102, tokenIndex102, depth102
			return false
		},
		/* 12 style <- <(('s' / 'S') ('t' / 'T') ('y' / 'Y') ('l' / 'L') ('e' / 'E') isp* '(' isp* htmlid isp* ')' Action7)> */
		func() bool {
			position118, tokenIndex118, depth118 := position, tokenIndex, depth
			{
				position119 := position
				depth++
				{
					position120, tokenIndex120, depth120 := position, tokenIndex, depth
					if buffer[position] != rune('s') {
						goto l121
					}
					position++
					goto l120
				l121:
					position, tokenIndex, depth = position120, tokenIndex120, depth120
					if buffer[position] != rune('S') {
						goto l118
					}
					position++
				}
			l120:
				{
					position122, tokenIndex122, depth122 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l123
					}
					position++
					goto l122
				l123:
					position, tokenIndex, depth = position122, tokenIndex122, depth122
					if buffer[position] != rune('T') {
						goto l118
					}
					position++
				}
			l122:
				{
					position124, tokenIndex124, depth124 := position, tokenIndex, depth
					if buffer[position] != rune('y') {
						goto l125
					}
					position++
					goto l124
				l125:
					position, tokenIndex, depth = position124, tokenIndex124, depth124
					if buffer[position] != rune('Y') {
						goto l118
					}
					position++
				}
			l124:
				{
					position126, tokenIndex126, depth126 := position, tokenIndex, depth
					if buffer[position] != rune('l') {
						goto l127
					}
					position++
					goto l126
				l127:
					position, tokenIndex, depth = position126, tokenIndex126, depth126
					if buffer[position] != rune('L') {
						goto l118
					}
					position++
				}
			l126:
				{
					position128, tokenIndex128, depth128 := position, tokenIndex, depth
					if buffer[position] != rune('e') {
						goto l129
					}
					position++
					goto l128
				l129:
					position, tokenIndex, depth = position128, tokenIndex128, depth128
					if buffer[position] != rune('E') {
						goto l118
					}
					position++
				}
			l128:
			l130:
				{
					position131, tokenIndex131, depth131 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l131
					}
					goto l130
				l131:
					position, tokenIndex, depth = position131, tokenIndex131, depth131
				}
				if buffer[position] != rune('(') {
					goto l118
				}
				position++
			l132:
				{
					position133, tokenIndex133, depth133 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l133
					}
					goto l132
				l133:
					position, tokenIndex, depth = position133, tokenIndex133, depth133
				}
				if !_rules[rulehtmlid]() {
					goto l118
				}
			l134:
				{
					position135, tokenIndex135, depth135 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l135
					}
					goto l134
				l135:
					position, tokenIndex, depth = position135, tokenIndex135, depth135
				}
				if buffer[position] != rune(')') {
					goto l118
				}
				position++
				if !_rules[ruleAction7]() {
					goto l118
				}
				depth--
				add(rulestyle, position119)
			}
			return true
		l118:
			position, tokenIndex, depth = position118, tokenIndex118, depth118
			return false
		},
		/* 13 class <- <(('c' / 'C') ('l' / 'L') ('a' / 'A') ('s' / 'S') ('s' / 'S') isp* '(' isp* htmlid isp* (',' isp* htmlid isp*)* ')' Action8)> */
		func() bool {
			position136, tokenIndex136, depth136 := position, tokenIndex, depth
			{
				position137 := position
				depth++
				{
					position138, tokenIndex138, depth138 := position, tokenIndex, depth
					if buffer[position] != rune('c') {
						goto l139
					}
					position++
					goto l138
				l139:
					position, tokenIndex, depth = position138, tokenIndex138, depth138
					if buffer[position] != rune('C') {
						goto l136
					}
					position++
				}
			l138:
				{
					position140, tokenIndex140, depth140 := position, tokenIndex, depth
					if buffer[position] != rune('l') {
						goto l141
					}
					position++
					goto l140
				l141:
					position, tokenIndex, depth = position140, tokenIndex140, depth140
					if buffer[position] != rune('L') {
						goto l136
					}
					position++
				}
			l140:
				{
					position142, tokenIndex142, depth142 := position, tokenIndex, depth
					if buffer[position] != rune('a') {
						goto l143
					}
					position++
					goto l142
				l143:
					position, tokenIndex, depth = position142, tokenIndex142, depth142
					if buffer[position] != rune('A') {
						goto l136
					}
					position++
				}
			l142:
				{
					position144, tokenIndex144, depth144 := position, tokenIndex, depth
					if buffer[position] != rune('s') {
						goto l145
					}
					position++
					goto l144
				l145:
					position, tokenIndex, depth = position144, tokenIndex144, depth144
					if buffer[position] != rune('S') {
						goto l136
					}
					position++
				}
			l144:
				{
					position146, tokenIndex146, depth146 := position, tokenIndex, depth
					if buffer[position] != rune('s') {
						goto l147
					}
					position++
					goto l146
				l147:
					position, tokenIndex, depth = position146, tokenIndex146, depth146
					if buffer[position] != rune('S') {
						goto l136
					}
					position++
				}
			l146:
			l148:
				{
					position149, tokenIndex149, depth149 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l149
					}
					goto l148
				l149:
					position, tokenIndex, depth = position149, tokenIndex149, depth149
				}
				if buffer[position] != rune('(') {
					goto l136
				}
				position++
			l150:
				{
					position151, tokenIndex151, depth151 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l151
					}
					goto l150
				l151:
					position, tokenIndex, depth = position151, tokenIndex151, depth151
				}
				if !_rules[rulehtmlid]() {
					goto l136
				}
			l152:
				{
					position153, tokenIndex153, depth153 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l153
					}
					goto l152
				l153:
					position, tokenIndex, depth = position153, tokenIndex153, depth153
				}
			l154:
				{
					position155, tokenIndex155, depth155 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l155
					}
					position++
				l156:
					{
						position157, tokenIndex157, depth157 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l157
						}
						goto l156
					l157:
						position, tokenIndex, depth = position157, tokenIndex157, depth157
					}
					if !_rules[rulehtmlid]() {
						goto l155
					}
				l158:
					{
						position159, tokenIndex159, depth159 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l159
						}
						goto l158
					l159:
						position, tokenIndex, depth = position159, tokenIndex159, depth159
					}
					goto l154
				l155:
					position, tokenIndex, depth = position155, tokenIndex155, depth155
				}
				if buffer[position] != rune(')') {
					goto l136
				}
				position++
				if !_rules[ruleAction8]() {
					goto l136
				}
				depth--
				add(ruleclass, position137)
			}
			return true
		l136:
			position, tokenIndex, depth = position136, tokenIndex136, depth136
			return false
		},
		/* 14 form <- <(('f' / 'F') ('o' / 'O') ('r' / 'R') ('m' / 'M') isp* '(' isp* htmlid isp* ')' Action9)> */
		func() bool {
			position160, tokenIndex160, depth160 := position, tokenIndex, depth
			{
				position161 := position
				depth++
				{
					position162, tokenIndex162, depth162 := position, tokenIndex, depth
					if buffer[position] != rune('f') {
						goto l163
					}
					position++
					goto l162
				l163:
					position, tokenIndex, depth = position162, tokenIndex162, depth162
					if buffer[position] != rune('F') {
						goto l160
					}
					position++
				}
			l162:
				{
					position164, tokenIndex164, depth164 := position, tokenIndex, depth
					if buffer[position] != rune('o') {
						goto l165
					}
					position++
					goto l164
				l165:
					position, tokenIndex, depth = position164, tokenIndex164, depth164
					if buffer[position] != rune('O') {
						goto l160
					}
					position++
				}
			l164:
				{
					position166, tokenIndex166, depth166 := position, tokenIndex, depth
					if buffer[position] != rune('r') {
						goto l167
					}
					position++
					goto l166
				l167:
					position, tokenIndex, depth = position166, tokenIndex166, depth166
					if buffer[position] != rune('R') {
						goto l160
					}
					position++
				}
			l166:
				{
					position168, tokenIndex168, depth168 := position, tokenIndex, depth
					if buffer[position] != rune('m') {
						goto l169
					}
					position++
					goto l168
				l169:
					position, tokenIndex, depth = position168, tokenIndex168, depth168
					if buffer[position] != rune('M') {
						goto l160
					}
					position++
				}
			l168:
			l170:
				{
					position171, tokenIndex171, depth171 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l171
					}
					goto l170
				l171:
					position, tokenIndex, depth = position171, tokenIndex171, depth171
				}
				if buffer[position] != rune('(') {
					goto l160
				}
				position++
			l172:
				{
					position173, tokenIndex173, depth173 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l173
					}
					goto l172
				l173:
					position, tokenIndex, depth = position173, tokenIndex173, depth173
				}
				if !_rules[rulehtmlid]() {
					goto l160
				}
			l174:
				{
					position175, tokenIndex175, depth175 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l175
					}
					goto l174
				l175:
					position, tokenIndex, depth = position175, tokenIndex175, depth175
				}
				if buffer[position] != rune(')') {
					goto l160
				}
				position++
				if !_rules[ruleAction9]() {
					goto l160
				}
				depth--
				add(ruleform, position161)
			}
			return true
		l160:
			position, tokenIndex, depth = position160, tokenIndex160, depth160
			return false
		},
		/* 15 event <- <(('e' / 'E') ('v' / 'V') ('e' / 'E') ('n' / 'N') ('t' / 'T') isp* '(' isp* jsid? isp* ')' Action10)> */
		func() bool {
			position176, tokenIndex176, depth176 := position, tokenIndex, depth
			{
				position177 := position
				depth++
				{
					position178, tokenIndex178, depth178 := position, tokenIndex, depth
					if buffer[position] != rune('e') {
						goto l179
					}
					position++
					goto l178
				l179:
					position, tokenIndex, depth = position178, tokenIndex178, depth178
					if buffer[position] != rune('E') {
						goto l176
					}
					position++
				}
			l178:
				{
					position180, tokenIndex180, depth180 := position, tokenIndex, depth
					if buffer[position] != rune('v') {
						goto l181
					}
					position++
					goto l180
				l181:
					position, tokenIndex, depth = position180, tokenIndex180, depth180
					if buffer[position] != rune('V') {
						goto l176
					}
					position++
				}
			l180:
				{
					position182, tokenIndex182, depth182 := position, tokenIndex, depth
					if buffer[position] != rune('e') {
						goto l183
					}
					position++
					goto l182
				l183:
					position, tokenIndex, depth = position182, tokenIndex182, depth182
					if buffer[position] != rune('E') {
						goto l176
					}
					position++
				}
			l182:
				{
					position184, tokenIndex184, depth184 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l185
					}
					position++
					goto l184
				l185:
					position, tokenIndex, depth = position184, tokenIndex184, depth184
					if buffer[position] != rune('N') {
						goto l176
					}
					position++
				}
			l184:
				{
					position186, tokenIndex186, depth186 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l187
					}
					position++
					goto l186
				l187:
					position, tokenIndex, depth = position186, tokenIndex186, depth186
					if buffer[position] != rune('T') {
						goto l176
					}
					position++
				}
			l186:
			l188:
				{
					position189, tokenIndex189, depth189 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l189
					}
					goto l188
				l189:
					position, tokenIndex, depth = position189, tokenIndex189, depth189
				}
				if buffer[position] != rune('(') {
					goto l176
				}
				position++
			l190:
				{
					position191, tokenIndex191, depth191 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l191
					}
					goto l190
				l191:
					position, tokenIndex, depth = position191, tokenIndex191, depth191
				}
				{
					position192, tokenIndex192, depth192 := position, tokenIndex, depth
					if !_rules[rulejsid]() {
						goto l192
					}
					goto l193
				l192:
					position, tokenIndex, depth = position192, tokenIndex192, depth192
				}
			l193:
			l194:
				{
					position195, tokenIndex195, depth195 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l195
					}
					goto l194
				l195:
					position, tokenIndex, depth = position195, tokenIndex195, depth195
				}
				if buffer[position] != rune(')') {
					goto l176
				}
				position++
				if !_rules[ruleAction10]() {
					goto l176
				}
				depth--
				add(ruleevent, position177)
			}
			return true
		l176:
			position, tokenIndex, depth = position176, tokenIndex176, depth176
			return false
		},
		/* 16 htmlid <- <(<((&('-') '-') | (&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))+> Action11)> */
		func() bool {
			position196, tokenIndex196, depth196 := position, tokenIndex, depth
			{
				position197 := position
				depth++
				{
					position198 := position
					depth++
					{
						switch buffer[position] {
						case '-':
							if buffer[position] != rune('-') {
								goto l196
							}
							position++
							break
						case '_':
							if buffer[position] != rune('_') {
								goto l196
							}
							position++
							break
						case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l196
							}
							position++
							break
						case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l196
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l196
							}
							position++
							break
						}
					}

				l199:
					{
						position200, tokenIndex200, depth200 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '-':
								if buffer[position] != rune('-') {
									goto l200
								}
								position++
								break
							case '_':
								if buffer[position] != rune('_') {
									goto l200
								}
								position++
								break
							case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
								if c := buffer[position]; c < rune('A') || c > rune('Z') {
									goto l200
								}
								position++
								break
							case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l200
								}
								position++
								break
							default:
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l200
								}
								position++
								break
							}
						}

						goto l199
					l200:
						position, tokenIndex, depth = position200, tokenIndex200, depth200
					}
					depth--
					add(rulePegText, position198)
				}
				if !_rules[ruleAction11]() {
					goto l196
				}
				depth--
				add(rulehtmlid, position197)
			}
			return true
		l196:
			position, tokenIndex, depth = position196, tokenIndex196, depth196
			return false
		},
		/* 17 jsid <- <(<(((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z])) ((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))*)> Action12)> */
		func() bool {
			position203, tokenIndex203, depth203 := position, tokenIndex, depth
			{
				position204 := position
				depth++
				{
					position205 := position
					depth++
					{
						switch buffer[position] {
						case '_':
							if buffer[position] != rune('_') {
								goto l203
							}
							position++
							break
						case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l203
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l203
							}
							position++
							break
						}
					}

				l207:
					{
						position208, tokenIndex208, depth208 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '_':
								if buffer[position] != rune('_') {
									goto l208
								}
								position++
								break
							case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
								if c := buffer[position]; c < rune('A') || c > rune('Z') {
									goto l208
								}
								position++
								break
							case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l208
								}
								position++
								break
							default:
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l208
								}
								position++
								break
							}
						}

						goto l207
					l208:
						position, tokenIndex, depth = position208, tokenIndex208, depth208
					}
					depth--
					add(rulePegText, position205)
				}
				if !_rules[ruleAction12]() {
					goto l203
				}
				depth--
				add(rulejsid, position204)
			}
			return true
		l203:
			position, tokenIndex, depth = position203, tokenIndex203, depth203
			return false
		},
		/* 18 expr <- <(<(commaless / enclosed)+> Action13)> */
		func() bool {
			position210, tokenIndex210, depth210 := position, tokenIndex, depth
			{
				position211 := position
				depth++
				{
					position212 := position
					depth++
					{
						position215, tokenIndex215, depth215 := position, tokenIndex, depth
						if !_rules[rulecommaless]() {
							goto l216
						}
						goto l215
					l216:
						position, tokenIndex, depth = position215, tokenIndex215, depth215
						if !_rules[ruleenclosed]() {
							goto l210
						}
					}
				l215:
				l213:
					{
						position214, tokenIndex214, depth214 := position, tokenIndex, depth
						{
							position217, tokenIndex217, depth217 := position, tokenIndex, depth
							if !_rules[rulecommaless]() {
								goto l218
							}
							goto l217
						l218:
							position, tokenIndex, depth = position217, tokenIndex217, depth217
							if !_rules[ruleenclosed]() {
								goto l214
							}
						}
					l217:
						goto l213
					l214:
						position, tokenIndex, depth = position214, tokenIndex214, depth214
					}
					depth--
					add(rulePegText, position212)
				}
				if !_rules[ruleAction13]() {
					goto l210
				}
				depth--
				add(ruleexpr, position211)
			}
			return true
		l210:
			position, tokenIndex, depth = position210, tokenIndex210, depth210
			return false
		},
		/* 19 commaless <- <((&('"' | '`') string) | (&('!' | '&' | '*' | '+' | '-' | '.' | '/' | ':' | '<' | '=' | '>' | '^' | '|') operators) | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') number) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '_' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') identifier))> */
		func() bool {
			position219, tokenIndex219, depth219 := position, tokenIndex, depth
			{
				position220 := position
				depth++
				{
					switch buffer[position] {
					case '"', '`':
						if !_rules[rulestring]() {
							goto l219
						}
						break
					case '!', '&', '*', '+', '-', '.', '/', ':', '<', '=', '>', '^', '|':
						if !_rules[ruleoperators]() {
							goto l219
						}
						break
					case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
						if !_rules[rulenumber]() {
							goto l219
						}
						break
					default:
						if !_rules[ruleidentifier]() {
							goto l219
						}
						break
					}
				}

				depth--
				add(rulecommaless, position220)
			}
			return true
		l219:
			position, tokenIndex, depth = position219, tokenIndex219, depth219
			return false
		},
		/* 20 number <- <[0-9]+> */
		func() bool {
			position222, tokenIndex222, depth222 := position, tokenIndex, depth
			{
				position223 := position
				depth++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l222
				}
				position++
			l224:
				{
					position225, tokenIndex225, depth225 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l225
					}
					position++
					goto l224
				l225:
					position, tokenIndex, depth = position225, tokenIndex225, depth225
				}
				depth--
				add(rulenumber, position223)
			}
			return true
		l222:
			position, tokenIndex, depth = position222, tokenIndex222, depth222
			return false
		},
		/* 21 operators <- <((&('>') '>') | (&('<') '<') | (&('!') '!') | (&('.') '.') | (&('=') '=') | (&(':') ':') | (&('^') '^') | (&('&') '&') | (&('|') '|') | (&('/') '/') | (&('*') '*') | (&('-') '-') | (&('+') '+'))+> */
		func() bool {
			position226, tokenIndex226, depth226 := position, tokenIndex, depth
			{
				position227 := position
				depth++
				{
					switch buffer[position] {
					case '>':
						if buffer[position] != rune('>') {
							goto l226
						}
						position++
						break
					case '<':
						if buffer[position] != rune('<') {
							goto l226
						}
						position++
						break
					case '!':
						if buffer[position] != rune('!') {
							goto l226
						}
						position++
						break
					case '.':
						if buffer[position] != rune('.') {
							goto l226
						}
						position++
						break
					case '=':
						if buffer[position] != rune('=') {
							goto l226
						}
						position++
						break
					case ':':
						if buffer[position] != rune(':') {
							goto l226
						}
						position++
						break
					case '^':
						if buffer[position] != rune('^') {
							goto l226
						}
						position++
						break
					case '&':
						if buffer[position] != rune('&') {
							goto l226
						}
						position++
						break
					case '|':
						if buffer[position] != rune('|') {
							goto l226
						}
						position++
						break
					case '/':
						if buffer[position] != rune('/') {
							goto l226
						}
						position++
						break
					case '*':
						if buffer[position] != rune('*') {
							goto l226
						}
						position++
						break
					case '-':
						if buffer[position] != rune('-') {
							goto l226
						}
						position++
						break
					default:
						if buffer[position] != rune('+') {
							goto l226
						}
						position++
						break
					}
				}

			l228:
				{
					position229, tokenIndex229, depth229 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '>':
							if buffer[position] != rune('>') {
								goto l229
							}
							position++
							break
						case '<':
							if buffer[position] != rune('<') {
								goto l229
							}
							position++
							break
						case '!':
							if buffer[position] != rune('!') {
								goto l229
							}
							position++
							break
						case '.':
							if buffer[position] != rune('.') {
								goto l229
							}
							position++
							break
						case '=':
							if buffer[position] != rune('=') {
								goto l229
							}
							position++
							break
						case ':':
							if buffer[position] != rune(':') {
								goto l229
							}
							position++
							break
						case '^':
							if buffer[position] != rune('^') {
								goto l229
							}
							position++
							break
						case '&':
							if buffer[position] != rune('&') {
								goto l229
							}
							position++
							break
						case '|':
							if buffer[position] != rune('|') {
								goto l229
							}
							position++
							break
						case '/':
							if buffer[position] != rune('/') {
								goto l229
							}
							position++
							break
						case '*':
							if buffer[position] != rune('*') {
								goto l229
							}
							position++
							break
						case '-':
							if buffer[position] != rune('-') {
								goto l229
							}
							position++
							break
						default:
							if buffer[position] != rune('+') {
								goto l229
							}
							position++
							break
						}
					}

					goto l228
				l229:
					position, tokenIndex, depth = position229, tokenIndex229, depth229
				}
				depth--
				add(ruleoperators, position227)
			}
			return true
		l226:
			position, tokenIndex, depth = position226, tokenIndex226, depth226
			return false
		},
		/* 22 string <- <(('`' ('!' / '`')* '`') / ('"' ((&('\\') ('\\' '"')) | (&('"') '"') | (&('!') '!'))* '"'))> */
		func() bool {
			position232, tokenIndex232, depth232 := position, tokenIndex, depth
			{
				position233 := position
				depth++
				{
					position234, tokenIndex234, depth234 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l235
					}
					position++
				l236:
					{
						position237, tokenIndex237, depth237 := position, tokenIndex, depth
						{
							position238, tokenIndex238, depth238 := position, tokenIndex, depth
							if buffer[position] != rune('!') {
								goto l239
							}
							position++
							goto l238
						l239:
							position, tokenIndex, depth = position238, tokenIndex238, depth238
							if buffer[position] != rune('`') {
								goto l237
							}
							position++
						}
					l238:
						goto l236
					l237:
						position, tokenIndex, depth = position237, tokenIndex237, depth237
					}
					if buffer[position] != rune('`') {
						goto l235
					}
					position++
					goto l234
				l235:
					position, tokenIndex, depth = position234, tokenIndex234, depth234
					if buffer[position] != rune('"') {
						goto l232
					}
					position++
				l240:
					{
						position241, tokenIndex241, depth241 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '\\':
								if buffer[position] != rune('\\') {
									goto l241
								}
								position++
								if buffer[position] != rune('"') {
									goto l241
								}
								position++
								break
							case '"':
								if buffer[position] != rune('"') {
									goto l241
								}
								position++
								break
							default:
								if buffer[position] != rune('!') {
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
					if buffer[position] != rune('"') {
						goto l232
					}
					position++
				}
			l234:
				depth--
				add(rulestring, position233)
			}
			return true
		l232:
			position, tokenIndex, depth = position232, tokenIndex232, depth232
			return false
		},
		/* 23 enclosed <- <((&('[') brackets) | (&('{') braces) | (&('(') parens))> */
		func() bool {
			position243, tokenIndex243, depth243 := position, tokenIndex, depth
			{
				position244 := position
				depth++
				{
					switch buffer[position] {
					case '[':
						if !_rules[rulebrackets]() {
							goto l243
						}
						break
					case '{':
						if !_rules[rulebraces]() {
							goto l243
						}
						break
					default:
						if !_rules[ruleparens]() {
							goto l243
						}
						break
					}
				}

				depth--
				add(ruleenclosed, position244)
			}
			return true
		l243:
			position, tokenIndex, depth = position243, tokenIndex243, depth243
			return false
		},
		/* 24 parens <- <('(' inner ')')> */
		func() bool {
			position246, tokenIndex246, depth246 := position, tokenIndex, depth
			{
				position247 := position
				depth++
				if buffer[position] != rune('(') {
					goto l246
				}
				position++
				if !_rules[ruleinner]() {
					goto l246
				}
				if buffer[position] != rune(')') {
					goto l246
				}
				position++
				depth--
				add(ruleparens, position247)
			}
			return true
		l246:
			position, tokenIndex, depth = position246, tokenIndex246, depth246
			return false
		},
		/* 25 braces <- <('{' inner '}')> */
		func() bool {
			position248, tokenIndex248, depth248 := position, tokenIndex, depth
			{
				position249 := position
				depth++
				if buffer[position] != rune('{') {
					goto l248
				}
				position++
				if !_rules[ruleinner]() {
					goto l248
				}
				if buffer[position] != rune('}') {
					goto l248
				}
				position++
				depth--
				add(rulebraces, position249)
			}
			return true
		l248:
			position, tokenIndex, depth = position248, tokenIndex248, depth248
			return false
		},
		/* 26 brackets <- <('[' inner ']')> */
		func() bool {
			position250, tokenIndex250, depth250 := position, tokenIndex, depth
			{
				position251 := position
				depth++
				if buffer[position] != rune('[') {
					goto l250
				}
				position++
				if !_rules[ruleinner]() {
					goto l250
				}
				if buffer[position] != rune(']') {
					goto l250
				}
				position++
				depth--
				add(rulebrackets, position251)
			}
			return true
		l250:
			position, tokenIndex, depth = position250, tokenIndex250, depth250
			return false
		},
		/* 27 inner <- <((&(',') ',') | (&('(' | '[' | '{') enclosed) | (&('!' | '"' | '&' | '*' | '+' | '-' | '.' | '/' | '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9' | ':' | '<' | '=' | '>' | 'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '^' | '_' | '`' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z' | '|') commaless))*> */
		func() bool {
			{
				position253 := position
				depth++
			l254:
				{
					position255, tokenIndex255, depth255 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case ',':
							if buffer[position] != rune(',') {
								goto l255
							}
							position++
							break
						case '(', '[', '{':
							if !_rules[ruleenclosed]() {
								goto l255
							}
							break
						default:
							if !_rules[rulecommaless]() {
								goto l255
							}
							break
						}
					}

					goto l254
				l255:
					position, tokenIndex, depth = position255, tokenIndex255, depth255
				}
				depth--
				add(ruleinner, position253)
			}
			return true
		},
		/* 28 identifier <- <(((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z])) ((&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') ([0-9] / [0-9])) | (&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))*)> */
		func() bool {
			position257, tokenIndex257, depth257 := position, tokenIndex, depth
			{
				position258 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l257
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l257
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l257
						}
						position++
						break
					}
				}

			l260:
				{
					position261, tokenIndex261, depth261 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
							{
								position263, tokenIndex263, depth263 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l264
								}
								position++
								goto l263
							l264:
								position, tokenIndex, depth = position263, tokenIndex263, depth263
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l261
								}
								position++
							}
						l263:
							break
						case '_':
							if buffer[position] != rune('_') {
								goto l261
							}
							position++
							break
						case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l261
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l261
							}
							position++
							break
						}
					}

					goto l260
				l261:
					position, tokenIndex, depth = position261, tokenIndex261, depth261
				}
				depth--
				add(ruleidentifier, position258)
			}
			return true
		l257:
			position, tokenIndex, depth = position257, tokenIndex257, depth257
			return false
		},
		/* 29 fields <- <(((&('\n') '\n') | (&('\t') '\t') | (&(' ') ' ') | (&(';') ';'))* field isp* (fsep isp* (fsep isp*)* field)* ((&('\n') '\n') | (&('\t') '\t') | (&(' ') ' ') | (&(';') ';'))* !.)> */
		func() bool {
			position265, tokenIndex265, depth265 := position, tokenIndex, depth
			{
				position266 := position
				depth++
			l267:
				{
					position268, tokenIndex268, depth268 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '\n':
							if buffer[position] != rune('\n') {
								goto l268
							}
							position++
							break
						case '\t':
							if buffer[position] != rune('\t') {
								goto l268
							}
							position++
							break
						case ' ':
							if buffer[position] != rune(' ') {
								goto l268
							}
							position++
							break
						default:
							if buffer[position] != rune(';') {
								goto l268
							}
							position++
							break
						}
					}

					goto l267
				l268:
					position, tokenIndex, depth = position268, tokenIndex268, depth268
				}
				if !_rules[rulefield]() {
					goto l265
				}
			l270:
				{
					position271, tokenIndex271, depth271 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l271
					}
					goto l270
				l271:
					position, tokenIndex, depth = position271, tokenIndex271, depth271
				}
			l272:
				{
					position273, tokenIndex273, depth273 := position, tokenIndex, depth
					if !_rules[rulefsep]() {
						goto l273
					}
				l274:
					{
						position275, tokenIndex275, depth275 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l275
						}
						goto l274
					l275:
						position, tokenIndex, depth = position275, tokenIndex275, depth275
					}
				l276:
					{
						position277, tokenIndex277, depth277 := position, tokenIndex, depth
						if !_rules[rulefsep]() {
							goto l277
						}
					l278:
						{
							position279, tokenIndex279, depth279 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l279
							}
							goto l278
						l279:
							position, tokenIndex, depth = position279, tokenIndex279, depth279
						}
						goto l276
					l277:
						position, tokenIndex, depth = position277, tokenIndex277, depth277
					}
					if !_rules[rulefield]() {
						goto l273
					}
					goto l272
				l273:
					position, tokenIndex, depth = position273, tokenIndex273, depth273
				}
			l280:
				{
					position281, tokenIndex281, depth281 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '\n':
							if buffer[position] != rune('\n') {
								goto l281
							}
							position++
							break
						case '\t':
							if buffer[position] != rune('\t') {
								goto l281
							}
							position++
							break
						case ' ':
							if buffer[position] != rune(' ') {
								goto l281
							}
							position++
							break
						default:
							if buffer[position] != rune(';') {
								goto l281
							}
							position++
							break
						}
					}

					goto l280
				l281:
					position, tokenIndex, depth = position281, tokenIndex281, depth281
				}
				{
					position283, tokenIndex283, depth283 := position, tokenIndex, depth
					if !matchDot() {
						goto l283
					}
					goto l265
				l283:
					position, tokenIndex, depth = position283, tokenIndex283, depth283
				}
				depth--
				add(rulefields, position266)
			}
			return true
		l265:
			position, tokenIndex, depth = position265, tokenIndex265, depth265
			return false
		},
		/* 30 fsep <- <(';' / '\n')> */
		func() bool {
			position284, tokenIndex284, depth284 := position, tokenIndex, depth
			{
				position285 := position
				depth++
				{
					position286, tokenIndex286, depth286 := position, tokenIndex, depth
					if buffer[position] != rune(';') {
						goto l287
					}
					position++
					goto l286
				l287:
					position, tokenIndex, depth = position286, tokenIndex286, depth286
					if buffer[position] != rune('\n') {
						goto l284
					}
					position++
				}
			l286:
				depth--
				add(rulefsep, position285)
			}
			return true
		l284:
			position, tokenIndex, depth = position284, tokenIndex284, depth284
			return false
		},
		/* 31 field <- <(name (isp* ',' isp* name)* isp+ type isp* ('=' isp* expr)? Action14)> */
		func() bool {
			position288, tokenIndex288, depth288 := position, tokenIndex, depth
			{
				position289 := position
				depth++
				if !_rules[rulename]() {
					goto l288
				}
			l290:
				{
					position291, tokenIndex291, depth291 := position, tokenIndex, depth
				l292:
					{
						position293, tokenIndex293, depth293 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l293
						}
						goto l292
					l293:
						position, tokenIndex, depth = position293, tokenIndex293, depth293
					}
					if buffer[position] != rune(',') {
						goto l291
					}
					position++
				l294:
					{
						position295, tokenIndex295, depth295 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l295
						}
						goto l294
					l295:
						position, tokenIndex, depth = position295, tokenIndex295, depth295
					}
					if !_rules[rulename]() {
						goto l291
					}
					goto l290
				l291:
					position, tokenIndex, depth = position291, tokenIndex291, depth291
				}
				if !_rules[ruleisp]() {
					goto l288
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
				if !_rules[ruletype]() {
					goto l288
				}
			l298:
				{
					position299, tokenIndex299, depth299 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l299
					}
					goto l298
				l299:
					position, tokenIndex, depth = position299, tokenIndex299, depth299
				}
				{
					position300, tokenIndex300, depth300 := position, tokenIndex, depth
					if buffer[position] != rune('=') {
						goto l300
					}
					position++
				l302:
					{
						position303, tokenIndex303, depth303 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l303
						}
						goto l302
					l303:
						position, tokenIndex, depth = position303, tokenIndex303, depth303
					}
					if !_rules[ruleexpr]() {
						goto l300
					}
					goto l301
				l300:
					position, tokenIndex, depth = position300, tokenIndex300, depth300
				}
			l301:
				if !_rules[ruleAction14]() {
					goto l288
				}
				depth--
				add(rulefield, position289)
			}
			return true
		l288:
			position, tokenIndex, depth = position288, tokenIndex288, depth288
			return false
		},
		/* 32 name <- <(<((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))+> Action15)> */
		func() bool {
			position304, tokenIndex304, depth304 := position, tokenIndex, depth
			{
				position305 := position
				depth++
				{
					position306 := position
					depth++
					{
						switch buffer[position] {
						case '_':
							if buffer[position] != rune('_') {
								goto l304
							}
							position++
							break
						case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l304
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l304
							}
							position++
							break
						}
					}

				l307:
					{
						position308, tokenIndex308, depth308 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '_':
								if buffer[position] != rune('_') {
									goto l308
								}
								position++
								break
							case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
								if c := buffer[position]; c < rune('A') || c > rune('Z') {
									goto l308
								}
								position++
								break
							default:
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l308
								}
								position++
								break
							}
						}

						goto l307
					l308:
						position, tokenIndex, depth = position308, tokenIndex308, depth308
					}
					depth--
					add(rulePegText, position306)
				}
				if !_rules[ruleAction15]() {
					goto l304
				}
				depth--
				add(rulename, position305)
			}
			return true
		l304:
			position, tokenIndex, depth = position304, tokenIndex304, depth304
			return false
		},
		/* 33 type <- <(sname / qname / ((&('*') pointer) | (&('[') array) | (&('M' | 'm') map)))> */
		func() bool {
			position311, tokenIndex311, depth311 := position, tokenIndex, depth
			{
				position312 := position
				depth++
				{
					position313, tokenIndex313, depth313 := position, tokenIndex, depth
					if !_rules[rulesname]() {
						goto l314
					}
					goto l313
				l314:
					position, tokenIndex, depth = position313, tokenIndex313, depth313
					if !_rules[ruleqname]() {
						goto l315
					}
					goto l313
				l315:
					position, tokenIndex, depth = position313, tokenIndex313, depth313
					{
						switch buffer[position] {
						case '*':
							if !_rules[rulepointer]() {
								goto l311
							}
							break
						case '[':
							if !_rules[rulearray]() {
								goto l311
							}
							break
						default:
							if !_rules[rulemap]() {
								goto l311
							}
							break
						}
					}

				}
			l313:
				depth--
				add(ruletype, position312)
			}
			return true
		l311:
			position, tokenIndex, depth = position311, tokenIndex311, depth311
			return false
		},
		/* 34 sname <- <(<((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))+> Action16)> */
		func() bool {
			position317, tokenIndex317, depth317 := position, tokenIndex, depth
			{
				position318 := position
				depth++
				{
					position319 := position
					depth++
					{
						switch buffer[position] {
						case '_':
							if buffer[position] != rune('_') {
								goto l317
							}
							position++
							break
						case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l317
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l317
							}
							position++
							break
						}
					}

				l320:
					{
						position321, tokenIndex321, depth321 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '_':
								if buffer[position] != rune('_') {
									goto l321
								}
								position++
								break
							case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
								if c := buffer[position]; c < rune('A') || c > rune('Z') {
									goto l321
								}
								position++
								break
							default:
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l321
								}
								position++
								break
							}
						}

						goto l320
					l321:
						position, tokenIndex, depth = position321, tokenIndex321, depth321
					}
					depth--
					add(rulePegText, position319)
				}
				if !_rules[ruleAction16]() {
					goto l317
				}
				depth--
				add(rulesname, position318)
			}
			return true
		l317:
			position, tokenIndex, depth = position317, tokenIndex317, depth317
			return false
		},
		/* 35 qname <- <(<(((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))+ '.' ((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))+)> Action17)> */
		func() bool {
			position324, tokenIndex324, depth324 := position, tokenIndex, depth
			{
				position325 := position
				depth++
				{
					position326 := position
					depth++
					{
						switch buffer[position] {
						case '_':
							if buffer[position] != rune('_') {
								goto l324
							}
							position++
							break
						case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l324
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l324
							}
							position++
							break
						}
					}

				l327:
					{
						position328, tokenIndex328, depth328 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '_':
								if buffer[position] != rune('_') {
									goto l328
								}
								position++
								break
							case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
								if c := buffer[position]; c < rune('A') || c > rune('Z') {
									goto l328
								}
								position++
								break
							default:
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l328
								}
								position++
								break
							}
						}

						goto l327
					l328:
						position, tokenIndex, depth = position328, tokenIndex328, depth328
					}
					if buffer[position] != rune('.') {
						goto l324
					}
					position++
					{
						switch buffer[position] {
						case '_':
							if buffer[position] != rune('_') {
								goto l324
							}
							position++
							break
						case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l324
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l324
							}
							position++
							break
						}
					}

				l331:
					{
						position332, tokenIndex332, depth332 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '_':
								if buffer[position] != rune('_') {
									goto l332
								}
								position++
								break
							case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
								if c := buffer[position]; c < rune('A') || c > rune('Z') {
									goto l332
								}
								position++
								break
							default:
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l332
								}
								position++
								break
							}
						}

						goto l331
					l332:
						position, tokenIndex, depth = position332, tokenIndex332, depth332
					}
					depth--
					add(rulePegText, position326)
				}
				if !_rules[ruleAction17]() {
					goto l324
				}
				depth--
				add(ruleqname, position325)
			}
			return true
		l324:
			position, tokenIndex, depth = position324, tokenIndex324, depth324
			return false
		},
		/* 36 array <- <('[' ']' type Action18)> */
		func() bool {
			position335, tokenIndex335, depth335 := position, tokenIndex, depth
			{
				position336 := position
				depth++
				if buffer[position] != rune('[') {
					goto l335
				}
				position++
				if buffer[position] != rune(']') {
					goto l335
				}
				position++
				if !_rules[ruletype]() {
					goto l335
				}
				if !_rules[ruleAction18]() {
					goto l335
				}
				depth--
				add(rulearray, position336)
			}
			return true
		l335:
			position, tokenIndex, depth = position335, tokenIndex335, depth335
			return false
		},
		/* 37 map <- <(('m' / 'M') ('a' / 'A') ('p' / 'P') '[' isp* keytype isp* ']' type Action19)> */
		func() bool {
			position337, tokenIndex337, depth337 := position, tokenIndex, depth
			{
				position338 := position
				depth++
				{
					position339, tokenIndex339, depth339 := position, tokenIndex, depth
					if buffer[position] != rune('m') {
						goto l340
					}
					position++
					goto l339
				l340:
					position, tokenIndex, depth = position339, tokenIndex339, depth339
					if buffer[position] != rune('M') {
						goto l337
					}
					position++
				}
			l339:
				{
					position341, tokenIndex341, depth341 := position, tokenIndex, depth
					if buffer[position] != rune('a') {
						goto l342
					}
					position++
					goto l341
				l342:
					position, tokenIndex, depth = position341, tokenIndex341, depth341
					if buffer[position] != rune('A') {
						goto l337
					}
					position++
				}
			l341:
				{
					position343, tokenIndex343, depth343 := position, tokenIndex, depth
					if buffer[position] != rune('p') {
						goto l344
					}
					position++
					goto l343
				l344:
					position, tokenIndex, depth = position343, tokenIndex343, depth343
					if buffer[position] != rune('P') {
						goto l337
					}
					position++
				}
			l343:
				if buffer[position] != rune('[') {
					goto l337
				}
				position++
			l345:
				{
					position346, tokenIndex346, depth346 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l346
					}
					goto l345
				l346:
					position, tokenIndex, depth = position346, tokenIndex346, depth346
				}
				if !_rules[rulekeytype]() {
					goto l337
				}
			l347:
				{
					position348, tokenIndex348, depth348 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l348
					}
					goto l347
				l348:
					position, tokenIndex, depth = position348, tokenIndex348, depth348
				}
				if buffer[position] != rune(']') {
					goto l337
				}
				position++
				if !_rules[ruletype]() {
					goto l337
				}
				if !_rules[ruleAction19]() {
					goto l337
				}
				depth--
				add(rulemap, position338)
			}
			return true
		l337:
			position, tokenIndex, depth = position337, tokenIndex337, depth337
			return false
		},
		/* 38 keytype <- <(type Action20)> */
		func() bool {
			position349, tokenIndex349, depth349 := position, tokenIndex, depth
			{
				position350 := position
				depth++
				if !_rules[ruletype]() {
					goto l349
				}
				if !_rules[ruleAction20]() {
					goto l349
				}
				depth--
				add(rulekeytype, position350)
			}
			return true
		l349:
			position, tokenIndex, depth = position349, tokenIndex349, depth349
			return false
		},
		/* 39 pointer <- <('*' type Action21)> */
		func() bool {
			position351, tokenIndex351, depth351 := position, tokenIndex, depth
			{
				position352 := position
				depth++
				if buffer[position] != rune('*') {
					goto l351
				}
				position++
				if !_rules[ruletype]() {
					goto l351
				}
				if !_rules[ruleAction21]() {
					goto l351
				}
				depth--
				add(rulepointer, position352)
			}
			return true
		l351:
			position, tokenIndex, depth = position351, tokenIndex351, depth351
			return false
		},
		/* 40 captures <- <(isp* capture isp* (',' isp* capture isp*)* !.)> */
		func() bool {
			position353, tokenIndex353, depth353 := position, tokenIndex, depth
			{
				position354 := position
				depth++
			l355:
				{
					position356, tokenIndex356, depth356 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l356
					}
					goto l355
				l356:
					position, tokenIndex, depth = position356, tokenIndex356, depth356
				}
				if !_rules[rulecapture]() {
					goto l353
				}
			l357:
				{
					position358, tokenIndex358, depth358 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l358
					}
					goto l357
				l358:
					position, tokenIndex, depth = position358, tokenIndex358, depth358
				}
			l359:
				{
					position360, tokenIndex360, depth360 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l360
					}
					position++
				l361:
					{
						position362, tokenIndex362, depth362 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l362
						}
						goto l361
					l362:
						position, tokenIndex, depth = position362, tokenIndex362, depth362
					}
					if !_rules[rulecapture]() {
						goto l360
					}
				l363:
					{
						position364, tokenIndex364, depth364 := position, tokenIndex, depth
						if !_rules[ruleisp]() {
							goto l364
						}
						goto l363
					l364:
						position, tokenIndex, depth = position364, tokenIndex364, depth364
					}
					goto l359
				l360:
					position, tokenIndex, depth = position360, tokenIndex360, depth360
				}
				{
					position365, tokenIndex365, depth365 := position, tokenIndex, depth
					if !matchDot() {
						goto l365
					}
					goto l353
				l365:
					position, tokenIndex, depth = position365, tokenIndex365, depth365
				}
				depth--
				add(rulecaptures, position354)
			}
			return true
		l353:
			position, tokenIndex, depth = position353, tokenIndex353, depth353
			return false
		},
		/* 41 capture <- <(eventid isp* ':' handlername isp* mappings isp* tags Action22)> */
		func() bool {
			position366, tokenIndex366, depth366 := position, tokenIndex, depth
			{
				position367 := position
				depth++
				if !_rules[ruleeventid]() {
					goto l366
				}
			l368:
				{
					position369, tokenIndex369, depth369 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l369
					}
					goto l368
				l369:
					position, tokenIndex, depth = position369, tokenIndex369, depth369
				}
				if buffer[position] != rune(':') {
					goto l366
				}
				position++
				if !_rules[rulehandlername]() {
					goto l366
				}
			l370:
				{
					position371, tokenIndex371, depth371 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l371
					}
					goto l370
				l371:
					position, tokenIndex, depth = position371, tokenIndex371, depth371
				}
				if !_rules[rulemappings]() {
					goto l366
				}
			l372:
				{
					position373, tokenIndex373, depth373 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l373
					}
					goto l372
				l373:
					position, tokenIndex, depth = position373, tokenIndex373, depth373
				}
				if !_rules[ruletags]() {
					goto l366
				}
				if !_rules[ruleAction22]() {
					goto l366
				}
				depth--
				add(rulecapture, position367)
			}
			return true
		l366:
			position, tokenIndex, depth = position366, tokenIndex366, depth366
			return false
		},
		/* 42 handlername <- <(<identifier> Action23)> */
		func() bool {
			position374, tokenIndex374, depth374 := position, tokenIndex, depth
			{
				position375 := position
				depth++
				{
					position376 := position
					depth++
					if !_rules[ruleidentifier]() {
						goto l374
					}
					depth--
					add(rulePegText, position376)
				}
				if !_rules[ruleAction23]() {
					goto l374
				}
				depth--
				add(rulehandlername, position375)
			}
			return true
		l374:
			position, tokenIndex, depth = position374, tokenIndex374, depth374
			return false
		},
		/* 43 eventid <- <(<[a-z]+> Action24)> */
		func() bool {
			position377, tokenIndex377, depth377 := position, tokenIndex, depth
			{
				position378 := position
				depth++
				{
					position379 := position
					depth++
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l377
					}
					position++
				l380:
					{
						position381, tokenIndex381, depth381 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l381
						}
						position++
						goto l380
					l381:
						position, tokenIndex, depth = position381, tokenIndex381, depth381
					}
					depth--
					add(rulePegText, position379)
				}
				if !_rules[ruleAction24]() {
					goto l377
				}
				depth--
				add(ruleeventid, position378)
			}
			return true
		l377:
			position, tokenIndex, depth = position377, tokenIndex377, depth377
			return false
		},
		/* 44 mappings <- <('(' (isp* mapping isp* (',' isp* mapping isp*)*)? ')')?> */
		func() bool {
			{
				position383 := position
				depth++
				{
					position384, tokenIndex384, depth384 := position, tokenIndex, depth
					if buffer[position] != rune('(') {
						goto l384
					}
					position++
					{
						position386, tokenIndex386, depth386 := position, tokenIndex, depth
					l388:
						{
							position389, tokenIndex389, depth389 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l389
							}
							goto l388
						l389:
							position, tokenIndex, depth = position389, tokenIndex389, depth389
						}
						if !_rules[rulemapping]() {
							goto l386
						}
					l390:
						{
							position391, tokenIndex391, depth391 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l391
							}
							goto l390
						l391:
							position, tokenIndex, depth = position391, tokenIndex391, depth391
						}
					l392:
						{
							position393, tokenIndex393, depth393 := position, tokenIndex, depth
							if buffer[position] != rune(',') {
								goto l393
							}
							position++
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
							if !_rules[rulemapping]() {
								goto l393
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
							goto l392
						l393:
							position, tokenIndex, depth = position393, tokenIndex393, depth393
						}
						goto l387
					l386:
						position, tokenIndex, depth = position386, tokenIndex386, depth386
					}
				l387:
					if buffer[position] != rune(')') {
						goto l384
					}
					position++
					goto l385
				l384:
					position, tokenIndex, depth = position384, tokenIndex384, depth384
				}
			l385:
				depth--
				add(rulemappings, position383)
			}
			return true
		},
		/* 45 mapping <- <(mappingname isp* '=' isp* bound Action25)> */
		func() bool {
			position398, tokenIndex398, depth398 := position, tokenIndex, depth
			{
				position399 := position
				depth++
				if !_rules[rulemappingname]() {
					goto l398
				}
			l400:
				{
					position401, tokenIndex401, depth401 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l401
					}
					goto l400
				l401:
					position, tokenIndex, depth = position401, tokenIndex401, depth401
				}
				if buffer[position] != rune('=') {
					goto l398
				}
				position++
			l402:
				{
					position403, tokenIndex403, depth403 := position, tokenIndex, depth
					if !_rules[ruleisp]() {
						goto l403
					}
					goto l402
				l403:
					position, tokenIndex, depth = position403, tokenIndex403, depth403
				}
				if !_rules[rulebound]() {
					goto l398
				}
				if !_rules[ruleAction25]() {
					goto l398
				}
				depth--
				add(rulemapping, position399)
			}
			return true
		l398:
			position, tokenIndex, depth = position398, tokenIndex398, depth398
			return false
		},
		/* 46 mappingname <- <(<identifier> Action26)> */
		func() bool {
			position404, tokenIndex404, depth404 := position, tokenIndex, depth
			{
				position405 := position
				depth++
				{
					position406 := position
					depth++
					if !_rules[ruleidentifier]() {
						goto l404
					}
					depth--
					add(rulePegText, position406)
				}
				if !_rules[ruleAction26]() {
					goto l404
				}
				depth--
				add(rulemappingname, position405)
			}
			return true
		l404:
			position, tokenIndex, depth = position404, tokenIndex404, depth404
			return false
		},
		/* 47 tags <- <('{' isp* tag isp* (',' isp* tag isp*)* '}')?> */
		func() bool {
			{
				position408 := position
				depth++
				{
					position409, tokenIndex409, depth409 := position, tokenIndex, depth
					if buffer[position] != rune('{') {
						goto l409
					}
					position++
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
					if !_rules[ruletag]() {
						goto l409
					}
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
				l415:
					{
						position416, tokenIndex416, depth416 := position, tokenIndex, depth
						if buffer[position] != rune(',') {
							goto l416
						}
						position++
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
						if !_rules[ruletag]() {
							goto l416
						}
					l419:
						{
							position420, tokenIndex420, depth420 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l420
							}
							goto l419
						l420:
							position, tokenIndex, depth = position420, tokenIndex420, depth420
						}
						goto l415
					l416:
						position, tokenIndex, depth = position416, tokenIndex416, depth416
					}
					if buffer[position] != rune('}') {
						goto l409
					}
					position++
					goto l410
				l409:
					position, tokenIndex, depth = position409, tokenIndex409, depth409
				}
			l410:
				depth--
				add(ruletags, position408)
			}
			return true
		},
		/* 48 tag <- <(tagname ('(' (isp* tagarg isp* (',' isp* tagarg isp*)*)? ')')? Action27)> */
		func() bool {
			position421, tokenIndex421, depth421 := position, tokenIndex, depth
			{
				position422 := position
				depth++
				if !_rules[ruletagname]() {
					goto l421
				}
				{
					position423, tokenIndex423, depth423 := position, tokenIndex, depth
					if buffer[position] != rune('(') {
						goto l423
					}
					position++
					{
						position425, tokenIndex425, depth425 := position, tokenIndex, depth
					l427:
						{
							position428, tokenIndex428, depth428 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l428
							}
							goto l427
						l428:
							position, tokenIndex, depth = position428, tokenIndex428, depth428
						}
						if !_rules[ruletagarg]() {
							goto l425
						}
					l429:
						{
							position430, tokenIndex430, depth430 := position, tokenIndex, depth
							if !_rules[ruleisp]() {
								goto l430
							}
							goto l429
						l430:
							position, tokenIndex, depth = position430, tokenIndex430, depth430
						}
					l431:
						{
							position432, tokenIndex432, depth432 := position, tokenIndex, depth
							if buffer[position] != rune(',') {
								goto l432
							}
							position++
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
							if !_rules[ruletagarg]() {
								goto l432
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
							goto l431
						l432:
							position, tokenIndex, depth = position432, tokenIndex432, depth432
						}
						goto l426
					l425:
						position, tokenIndex, depth = position425, tokenIndex425, depth425
					}
				l426:
					if buffer[position] != rune(')') {
						goto l423
					}
					position++
					goto l424
				l423:
					position, tokenIndex, depth = position423, tokenIndex423, depth423
				}
			l424:
				if !_rules[ruleAction27]() {
					goto l421
				}
				depth--
				add(ruletag, position422)
			}
			return true
		l421:
			position, tokenIndex, depth = position421, tokenIndex421, depth421
			return false
		},
		/* 49 tagname <- <(<identifier> Action28)> */
		func() bool {
			position437, tokenIndex437, depth437 := position, tokenIndex, depth
			{
				position438 := position
				depth++
				{
					position439 := position
					depth++
					if !_rules[ruleidentifier]() {
						goto l437
					}
					depth--
					add(rulePegText, position439)
				}
				if !_rules[ruleAction28]() {
					goto l437
				}
				depth--
				add(ruletagname, position438)
			}
			return true
		l437:
			position, tokenIndex, depth = position437, tokenIndex437, depth437
			return false
		},
		/* 50 tagarg <- <(<identifier> Action29)> */
		func() bool {
			position440, tokenIndex440, depth440 := position, tokenIndex, depth
			{
				position441 := position
				depth++
				{
					position442 := position
					depth++
					if !_rules[ruleidentifier]() {
						goto l440
					}
					depth--
					add(rulePegText, position442)
				}
				if !_rules[ruleAction29]() {
					goto l440
				}
				depth--
				add(ruletagarg, position441)
			}
			return true
		l440:
			position, tokenIndex, depth = position440, tokenIndex440, depth440
			return false
		},
		/* 52 Action0 <- <{
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
		/* 54 Action1 <- <{
			p.goVal.Name = buffer[begin:end]
		}> */
		func() bool {
			{
				add(ruleAction1, position)
			}
			return true
		},
		/* 55 Action2 <- <{
			p.goVal.Type = p.valuetype
			p.valuetype = nil
		}> */
		func() bool {
			{
				add(ruleAction2, position)
			}
			return true
		},
		/* 56 Action3 <- <{
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
		/* 57 Action4 <- <{
			p.bv.Kind = data.BoundSelf
		}> */
		func() bool {
			{
				add(ruleAction4, position)
			}
			return true
		},
		/* 58 Action5 <- <{
			p.bv.Kind = data.BoundData
		}> */
		func() bool {
			{
				add(ruleAction5, position)
			}
			return true
		},
		/* 59 Action6 <- <{
			p.bv.Kind = data.BoundProperty
		}> */
		func() bool {
			{
				add(ruleAction6, position)
			}
			return true
		},
		/* 60 Action7 <- <{
			p.bv.Kind = data.BoundStyle
		}> */
		func() bool {
			{
				add(ruleAction7, position)
			}
			return true
		},
		/* 61 Action8 <- <{
			p.bv.Kind = data.BoundClass
		}> */
		func() bool {
			{
				add(ruleAction8, position)
			}
			return true
		},
		/* 62 Action9 <- <{
			p.bv.Kind = data.BoundFormValue
		}> */
		func() bool {
			{
				add(ruleAction9, position)
			}
			return true
		},
		/* 63 Action10 <- <{
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
		/* 64 Action11 <- <{
			p.bv.IDs = append(p.bv.IDs, buffer[begin:end])
		}> */
		func() bool {
			{
				add(ruleAction11, position)
			}
			return true
		},
		/* 65 Action12 <- <{
			p.bv.IDs = append(p.bv.IDs, buffer[begin:end])
		}> */
		func() bool {
			{
				add(ruleAction12, position)
			}
			return true
		},
		/* 66 Action13 <- <{
			p.expr = buffer[begin:end]
		}> */
		func() bool {
			{
				add(ruleAction13, position)
			}
			return true
		},
		/* 67 Action14 <- <{
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
		/* 68 Action15 <- <{
			p.names = append(p.names, buffer[begin:end])
		}> */
		func() bool {
			{
				add(ruleAction15, position)
			}
			return true
		},
		/* 69 Action16 <- <{
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
		/* 70 Action17 <- <{
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
		/* 71 Action18 <- <{
			p.valuetype = &data.ParamType{Kind: data.ArrayType, ValueType: p.valuetype}
		}> */
		func() bool {
			{
				add(ruleAction18, position)
			}
			return true
		},
		/* 72 Action19 <- <{
			p.valuetype = &data.ParamType{Kind: data.MapType, KeyType: p.keytype, ValueType: p.valuetype}
		}> */
		func() bool {
			{
				add(ruleAction19, position)
			}
			return true
		},
		/* 73 Action20 <- <{
			p.keytype = p.valuetype
		}> */
		func() bool {
			{
				add(ruleAction20, position)
			}
			return true
		},
		/* 74 Action21 <- <{
			p.valuetype = &data.ParamType{Kind: data.PointerType, ValueType: p.valuetype}
		}> */
		func() bool {
			{
				add(ruleAction21, position)
			}
			return true
		},
		/* 75 Action22 <- <{
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
		/* 76 Action23 <- <{
			p.handlername = buffer[begin:end]
		}> */
		func() bool {
			{
				add(ruleAction23, position)
			}
			return true
		},
		/* 77 Action24 <- <{
			p.expr = buffer[begin:end]
		}> */
		func() bool {
			{
				add(ruleAction24, position)
			}
			return true
		},
		/* 78 Action25 <- <{
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
		/* 79 Action26 <- <{
			p.tagname = buffer[begin:end]
		}> */
		func() bool {
			{
				add(ruleAction26, position)
			}
			return true
		},
		/* 80 Action27 <- <{
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
		/* 81 Action28 <- <{
			p.tagname = buffer[begin:end]
		}> */
		func() bool {
			{
				add(ruleAction28, position)
			}
			return true
		},
		/* 82 Action29 <- <{
			p.names = append(p.names, buffer[begin:end])
		}> */
		func() bool {
			{
				add(ruleAction29, position)
			}
			return true
		},
	}
	p.rules = _rules
}
