builder := 'go build -ldflags "-s -w"'
testConfig := '"test.toml"'

build: build-lasa build-lasad

build-lasa:
    {{builder}} -o build/lasa ./cmd/lasa/
    just build-doc lasa

build-lasad:
    {{builder}} -o build/lasad ./cmd/lasad/
    just build-doc lasad

test:
    if [[ ! -f {{testConfig}} ]]; then go run ./cmd/lasad/ gen-config -c {{testConfig}}; fi
    go run ./cmd/lasad/ -c {{testConfig}}

build-doc file:
    scdoc < {{file}}.1.scd > build/{{file}}.1
