.PHONY: test
test:
	go test -race ./.../handler

.PHONY: build
build:
	go build -o go-http cmd/main.go
.DEFAULT_TARGET: test