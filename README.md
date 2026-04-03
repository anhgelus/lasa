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

## Deploy

Lasa is a standalone binary that requires nothing.
You can use Valkey as a cache.

Building binaries:
```bash
just build
```

`build/lasad` is the daemon running the web server.
Run `lasad -h` to get the help.

`build/lasa` is a CLI.
Run `lasa -h` to get the help.

You must have **scdoc** installed to build the man pages.
