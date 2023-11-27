# general
BINARY_NAME=os-diff

# build
build:
	go build -o ${BINARY_NAME} main.go

# run
run: build
	./${BINARY_NAME}

clean:
	go clean
	rm ${BINARY_NAME}

# run linter
lint:
	golangci-lint run --enable-all
