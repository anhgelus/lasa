# Lasa

Lasa is a stateless proxy that generates a RSS or an Atom feed from a [Standard.site](https://standard.site) 
publication.

Try it at [https://lasa.anhgelus.world](https://lasa.anhgelus.world)! 

## Usage

To list publications from an account:
```
https://lasa.example.org/{DID or Handle}
```

To get the RSS feed from a publication:
```
https://lasa.example.org/{DID or Handle}/{Record Key}/rss
```

To get the Atom feed from a publication:
```
https://lasa.example.org/{DID or Handle}/{Record Key}/atom
```

Examples:
```
https://lasa.example.org/did:plc:revjuqmkvrw6fnkxppqtszpv
https://lasa.example.org/did:plc:revjuqmkvrw6fnkxppqtszpv/3lwafzkjqm25s/rss
https://lasa.example.org/did:plc:revjuqmkvrw6fnkxppqtszpv/3lwafzkjqm25s/atom
```

## Dev

Requires **just** as a command runner.

Starts the web server:
```bash
just
# or
just dev
```

Starts Redis in Docker and exposes its port:
```bash
just redis
```

## Deploy

Lasa is a standalone binary that requires nothing.
You can use Redis as a cache.

Check [DEPLOYMENT.md](./DEPLOYMENT.md) for more information.

### Building

Building binaries:
```bash
just build
```

`build/lasad` is the daemon running the web server.
Run `lasad -h` to get the help.
Read `lasad(1)` for more information.

`build/lasa` is a CLI.
Run `lasa -h` to get the help.
Read `lasa(1)` for more information.

You must have **scdoc** installed to build the man pages.
If scdoc is not installed, it skips the building.

### Installing

Building and installing binaries and man pages to `/usr/local/`:
```bash
just install
```

### Docker

Lasa can be used easily with Docker.

You can build the Docker image containing `lasa` and `lasad` with:
```bash
just build-docker localhost/lasa
```
where `localhost/lasa` by the name of the image.

You can start the compose file in dev mode with:
```bash
just dev-docker
```

You can deploy the container in production by copying the `compose.yml` and running:
```bash
docker compose --profile prod up -d
# if you use podman
podman compose --profile prod up -d
```
It will use the official image [`anhgelus.world/lasa`](https://atcr.io/r/anhgelus.world/lasa) hosted on ATCR.
