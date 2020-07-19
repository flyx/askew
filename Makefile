all: askew
test: test/site/index.html test/site/main.js

askew:
	go build

run-askew: askew test/site
	./askew -o test/site test

.PHONY: askew test-run-askew test

test/ui/ui.go: run-askew
test/extra/additionals.go: run-askew
test/site/index.html: run-askew

test/site:
	mkdir -p test/site

test/site/main.js: export GOPHERJS_GOROOT = $(shell go1.12.16 env GOROOT)
test/site/main.js: test/site test/ui/ui.go test/ui/controllers.go test/extra/additionals.go test/extra/init.go
	cd test && gopherjs build -o site/main.js