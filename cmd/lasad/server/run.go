package server

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	glide "github.com/valkey-io/valkey-glide/go/v2"
	"tangled.org/anhgelus.world/lasa"
	"tangled.org/anhgelus.world/lasa/cmd/internal"
	"tangled.org/anhgelus.world/lasa/cmd/lasad/config"
	"tangled.org/anhgelus.world/xrpc"
)

//go:embed index.html author.html
var files embed.FS

type Publication struct {
	URL  string
	Link string
	Name string
	RKey string
}

func Run(ctx context.Context, cfg *config.Config, client xrpc.Client, cache *glide.Client, dur time.Duration) error {
	ctx = context.WithValue(ctx, keyCfg, cfg)
	ctx = context.WithValue(ctx, keyClient, client)
	ctx = context.WithValue(ctx, keyDir, NewDirectory(cache, dur))
	ctx = context.WithValue(ctx, keyLimiter, &Limiter{limited: make(map[string]*limited)})

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{id}/{rkey}/rss", func(w http.ResponseWriter, r *http.Request) {
		dir := r.Context().Value(keyDir).(*Directory)
		err := dir.Feed(r.Context(), w, r, "rss", lasa.GenerateRSS)
		if err != nil {
			HandleErrors(w, err)
			return
		}
	})
	mux.HandleFunc("GET /{id}/{rkey}/atom", func(w http.ResponseWriter, r *http.Request) {
		dir := r.Context().Value(keyDir).(*Directory)
		err := dir.Feed(r.Context(), w, r, "atom", lasa.GenerateAtom)
		if err != nil {
			HandleErrors(w, err)
			return
		}
	})
	mux.HandleFunc("GET /{id}/{$}", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		client := ctx.Value(keyClient).(xrpc.Client)
		did, err := lasa.Resolve(ctx, client.Directory(), r.PathValue("id"))
		if err != nil {
			HandleErrors(w, err)
			return
		}
		dir := ctx.Value(keyDir).(*Directory)
		b, err := dir.Author(ctx, did)
		if err != nil {
			HandleErrors(w, err)
			return
		}
		w.Write(b)
	})
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		b, err := files.ReadFile("index.html")
		if err != nil {
			HandleErrors(w, err)
			return
		}
		w.Write(b)
	})

	m := internal.NewMux(mux)
	m.Use(func(next internal.Handler, w *internal.StatusWriter, r *http.Request) {
		// not found favicon
		if strings.HasPrefix(r.RequestURI, "/favicon.ico") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		next(w, r)
	})
	m.Use(func(next internal.Handler, w *internal.StatusWriter, r *http.Request) {
		// timeouts request handling
		ctx, cancel := context.WithTimeoutCause(ctx, 15*time.Second, errors.New("handling timeouts"))
		defer cancel()

		var cancelCause context.CancelCauseFunc
		ctx, cancelCause = context.WithCancelCause(ctx)
		defer cancelCause(errors.New("handling finished"))

		ctx = context.WithValue(ctx, keyCancelCause, cancelCause)

		next(w, r.WithContext(ctx))
	})
	m.Use(internal.MiddlewareLog(func(ctx context.Context) context.CancelCauseFunc {
		return ctx.Value(keyCancelCause).(context.CancelCauseFunc)
	}, cfg.LogNotFound, cfg.LogBadRequest))
	m.Use(func(next internal.Handler, w *internal.StatusWriter, r *http.Request) {
		// rate limits
		limiter := ctx.Value(keyLimiter).(*Limiter)
		if limiter.isLimited(r) {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		next(w, r)
		limiter.handle(w, r)
	})

	return http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), m.Handle())
}
