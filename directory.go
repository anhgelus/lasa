package lasa

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/valkey-io/valkey-go"
	"tangled.org/anhgelus.world/xrpc/atproto"
)

type Directory struct {
	inner    atproto.Directory
	cache    valkey.Client
	duration time.Duration
	limiter  *limitManyRequests[*atproto.DIDDocument]
}

func NewDirectory(dir atproto.Directory, cache valkey.Client) *Directory {
	return &Directory{
		inner: dir,
		cache: cache,
	}
}

func (d *Directory) fromCache(ctx context.Context, key string) *atproto.DIDDocument {
	if d.cache == nil {
		return nil
	}
	resp := d.cache.Do(ctx, d.cache.B().Get().Key(key).Build())
	err := resp.Error()
	var doc *atproto.DIDDocument
	if err == nil {
		b, err := resp.AsBytes()
		if err == nil {
			err = json.Unmarshal(b, &doc)
			if err == nil {
				return doc
			} else {
				slog.Warn("cannot unmarshal cache response into DIDDocument", "resp", b)
			}
		} else {
			slog.Warn("cannot convert cache response into bytes", "resp", resp)
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
	err = d.cache.Do(ctx, d.cache.B().Set().Key(key).Value(string(b)).Build()).Error()
	if err != nil {
		slog.Warn("cannot set DIDDocument in cache", "document", doc, "error", err)
	}
}

func (d *Directory) ResolveHandle(ctx context.Context, h atproto.Handle) (*atproto.DIDDocument, error) {
	key := h.String()
	doc := d.fromCache(ctx, key)
	if doc != nil {
		return doc, nil
	}
	slog.Debug("cannot get DIDDocument from cache")

	return d.limiter.Do(key, func() (*atproto.DIDDocument, error) {
		doc, err := d.inner.ResolveHandle(ctx, h)
		if err != nil {
			return nil, err
		}
		d.toCache(ctx, key, doc)
		return doc, nil
	})
}

func (d *Directory) ResolveDID(ctx context.Context, did *atproto.DID) (*atproto.DIDDocument, error) {
	key := did.String()
	doc := d.fromCache(ctx, key)
	if doc != nil {
		return doc, nil
	}
	slog.Debug("cannot get DIDDocument from cache")

	return d.limiter.Do(key, func() (*atproto.DIDDocument, error) {
		doc, err := d.inner.ResolveDID(ctx, did)
		if err != nil {
			return nil, err
		}
		d.toCache(ctx, key, doc)
		return doc, nil
	})
}
