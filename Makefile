all: build

build: cli/main.go server/main.go
	@go build -o bin/cli cli/main.go 
	@go build -o bin/server server/main.go 

