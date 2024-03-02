.PHONY: test
test:
	go test -race ./...
.DEFAULT_TARGET: test