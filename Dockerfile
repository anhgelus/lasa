FROM golang:alpine as builder

WORKDIR /app

COPY . .

RUN apk add just

RUN just build

FROM alpine:latest

WORKDIR /app

# expose default port
EXPOSE 8000

LABEL org.opencontainers.image.title='Lasa'
LABEL org.opencontainers.image.description='Stateless proxy that generates a RSS or an Atom feed from a Standard.site publication'
LABEL org.opencontainers.image.source='https://tangled.org/anhgelus.world/lasa'
LABEL org.opencontainers.image.url='https://tangled.org/anhgelus.world/lasa'
LABEL org.opencontainers.image.licenses='AGPL-3-only'

COPY README.md .
COPY --from=builder /app/build/lasad .
COPY --from=builder /app/build/lasa .

# generate default config file
RUN mkdir -p /etc/lasad && /app/lasad gen-config -c "/etc/lasad/config.toml"

ENTRYPOINT [ "/app/lasad", "-c", "/etc/lasad/config.toml" ]
