
.PHONY: build

build:
	@mkdir -p tmp
	@go build -o tmp/server .

run: build
	@./tmp/server
