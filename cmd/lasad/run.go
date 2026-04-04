package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	glide "github.com/valkey-io/valkey-glide/go/v2"
	site "tangled.org/anhgelus.world/goat-site"
	"tangled.org/anhgelus.world/lasa"
	"tangled.org/anhgelus.world/lasa/cmd/internal"
	"tangled.org/anhgelus.world/lasa/cmd/lasad/config"
	"tangled.org/anhgelus.world/xrpc"
	"tangled.org/anhgelus.world/xrpc/atproto"
)

//go:embed index.html author.html
var files embed.FS

func handleRunHelp() {
	internal.Usage(
		`lasad run`,
		`Run the daemon`,
		nil,
		flags,
		[]string{
			"lased run\t-\trun the daemon with the default config file",
		},
	)
	if !help {
		os.Exit(1)
	}
}

type Publication struct {
	Link string
	Name string
}

func handleRun(args []string) {
	if len(args) != 0 || help {
		handleRunHelp()
		return
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	slog.Info("loading config", "path", configPath)
	cfg, err := config.Load(configPath)
	if err != nil {
		panic(err)
	}
	ctx = context.WithValue(ctx, keyCfg, cfg)

	var cache *glide.Client
	var dur time.Duration
	if cfg.Cache != nil {
		cache, err = cfg.Cache.Connect()
		if err != nil {
			panic(err)
		}
		dur = time.Duration(cfg.Cache.Duration) * time.Minute
	}
	client := lasa.NewClient(http.DefaultClient, net.DefaultResolver, cache, dur, cfg.Domain)
	ctx = context.WithValue(ctx, keyClient, client)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{id}/{rkey}/rss", func(w http.ResponseWriter, r *http.Request) {
		did, pub, ok := getPub(w, r)
		if !ok {
			return
		}
		w.Header().Set("Content-Type", "application/rss+xml")
		err = lasa.GenerateRSS(ctx, client, w, did, pub)
		if err != nil {
			panic(err)
		}
	})
	mux.HandleFunc("GET /{id}/{rkey}/atom", func(w http.ResponseWriter, r *http.Request) {
		did, pub, ok := getPub(w, r)
		if !ok {
			return
		}
		w.Header().Set("Content-Type", "application/atom+xml")
		err = lasa.GenerateAtom(ctx, client, w, did, pub)
		if err != nil {
			panic(err)
		}
	})
	mux.HandleFunc("GET /{id}/{$}", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		client := ctx.Value(keyClient).(xrpc.Client)
		did, err := lasa.Resolve(ctx, client.Directory(), r.PathValue("id"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		doc, err := client.Directory().ResolveDID(ctx, did)
		if err != nil {
			panic(err)
		}
		h, _ := doc.Handle()
		v := struct {
			Author       string
			Publications []Publication
		}{Author: h.String()}
		pubs, _, err := xrpc.ListRecords[*site.Publication](ctx, client, did, 0, "", false)
		if err != nil {
			panic(err)
		}
		v.Publications = make([]Publication, len(pubs))
		for i, pub := range pubs {
			uri, err := pub.URI.URI(ctx, client.Directory())
			if err != nil {
				panic(err)
			}
			link := fmt.Sprintf("/%s/%s", did, uri.RecordKey())
			v.Publications[i] = Publication{link, pub.Value.Name}
		}
		err = template.Must(template.ParseFS(files, "author.html")).ExecuteTemplate(w, "author.html", v)
		if err != nil {
			panic(err)
		}
	})
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		b, err := files.ReadFile("index.html")
		if err != nil {
			panic(err)
		}
		w.Write(b)
	})

	ch := make(chan error, 1)

	go func() {
		slog.Info("starting")
		ch <- http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), middlewares(mux, ctx))
	}()
	select {
	case <-ctx.Done():
		err = context.Cause(ctx)
	case err = <-ch:
	}
	slog.Info("exiting")
	if err != nil {
		panic(err)
	}
}

func middlewares(h http.Handler, ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("request", "path", r.URL.Path)
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic!", "err", err)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getPub(w http.ResponseWriter, r *http.Request) (*atproto.DID, xrpc.RecordStored[*site.Publication], bool) {
	var pub xrpc.RecordStored[*site.Publication]
	ctx := r.Context()
	client := ctx.Value(keyClient).(xrpc.Client)
	did, err := lasa.Resolve(ctx, client.Directory(), r.PathValue("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, pub, false
	}
	rkey, err := atproto.ParseRecordKey(r.PathValue("rkey"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, pub, false
	}
	pub, err = xrpc.GetRecord[*site.Publication](ctx, client, did, rkey, nil)
	if err != nil {
		if err, ok := errors.AsType[xrpc.ErrStandardResponse](err); ok {
			if errors.Is(err, xrpc.ErrRecordNotFound) {
				w.WriteHeader(http.StatusNotFound)
				return nil, pub, false
			}
			panic(err)
		} else {
			panic(err)
		}
	}
	return did, pub, true
}
