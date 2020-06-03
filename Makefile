all: askew
test: test/site/index.html test/site/main.js

askew:
	go build

run-askew: askew test/site
	./askew -o test -s test/skeleton.html -i test/site/index.html test/components.html

.PHONY: askew test-run-askew test

test/site/index.html: run-askew
test/ui/nameform.go: run-askew
test/ui/nameforms.go: run-askew
test/ui/macrotest.go: run-askew

test/site:
	mkdir -p test/site

test/site/main.js: export GOPHERJS_GOROOT = $(shell go1.12.16 env GOROOT)
test/site/main.js: test/site test/ui/nameform.go test/ui/nameforms.go test/ui/macrotest.go test/ui/macrotestcontrol.go
	cd test && gopherjs build -o site/main.js