# Lasa

Lasa is a stateless proxy that generates a RSS or an Atom feed from a [Standard.site](https://standard.site) 
publication.

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

You can use
```bash
just build-docker localhost/lasa
```
to build the Dockerfile containing `lasa` and `lasad`.
You can replace `localhost/lasa` by the name of the image.

### Installing

Building and installing binaries and man pages to `/usr/local/`:
```bash
just install
```
