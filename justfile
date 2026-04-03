builder := 'go build -ldflags "-s -w"'

build: build-lasa build-lasad

build-lasa:
    {{builder}} -o build/lasa ./cmd/lasa/

build-lasad:
    {{builder}} -o build/lasad ./cmd/lasad/

