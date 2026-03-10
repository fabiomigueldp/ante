.PHONY: build build-sim test sim run clean

build:
	go build -o ./bin/ante ./cmd/ante

build-sim:
	go build -o ./bin/sim ./cmd/sim

run:
	go run ./cmd/ante

test:
	go test ./... -count=1

sim:
	go run ./cmd/sim -hands 1000

clean:
	go clean
