package lasa

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/valkey-io/valkey-go"
	site "tangled.org/anhgelus.world/goat-site"
	"tangled.org/anhgelus.world/xrpc"
	"tangled.org/anhgelus.world/xrpc/atproto"
)

func NewClient(client *http.Client, resolver *net.Resolver, cache valkey.Client, dur time.Duration) xrpc.Client {
	client.Timeout = 30 * time.Second
	dir := NewDirectory(atproto.NewDirectory(client, resolver), cache, dur)
	return xrpc.NewClient(client, dir, "Lasa/v0.1.0 (Linux; +https://tangled.org/anhgelus.world/lasa)")
}

func ListDocuments(
	ctx context.Context,
	client xrpc.Client,
	did *atproto.DID,
	pub atproto.RawURI,
) ([]xrpc.RecordStored[*site.Document], error) {
	rawDocs, _, err := xrpc.ListRecords[*site.Document](ctx, client, did, 0, "", false)
	if err != nil {
		return nil, err
	}
	var docs []xrpc.RecordStored[*site.Document]
	for _, doc := range rawDocs {
		if !doc.Value.Site.IsAT() {
			continue
		}
		if doc.Value.Site.AT().String() == pub.String() {
			docs = append(docs, doc)
		}
	}
	return docs, nil
}
