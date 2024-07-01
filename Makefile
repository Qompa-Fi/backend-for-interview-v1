
.PHONY: build

build:
	@mkdir -p build
	@go build -o build/server .

run: build
	@./build/server
