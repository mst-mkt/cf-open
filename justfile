default: build

build:
    go build -o bin/cf-open ./cmd/cf-open

test:
    go test ./...

lint:
    golangci-lint run

fix:
    golangci-lint run --fix

clean:
    rm -rf bin

run: build
    ./bin/cf-open