# general
BINARY_NAME=os-diff

# build
build:
	CGO_ENABLED=0 go build -o ${BINARY_NAME} .

# run
run: build
	./${BINARY_NAME}

clean:
	go clean
	rm ${BINARY_NAME}

# run linter
lint:
	golangci-lint run --enable-all

# run unit tests
test:
	go test -v ./...
