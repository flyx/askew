all: askew
testjs: run-askew-js test/site/main.js
testwasm: run-askew-wasm test/site/main.wasm test/site/wasm_exec.js

askew:
	go build

run-askew-js: askew test/site
	./askew -o test/site test

run-askew-wasm: askew test/site
	./askew -b wasm -o test/site test

.PHONY: askew run-askew-js run-askew-wasm testjs testwasm test/site/main.js test/site/main.wasm

test/site:
	mkdir -p test/site

test/site/main.js: export GOPHERJS_GOROOT = $(shell go1.12.16 env GOROOT)
test/site/main.js: test/site
	cd test && gopherjs build -o site/main.js

test/site/main.wasm: export GOOS = js
test/site/main.wasm: export GOARCH = wasm
test/site/main.wasm: test/site
	cd test && go build -o site/main.wasm

test/site/wasm_exec.js:
	cp $(shell go env GOROOT)/misc/wasm/wasm_exec.js $@