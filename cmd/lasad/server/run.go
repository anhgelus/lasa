package server

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	glide "github.com/valkey-io/valkey-glide/go/v2"
	"tangled.org/anhgelus.world/lasa"
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

	return http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), middlewares(mux, ctx))
}

type statusWriter struct {
	http.ResponseWriter
	code int
}

func (w *statusWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.code = statusCode
}

func middlewares(h http.Handler, parent context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.RequestURI, "/favicon.ico") {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// timeouts request handling
		ctx, cancel := context.WithTimeoutCause(parent, 15*time.Second, errors.New("handling timeouts"))
		defer cancel()

		var cancelCause context.CancelCauseFunc
		ctx, cancelCause = context.WithCancelCause(ctx)
		defer cancelCause(errors.New("handling finished"))

		limiter := ctx.Value(keyLimiter).(*Limiter)
		status := &statusWriter{w, http.StatusOK}
		log := slog.With("uri", r.RequestURI)
		if limiter.isLimited(r) {
			status.WriteHeader(http.StatusTooManyRequests)
			limiter.handle(status, r, log.With("status", http.StatusTooManyRequests))
			return
		}

		now := time.Now()
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Error("panic!", "error", err, "duration", time.Since(now))
				switch e := err.(type) {
				case error:
					cancelCause(e)
				case string:
					cancelCause(errors.New(e))
				default:
					log.Warn(
						"cannot set cancel cause, because error type is not supported",
						"type", fmt.Sprintf("%T", e),
					)
				}
			}
		}()

		h.ServeHTTP(status, r.WithContext(ctx))

		log = log.With("status", status.code, "duration", time.Since(now))
		if status.code < 400 {
			log.Debug("handled")
		} else if status.code < 500 {
			cfg := ctx.Value(keyCfg).(*config.Config)
			level := slog.LevelDebug
			if (status.code == http.StatusNotFound && cfg.LogNotFound) ||
				(status.code == http.StatusBadRequest && cfg.LogBadRequest) ||
				(status.code != http.StatusNotFound && status.code != http.StatusBadRequest) {
				level = slog.LevelWarn
			}
			log.Log(context.Background(), level, "invalid request")
		} else {
			log.Error("error while handling request")
		}

		limiter.handle(status, r, log)
	})
}
