.PHONY: test vet fmt

test:
	go test -v ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

lint:
	golangci-lint run

check: test vet fmt