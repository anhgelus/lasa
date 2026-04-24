# Deploying Lasa

Lasa can be deployed as a standalone binary or inside a container.

## Standalone binary

Clone the repo with:
```bash
git clone -b <tag> https://tangled.org/anhgelus.world/lasa .
```
where `<tag>` is the tag that you want to use.

We recommend you to install `just` (a command runner).
Of course, you can manually execute commands.
If you want to have the man pages installed, ensure that `scdoc` is here too.

Then, you can run
```bash
just install
```
to install the binaries and the man pages in `/usr/local/`.
If `/usr/local/man/man1` doesn't exist, it tries to create it.
You can create it before to avoid running the command as root.

If you don't have a supported Go version (e.g., the version installed is too old), you can use the environment variable
`GOTOOLCHAIN` to set it, e.g.,
```bash
GOTOOLCHAIN=go1.26.2 just install
```

Then, you can generate the config file with `lasad gen-config` at `/etc/lasad.toml`.
See `lasad(1)` for more information.

## Container

The official image is `atcr.io/anhgelus.world/lasa` and is based on Alpine Linux.
Sadly, to download it, you must be connected to `actr.io`.

You can also build the image by yourself by simply cloning the repo and running:
```bash
docker build -t lasa .
```

An example `compose.yml` is available in the repo.
The profile `prod` uses the official image.

The config file is stored in `/etc/lasad/config.toml`.
You can mount it with `-v ./config:/etc/lasad/`.
The default config file is already generated.
The exposed port is `8000`.

## Configuration

The config file only requires two informations: the port and the domain (for security headers).
```toml
domain = "lasa.example.org"
port = 8000
```

You can specify the legal notice with
```toml
legal_notice_url = "https://example.org/legal"
```

If you want to log 400 and 404 as warning, uncomment these lines:
```toml
# if you want to log HTTP 404 responses
log_not_found = true
# if you want to log HTTP 400 responses
log_bad_request = true
```

### Redis

Lasa supports Redis as a cache.
You can connect it by uncommenting and filling the required information in `cache` section of the config file, e.g.
```toml
[cache]
host = "localhost"
port = 6379
db = 0
duration = 60 # cache duration in minutes
```

If your Redis server requires auth, you can fill these information in `cache.auth` section, e.g.
```toml
[cache.auth]
username = "foo"
password = "bar"
client_name = "baz"
```
