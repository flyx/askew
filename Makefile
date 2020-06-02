all: tbc
test: test/site/index.html

test/generated:
	mkdir test/generated

tbc:
	go build

run-tbc: test/generated tbc
	./tbc -o test/generated test/components.html

.PHONY: tbc test-run-tbc test

test/generated/templates.html: run-tbc
test/generated/ui/*.go: run-tbc

test/site:
	mkdir -p test/site

test/site/main.js: export GOPHERJS_GOROOT = $(shell go1.12.16 env GOROOT)
test/site/main.js: test/site test/generated/templates.html test/generated/ui/nameform.go
	cd test/main && gopherjs build -o ../site/main.js

test/site/index.html: test/site/main.js
	cat test/index-top.html > test/site/index.html
	cat test/generated/templates.html >> test/site/index.html
	cat test/index-bottom.html >> test/site/index.html