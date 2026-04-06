package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"time"

	glide "github.com/valkey-io/valkey-glide/go/v2"
	site "tangled.org/anhgelus.world/goat-site"
	"tangled.org/anhgelus.world/lasa"
	"tangled.org/anhgelus.world/xrpc"
	"tangled.org/anhgelus.world/xrpc/atproto"
)

type Directory struct {
	cache    *glide.Client
	duration time.Duration
	limiter  *lasa.LimitManyRequests[[]byte]
}

func NewDirectory(cache *glide.Client, dur time.Duration) *Directory {
	return &Directory{
		cache:    cache,
		duration: dur,
		limiter:  lasa.NewLimitManyRequests[[]byte](),
	}
}

func (d *Directory) fromCache(ctx context.Context, key string) []byte {
	if d.cache == nil {
		return nil
	}
	resp, err := d.cache.Get(ctx, key)
	if err != nil || resp.IsNil() {
		return nil
	}
	return []byte(resp.Value())
}

func (d *Directory) toCache(ctx context.Context, key string, b []byte) {
	if d.cache == nil {
		return
	}
	_, err := d.cache.Set(ctx, key, string(b))
	if err != nil {
		slog.Warn("cannot set bytes in cache", "bytes", b, "error", err)
		return
	}
	slog.Debug("bytes set in cache")
	_, err = d.cache.Expire(ctx, key, d.duration)
	if err != nil {
		slog.Warn("cannot set bytes expire", "bytes", b, "error", err)
	}
}

func (d *Directory) Author(ctx context.Context, did *atproto.DID) ([]byte, error) {
	key := did.String() + ":publications"
	b := d.fromCache(ctx, key)
	if b != nil {
		slog.Debug("author got from cache")
	}
	slog.Debug("cannot get author from cache", "did", did)

	return d.limiter.Do(key, func() ([]byte, error) {
		client := ctx.Value(keyClient).(xrpc.Client)
		doc, err := client.Directory().ResolveDID(ctx, did)
		if err != nil {
			return nil, err
		}
		h, _ := doc.Handle()
		v := struct {
			Author       string
			Publications []Publication
		}{Author: h.String()}
		pubs, _, err := xrpc.ListRecords[*site.Publication](ctx, client, did, 0, "", false)
		if err != nil {
			return nil, err
		}
		v.Publications = make([]Publication, len(pubs))
		for i, pub := range pubs {
			uri, err := pub.URI.URI(ctx, client.Directory())
			if err != nil {
				slog.Error("cannot get uri for publication", "pub", pub.URI)
				continue
			}
			link := fmt.Sprintf("/%s/%s", did, uri.RecordKey())
			v.Publications[i] = Publication{pub.Value.URL.String(), link, pub.Value.Name, uri.RecordKey().String()}
		}
		var bf bytes.Buffer
		err = template.Must(template.ParseFS(files, "author.html")).ExecuteTemplate(&bf, "author.html", v)
		if err != nil {
			return nil, err
		}
		b, err := io.ReadAll(&bf)
		if err != nil {
			return nil, err
		}
		d.toCache(ctx, key, b)
		return b, nil
	})
}

func (d *Directory) Feed(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	kind string,
	gen func(context.Context, xrpc.Client, io.Writer, *atproto.DID, xrpc.RecordStored[*site.Publication]) error,
) error {
	w.Header().Set("Content-Type", "application/"+kind+"+xml")
	client := ctx.Value(keyClient).(xrpc.Client)
	did, err := lasa.Resolve(ctx, client.Directory(), r.PathValue("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}
	rkey, err := atproto.ParseRecordKey(r.PathValue("rkey"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}
	key := did.String() + ":" + rkey.String() + ":" + kind
	b := d.fromCache(ctx, key)
	if b != nil {
		slog.Debug("feed got from cache", "kind", kind)
		w.Write(b)
		return nil
	}
	slog.Debug("cannot get feed from cache", "did", did, "kind", kind)

	b, err = d.limiter.Do(key, func() ([]byte, error) {
		pub, ok, err := getPub(ctx, did, rkey)
		if !ok {
			return nil, nil
		}
		var bf bytes.Buffer
		err = gen(ctx, client, &bf, did, pub)
		if err != nil {
			return nil, err
		}
		b, err := io.ReadAll(&bf)
		if err != nil {
			return nil, err
		}
		d.toCache(ctx, key, b)
		return b, nil
	})
	if err != nil {
		return err
	}
	if b == nil {
		w.WriteHeader(http.StatusNotFound)
		return nil
	}
	w.Write(b)
	return err
}

func getPub(ctx context.Context, did *atproto.DID, rkey atproto.RecordKey) (xrpc.RecordStored[*site.Publication], bool, error) {
	client := ctx.Value(keyClient).(xrpc.Client)
	pub, err := xrpc.GetRecord[*site.Publication](ctx, client, did, rkey, nil)
	if err != nil {
		if err, ok := errors.AsType[xrpc.ErrStandardResponse](err); ok {
			if errors.Is(err, xrpc.ErrRecordNotFound) {
				return pub, false, nil
			}
			return pub, false, err
		} else {
			return pub, false, err
		}
	}
	return pub, true, nil
}
