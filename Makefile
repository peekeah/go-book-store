build:
	@go build -o bin/book-store

run: build
	@./bin/book-store

test:
	go test -v
