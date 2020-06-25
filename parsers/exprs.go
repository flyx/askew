package parsers

var exprSyntax = `
  EXPR      ← (COMMALESS / ENCLOSING)+
	COMMALESS ← IDENTIFIER / NUMBER / OPERATORS / STRING
	NUMBER    ← < [0-9]+ >
	OPERATORS ← < [+-*/|&^:=.]+ >
	STRING    ← '` + "`" + `' (!'` + "`" + `' .)* '` + "`" + `'
	ENCLOSING ← PARENS / BRACES / BRACKETS
	PARENS    ← '(' ENCLOSED ')'
	BRACES    ← '{' ENCLOSED '}'
	BRACKETS  ← '[' ENCLOSED ']'
	ENCLOSED  ← (COMMALESS / ENCLOSING / ',')*
` + identifierSyntax + whitespace
