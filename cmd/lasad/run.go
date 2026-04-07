package main

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	glide "github.com/valkey-io/valkey-glide/go/v2"
	"tangled.org/anhgelus.world/lasa"
	"tangled.org/anhgelus.world/lasa/cmd/internal"
	"tangled.org/anhgelus.world/lasa/cmd/lasad/config"
	"tangled.org/anhgelus.world/xrpc"
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
	URL  string
	Link string
	Name string
	RKey string
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
		slog.Info("connected to valkey")
		dur = time.Duration(cfg.Cache.Duration) * time.Minute
	}
	client := lasa.NewClient(http.DefaultClient, net.DefaultResolver, cache, dur, cfg.Domain)
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

type statusWriter struct {
	http.ResponseWriter
	code int
}

func (w *statusWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.code = statusCode
}

func middlewares(h http.Handler, ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.RequestURI, "/favicon.ico") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
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
				debug.PrintStack()
				w.WriteHeader(http.StatusInternalServerError)
				log.Error("panic! (stack trace printed to stderr)", "duration", time.Since(now))
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
