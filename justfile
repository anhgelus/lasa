builder := 'go build -ldflags "-s -w"'
testConfig := '"test.toml"'
redis_container := 'redis'

docker := 'podman'
docker_profile := 'dev'

dev:
    if [[ ! -f {{testConfig}} ]]; then go run ./cmd/lasad/ gen-config -c {{testConfig}}; fi
    go run ./cmd/lasad/ -c {{testConfig}} -v

dev-docker:
    {{docker}} compose --profile {{docker_profile}} build --no-cache
    {{docker}} compose --profile {{docker_profile}} up -d

redis:
    {{docker}} run --rm --name {{redis_container}} -p 6379:6379 -d docker.io/library/redis:alpine

stop:
    {{docker}} stop {{redis_container}}

stop-docker:
    {{docker}} compose --profile {{docker_profile}} down

logs-docker:
    {{docker}} compose --profile {{docker_profile}} logs

build: build-lasa build-lasad

build-lasa:
    @{{builder}} -o build/lasa ./cmd/lasa/
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

publish-docker registry name tag:
    just build-docker {{registry / name}}:{{tag}}
    {{docker}} tag {{registry / name}}:{{tag}} {{registry / name}}:latest
    #{{docker}} push {{registry / name}}:{{tag}}
    #{{docker}} push {{registry / name}}:latest
