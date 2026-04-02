package lasa

import (
	"net"
	"net/http"
	"time"

	"github.com/valkey-io/valkey-go"
	"tangled.org/anhgelus.world/xrpc"
	"tangled.org/anhgelus.world/xrpc/atproto"
)

func NewClient(client *http.Client, resolver *net.Resolver, cache valkey.Client, dur time.Duration) xrpc.Client {
	dir := NewDirectory(atproto.NewDirectory(client, resolver), cache, dur)
	return xrpc.NewClient(client, dir)
}
