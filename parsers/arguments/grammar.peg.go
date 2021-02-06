package arguments

import (
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
	rulearg
	rulecontent
	ruletext
	rulequoted
	ruleenclosed
	ruleinner
	ruleAction0

	rulePre
	ruleIn
	ruleSuf
)

var rul3s = [...]string{
	"Unknown",
	"e",
	"arg",
	"content",
	"text",
	"quoted",
	"enclosed",
	"inner",
	"Action0",

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

type ArgumentParser struct {
	count int

	Buffer string
	buffer []rune
	rules  [9]func() bool
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
	p   *ArgumentParser
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

func (p *ArgumentParser) PrintSyntaxTree() {
	p.tokens32.PrintSyntaxTree(p.Buffer)
}

func (p *ArgumentParser) Highlighter() {
	p.PrintSyntax()
}

func (p *ArgumentParser) Execute() {
	buffer, _buffer, text, begin, end := p.Buffer, p.buffer, "", 0, 0
	for token := range p.Tokens() {
		switch token.pegRule {

		case ruleAction0:

			p.count++

		}
	}
	_, _, _, _, _ = buffer, _buffer, text, begin, end
}

func (p *ArgumentParser) Init() {
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

	_rules = [...]func() bool{
		nil,
		/* 0 e <- <((arg (',' arg)*)? !.)> */
		func() bool {
			position0, tokenIndex0, depth0 := position, tokenIndex, depth
			{
				position1 := position
				depth++
				{
					position2, tokenIndex2, depth2 := position, tokenIndex, depth
					if !_rules[rulearg]() {
						goto l2
					}
				l4:
					{
						position5, tokenIndex5, depth5 := position, tokenIndex, depth
						if buffer[position] != rune(',') {
							goto l5
						}
						position++
						if !_rules[rulearg]() {
							goto l5
						}
						goto l4
					l5:
						position, tokenIndex, depth = position5, tokenIndex5, depth5
					}
					goto l3
				l2:
					position, tokenIndex, depth = position2, tokenIndex2, depth2
				}
			l3:
				{
					position6, tokenIndex6, depth6 := position, tokenIndex, depth
					if !matchDot() {
						goto l6
					}
					goto l0
				l6:
					position, tokenIndex, depth = position6, tokenIndex6, depth6
				}
				depth--
				add(rulee, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 arg <- <(content Action0)> */
		func() bool {
			position7, tokenIndex7, depth7 := position, tokenIndex, depth
			{
				position8 := position
				depth++
				if !_rules[rulecontent]() {
					goto l7
				}
				if !_rules[ruleAction0]() {
					goto l7
				}
				depth--
				add(rulearg, position8)
			}
			return true
		l7:
			position, tokenIndex, depth = position7, tokenIndex7, depth7
			return false
		},
		/* 2 content <- <(text / quoted / enclosed)+> */
		func() bool {
			position9, tokenIndex9, depth9 := position, tokenIndex, depth
			{
				position10 := position
				depth++
				{
					position13, tokenIndex13, depth13 := position, tokenIndex, depth
					if !_rules[ruletext]() {
						goto l14
					}
					goto l13
				l14:
					position, tokenIndex, depth = position13, tokenIndex13, depth13
					if !_rules[rulequoted]() {
						goto l15
					}
					goto l13
				l15:
					position, tokenIndex, depth = position13, tokenIndex13, depth13
					if !_rules[ruleenclosed]() {
						goto l9
					}
				}
			l13:
			l11:
				{
					position12, tokenIndex12, depth12 := position, tokenIndex, depth
					{
						position16, tokenIndex16, depth16 := position, tokenIndex, depth
						if !_rules[ruletext]() {
							goto l17
						}
						goto l16
					l17:
						position, tokenIndex, depth = position16, tokenIndex16, depth16
						if !_rules[rulequoted]() {
							goto l18
						}
						goto l16
					l18:
						position, tokenIndex, depth = position16, tokenIndex16, depth16
						if !_rules[ruleenclosed]() {
							goto l12
						}
					}
				l16:
					goto l11
				l12:
					position, tokenIndex, depth = position12, tokenIndex12, depth12
				}
				depth--
				add(rulecontent, position10)
			}
			return true
		l9:
			position, tokenIndex, depth = position9, tokenIndex9, depth9
			return false
		},
		/* 3 text <- <(!((&('\'') '\'') | (&('"') '"') | (&('}') '}') | (&('{') '{') | (&(')') ')') | (&('(') '(') | (&(',') ',')) .)+> */
		func() bool {
			position19, tokenIndex19, depth19 := position, tokenIndex, depth
			{
				position20 := position
				depth++
				{
					position23, tokenIndex23, depth23 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '\'':
							if buffer[position] != rune('\'') {
								goto l23
							}
							position++
							break
						case '"':
							if buffer[position] != rune('"') {
								goto l23
							}
							position++
							break
						case '}':
							if buffer[position] != rune('}') {
								goto l23
							}
							position++
							break
						case '{':
							if buffer[position] != rune('{') {
								goto l23
							}
							position++
							break
						case ')':
							if buffer[position] != rune(')') {
								goto l23
							}
							position++
							break
						case '(':
							if buffer[position] != rune('(') {
								goto l23
							}
							position++
							break
						default:
							if buffer[position] != rune(',') {
								goto l23
							}
							position++
							break
						}
					}

					goto l19
				l23:
					position, tokenIndex, depth = position23, tokenIndex23, depth23
				}
				if !matchDot() {
					goto l19
				}
			l21:
				{
					position22, tokenIndex22, depth22 := position, tokenIndex, depth
					{
						position25, tokenIndex25, depth25 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '\'':
								if buffer[position] != rune('\'') {
									goto l25
								}
								position++
								break
							case '"':
								if buffer[position] != rune('"') {
									goto l25
								}
								position++
								break
							case '}':
								if buffer[position] != rune('}') {
									goto l25
								}
								position++
								break
							case '{':
								if buffer[position] != rune('{') {
									goto l25
								}
								position++
								break
							case ')':
								if buffer[position] != rune(')') {
									goto l25
								}
								position++
								break
							case '(':
								if buffer[position] != rune('(') {
									goto l25
								}
								position++
								break
							default:
								if buffer[position] != rune(',') {
									goto l25
								}
								position++
								break
							}
						}

						goto l22
					l25:
						position, tokenIndex, depth = position25, tokenIndex25, depth25
					}
					if !matchDot() {
						goto l22
					}
					goto l21
				l22:
					position, tokenIndex, depth = position22, tokenIndex22, depth22
				}
				depth--
				add(ruletext, position20)
			}
			return true
		l19:
			position, tokenIndex, depth = position19, tokenIndex19, depth19
			return false
		},
		/* 4 quoted <- <((&('`') ('`' (!'`' .)* '`')) | (&('\'') ('\'' . '\'')) | (&('"') ('"' ((!('"' / '\n') .) / '"')* '"')))> */
		func() bool {
			position27, tokenIndex27, depth27 := position, tokenIndex, depth
			{
				position28 := position
				depth++
				{
					switch buffer[position] {
					case '`':
						if buffer[position] != rune('`') {
							goto l27
						}
						position++
					l30:
						{
							position31, tokenIndex31, depth31 := position, tokenIndex, depth
							{
								position32, tokenIndex32, depth32 := position, tokenIndex, depth
								if buffer[position] != rune('`') {
									goto l32
								}
								position++
								goto l31
							l32:
								position, tokenIndex, depth = position32, tokenIndex32, depth32
							}
							if !matchDot() {
								goto l31
							}
							goto l30
						l31:
							position, tokenIndex, depth = position31, tokenIndex31, depth31
						}
						if buffer[position] != rune('`') {
							goto l27
						}
						position++
						break
					case '\'':
						if buffer[position] != rune('\'') {
							goto l27
						}
						position++
						if !matchDot() {
							goto l27
						}
						if buffer[position] != rune('\'') {
							goto l27
						}
						position++
						break
					default:
						if buffer[position] != rune('"') {
							goto l27
						}
						position++
					l33:
						{
							position34, tokenIndex34, depth34 := position, tokenIndex, depth
							{
								position35, tokenIndex35, depth35 := position, tokenIndex, depth
								{
									position37, tokenIndex37, depth37 := position, tokenIndex, depth
									{
										position38, tokenIndex38, depth38 := position, tokenIndex, depth
										if buffer[position] != rune('"') {
											goto l39
										}
										position++
										goto l38
									l39:
										position, tokenIndex, depth = position38, tokenIndex38, depth38
										if buffer[position] != rune('\n') {
											goto l37
										}
										position++
									}
								l38:
									goto l36
								l37:
									position, tokenIndex, depth = position37, tokenIndex37, depth37
								}
								if !matchDot() {
									goto l36
								}
								goto l35
							l36:
								position, tokenIndex, depth = position35, tokenIndex35, depth35
								if buffer[position] != rune('"') {
									goto l34
								}
								position++
							}
						l35:
							goto l33
						l34:
							position, tokenIndex, depth = position34, tokenIndex34, depth34
						}
						if buffer[position] != rune('"') {
							goto l27
						}
						position++
						break
					}
				}

				depth--
				add(rulequoted, position28)
			}
			return true
		l27:
			position, tokenIndex, depth = position27, tokenIndex27, depth27
			return false
		},
		/* 5 enclosed <- <(('{' inner '}') / ('(' inner ')'))> */
		func() bool {
			position40, tokenIndex40, depth40 := position, tokenIndex, depth
			{
				position41 := position
				depth++
				{
					position42, tokenIndex42, depth42 := position, tokenIndex, depth
					if buffer[position] != rune('{') {
						goto l43
					}
					position++
					if !_rules[ruleinner]() {
						goto l43
					}
					if buffer[position] != rune('}') {
						goto l43
					}
					position++
					goto l42
				l43:
					position, tokenIndex, depth = position42, tokenIndex42, depth42
					if buffer[position] != rune('(') {
						goto l40
					}
					position++
					if !_rules[ruleinner]() {
						goto l40
					}
					if buffer[position] != rune(')') {
						goto l40
					}
					position++
				}
			l42:
				depth--
				add(ruleenclosed, position41)
			}
			return true
		l40:
			position, tokenIndex, depth = position40, tokenIndex40, depth40
			return false
		},
		/* 6 inner <- <(content / ',')*> */
		func() bool {
			{
				position45 := position
				depth++
			l46:
				{
					position47, tokenIndex47, depth47 := position, tokenIndex, depth
					{
						position48, tokenIndex48, depth48 := position, tokenIndex, depth
						if !_rules[rulecontent]() {
							goto l49
						}
						goto l48
					l49:
						position, tokenIndex, depth = position48, tokenIndex48, depth48
						if buffer[position] != rune(',') {
							goto l47
						}
						position++
					}
				l48:
					goto l46
				l47:
					position, tokenIndex, depth = position47, tokenIndex47, depth47
				}
				depth--
				add(ruleinner, position45)
			}
			return true
		},
		/* 8 Action0 <- <{
		  p.count++
		}> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
	}
	p.rules = _rules
}
