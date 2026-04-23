builder := 'go build -ldflags "-s -w"'
testConfig := '"test.toml"'
redis_container := 'redis'

docker := 'podman'

dev:
    if [[ ! -f {{testConfig}} ]]; then go run ./cmd/lasad/ gen-config -c {{testConfig}}; fi
    go run ./cmd/lasad/ -c {{testConfig}} -v

dev-docker:
    {{docker}} compose build
    {{docker}} compose up -d

redis:
    {{docker}} run --rm --name {{redis_container}} -p 6379:6379 -d docker.io/library/redis:alpine

stop:
    {{docker}} stop {{redis_container}}

build: build-lasa build-lasad

build-lasa:
    {{builder}} -o build/lasa ./cmd/lasa/
    # do not require building man pages
    -just build-doc lasa

build-lasad:
    {{builder}} -o build/lasad ./cmd/lasad/
    # do not require building man pages
    -just build-doc lasad

build-doc file:
    scdoc < {{file}}.1.scd > build/{{file}}.1

build-docker name:
    {{docker}} build -t {{name}} .

install: build
    mv build/lasa /usr/local/bin/
    mv build/lasad /usr/local/bin/
    # if cannot install man pages, skip
    -mkdir -p /usr/local/man/man1
    -mv build/lasa.1 /usr/local/man/man1/
    -mv build/lasad.1 /usr/local/man/man1/
