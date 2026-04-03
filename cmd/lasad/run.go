package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"tangled.org/anhgelus.world/lasa/cmd/internal"
	"tangled.org/anhgelus.world/lasa/cmd/lasad/config"
)

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

func handleRun(args []string) {
	if len(args) != 0 || help {
		handleRunHelp()
		return
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	slog.Info("loading config...", "path", configPath)
	cfg, err := config.Load(configPath)
	if err != nil {
		panic(err)
	}
	ctx = context.WithValue(ctx, keyCfg, cfg)

	mux := http.NewServeMux()

	ch := make(chan error, 1)

	go func() {
		slog.Info("starting...")
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
