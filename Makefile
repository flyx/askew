all: tbc
test: test/site/index.html test/site/main.js

tbc:
	go build

run-tbc: tbc test/site
	./tbc -o test -s test/skeleton.html -i test/site/index.html test/components.html

.PHONY: tbc test-run-tbc test

test/site/index.html: run-tbc
test/ui/nameform.go: run-tbc
test/ui/nameforms.go: run-tbc
test/ui/macrotest.go: run-tbc

test/site:
	mkdir -p test/site

test/site/main.js: export GOPHERJS_GOROOT = $(shell go1.12.16 env GOROOT)
test/site/main.js: test/site test/ui/nameform.go test/ui/nameforms.go test/ui/macrotest.go test/ui/macrotestcontrol.go
	cd test && gopherjs build -o site/main.js