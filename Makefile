all: generator/tbc
test: test/site/index.html

test/generated:
	mkdir test/generated

generator/tbc:
	cd generator && go build -o tbc

run-tbc: test/generated generator/tbc
	generator/tbc -o test/generated/templates.html -p test/generated/ui test/templates.html

.PHONY: generator/tbc test-run-tbc test

test/generated/templates.html: run-tbc
test/generated/ui/nameform.go: run-tbc

test/site:
	mkdir site

test/site/main.js: test/templates.html test/site test/generated/templates.html test/generated/ui/nameform.go
	cd test/main && gopherjs build -o ../site/main.js

test/site/index.html: test/site/main.js
	cat test/index-top.html > test/site/index.html
	cat test/generated/templates.html >> test/site/index.html
	cat test/index-bottom.html >> test/site/index.html