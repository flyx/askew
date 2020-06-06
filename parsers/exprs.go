package parsers

var exprSyntax = `
  EXPR      ← (COMMALESS / ENCLOSING)+
	COMMALESS ← IDENTIFIER / OPERATORS / STRING
	OPERATORS ← < [+-*/|&^:=.]+ >
	STRING    ← '` + "`" + `' (!'` + "`" + `' .)* '` + "`" + `'
	ENCLOSING ← PARENS / BRACES / BRACKETS
	PARENS    ← '(' ENCLOSED ')'
	BRACES    ← '{' ENCLOSED '}'
	BRACKETS  ← '[' ENCLOSED ']'
	ENCLOSED  ← (COMMALESS / ENCLOSING / ',')*
` + identifierSyntax + whitespace
