.PHONY: test
test:
	go test -race ./...

.PHONY: build
build:
	go build -o go-http cmd/main.go
.DEFAULT_TARGET: test