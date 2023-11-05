BINARY_NAME=lego

build:
	GOARCH=amd64 GOOS=darwin go build -o ./bin/${BINARY_NAME}-darwin *.go
	GOARCH=amd64 GOOS=linux go build -o ./bin/${BINARY_NAME}-linux *.go
	GOARCH=amd64 GOOS=windows go build -o ./bin/${BINARY_NAME}-windows *.go

clean:
	go clean
	rm ./bin/${BINARY_NAME}-darwin
	rm ./bin/${BINARY_NAME}-linux
	rm ./bin/${BINARY_NAME}-windows
