all:
	go install ./...

clean:
	go clean

test:
	go test ./... -v

lint:
	golangci-lint run
