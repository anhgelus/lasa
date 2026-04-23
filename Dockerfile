FROM golang:alpine as builder

WORKDIR /app

COPY . .

RUN apk add just

RUN just build

FROM alpine:latest

WORKDIR /app

# expose default port
EXPOSE 8000

COPY --from=builder /app/build/lasad .
COPY --from=builder /app/build/lasa .

# generate default config file
RUN mkdir -p /etc/lasad && /app/lasad gen-config -c "/etc/lasad/config.toml"

ENTRYPOINT [ "/app/lasad", "-c", "/etc/lasad/config.toml" ]
