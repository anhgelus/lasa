package lasa

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	glide "github.com/valkey-io/valkey-glide/go/v2"
	"tangled.org/anhgelus.world/xrpc/atproto"
)

type Directory struct {
	inner    atproto.Directory
	cache    *glide.Client
	duration time.Duration
	limiter  *limitManyRequests[*atproto.DIDDocument]
}

func NewDirectory(dir atproto.Directory, cache *glide.Client, dur time.Duration) *Directory {
	return &Directory{
		inner:    dir,
		cache:    cache,
		limiter:  newLimitManyRequests[*atproto.DIDDocument](),
		duration: dur,
	}
}

func (d *Directory) fromCache(ctx context.Context, key string) *atproto.DIDDocument {
	if d.cache == nil {
		return nil
	}
	resp, err := d.cache.Get(ctx, key)
	var doc *atproto.DIDDocument
	if err == nil {
		b := resp.Value()
		err = json.Unmarshal([]byte(b), &doc)
		if err == nil {
			return doc
		} else {
			slog.Warn("cannot unmarshal cache response into DIDDocument", "resp", b)
		}
	}
	return nil
}

func (d *Directory) toCache(ctx context.Context, key string, doc *atproto.DIDDocument) {
	if d.cache == nil {
		return
	}
	b, err := json.Marshal(doc)
	if err != nil {
		slog.Warn("cannot marshal DIDDocument", "document", doc, "error", err)
		return
	}
	_, err = d.cache.Set(ctx, key, string(b))
	if err != nil {
		slog.Warn("cannot set DIDDocument in cache", "document", doc, "error", err)
		return
	}
	slog.Debug("DIDDocument set in cache")
	_, err = d.cache.ExpireAt(ctx, key, time.Now().Add(d.duration))
	if err != nil {
		slog.Warn("cannot set DIDDocument expire at", "document", doc, "error", err)
	}
}

func (d *Directory) ResolveHandle(ctx context.Context, h atproto.Handle) (*atproto.DIDDocument, error) {
	return resolve(ctx, d, h, d.inner.ResolveHandle)
}

func (d *Directory) ResolveDID(ctx context.Context, did *atproto.DID) (*atproto.DIDDocument, error) {
	return resolve(ctx, d, did, d.inner.ResolveDID)
}

func resolve[T fmt.Stringer](
	ctx context.Context,
	d *Directory,
	authority T,
	res func(context.Context, T) (*atproto.DIDDocument, error),
) (*atproto.DIDDocument, error) {
	key := authority.String()
	doc := d.fromCache(ctx, key)
	if doc != nil {
		slog.Debug("DIDDocument got from cache")
		return doc, nil
	}
	slog.Debug("cannot get DIDDocument from cache, requesting")

	return d.limiter.Do(key, func() (*atproto.DIDDocument, error) {
		doc, err := res(ctx, authority)
		if err != nil {
			return nil, err
		}
		d.toCache(ctx, key, doc)
		return doc, nil
	})
}
