# Lasa

Lasa is stateless proxy that generates an RSS/Atom feed from a [Standard.site](https://Standard.site) publication.

## Usage

To get the RSS feed from a publication:
```
https://lasa.example.org/{DID or Handle}/{Record Key}/rss
```

To get the Atom feed from a publication:
```
https://lasa.example.org/{DID or Handle}/{Record Key}/atom
```

To list publications from an account:
```
https://lasa.example.org/{DID or Handle}
```

## Deploy

Lasa is a standalone binary that requires nothing.
You can use Valkey as a cache.

Building binaries:
```bash
just build
```

`build/lasad` is the daemon running the web server.

`build/lasa` is a CLI.
