build: build-lasa

build-lasa:
    go build -ldflags "-s -w" -o build/lasa ./cmd/lasa/
