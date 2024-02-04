build:
	go build -o bin/executor main.go

test:
	go test -timeout 30s -v ./...
