package server

import (
	"context"
	"embed"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"tangled.org/anhgelus.world/lasa"
	"tangled.org/anhgelus.world/lasa/cmd/lasad/config"
	"tangled.org/anhgelus.world/ljus"
	"tangled.org/anhgelus.world/ljus/middleware"
	"tangled.org/anhgelus.world/xrpc"
)

//go:embed index.html author.html style.css
var files embed.FS

type Publication struct {
	URL  string
	Link string
	Name string
	RKey string
}

func New(ctx context.Context, cfg *config.Config, client xrpc.Client, cache *redis.Client, dur time.Duration) (*ljus.Server, error) {
	ctx = context.WithValue(ctx, keyCfg, cfg)
	ctx = context.WithValue(ctx, keyClient, client)
	ctx = context.WithValue(ctx, keyDir, NewDirectory(cache, dur))
	ctx = context.WithValue(ctx, keyLimiter, &Limiter{limited: make(map[string]*limited)})

	mux := ljus.New()
	mux.Use(func(next ljus.Handler, w *ljus.StatusWriter, r *http.Request) {
		// not found favicon
		if strings.HasPrefix(r.RequestURI, "/favicon.ico") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		next(w, r)
	})
	mux.Use(func(next ljus.Handler, w *ljus.StatusWriter, r *http.Request) {
		// timeouts request handling
		ctx, cancel := context.WithTimeoutCause(ctx, 15*time.Second, errors.New("handling timeouts"))
		defer cancel()

		var cancelCause context.CancelCauseFunc
		ctx, cancelCause = context.WithCancelCause(ctx)
		defer cancelCause(errors.New("handling finished"))

		ctx = context.WithValue(ctx, keyCancelCause, cancelCause)

		next(w, r.WithContext(ctx))
	})
	mux.Use(middleware.SecurityHeaders(cfg.Domain, dur))
	mux.Use(middleware.Log(slog.Default(), func(ctx context.Context) context.CancelCauseFunc {
		return ctx.Value(keyCancelCause).(context.CancelCauseFunc)
	}, cfg.LogNotFound, cfg.LogBadRequest))
	mux.Use(func(next ljus.Handler, w *ljus.StatusWriter, r *http.Request) {
		// rate limits
		limiter := ctx.Value(keyLimiter).(*Limiter)
		if limiter.isLimited(r) {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		next(w, r)
		limiter.handle(w, r)
	})

	mux.HandleFunc("GET /{id}/{rkey}/rss", func(w http.ResponseWriter, r *http.Request) {
		dir := r.Context().Value(keyDir).(*Directory)
		err := dir.Feed(r.Context(), w, r, "rss", lasa.GenerateRSS)
		if err != nil {
			HandleErrors(w, err)
			return
		}
	}).Name("rss")
	mux.HandleFunc("GET /{id}/{rkey}/atom", func(w http.ResponseWriter, r *http.Request) {
		dir := r.Context().Value(keyDir).(*Directory)
		err := dir.Feed(r.Context(), w, r, "atom", lasa.GenerateAtom)
		if err != nil {
			HandleErrors(w, err)
			return
		}
	}).Name("atom")
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
	}).Name("list")
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		b, err := files.ReadFile("index.html")
		if err != nil {
			HandleErrors(w, err)
			return
		}
		w.Write(b)
	}).Name("root")
	mux.HandleFunc("GET /style.css", func(w http.ResponseWriter, r *http.Request) {
		b, err := files.ReadFile("style.css")
		if err != nil {
			HandleErrors(w, err)
			return
		}
		w.Header().Add("Content-Type", "text/css")
		w.Write(b)
	}).Name("css")

	return mux, nil
}
