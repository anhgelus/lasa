package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"tangled.org/anhgelus.world/lasa"
	"tangled.org/anhgelus.world/lasa/cmd/internal"
	"tangled.org/anhgelus.world/lasa/cmd/lasad/config"
	"tangled.org/anhgelus.world/lasa/cmd/lasad/server"
	"tangled.org/anhgelus.world/ljus"
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

	slog.Info("loading config", "path", configPath, "version", internal.Version)
	cfg, err := config.Load(configPath)
	if err != nil {
		panic(err)
	}
	var cache *redis.Client
	var dur time.Duration
	if cfg.Cache != nil {
		cache, err = cfg.Cache.Connect()
		if err != nil {
			slog.Error("cannot connect to redis", "error", err)
			os.Exit(3)
		}
		slog.Info("connected to redis")
		dur = time.Duration(cfg.Cache.Duration) * time.Minute
	}
	client := lasa.NewClient(http.DefaultClient, net.DefaultResolver, cache, dur, cfg.Domain)

	ch := make(chan error, 1)

	go func() {
		slog.Info("starting")
		if cfg.Listen.TCP == nil && cfg.Listen.Unix == nil {
			panic("no listen address set")
		}
		var l net.Listener
		if cfg.Listen.Unix != nil {
			l, err = net.Listen("unix", *cfg.Listen.Unix)
			defer func() {
				err := os.Remove(*cfg.Listen.Unix)
				if err != nil {
					slog.Error("cannot delete socket", "path", *cfg.Listen.Unix)
				}
			}()
		} else {
			l, err = net.Listen("tcp", *cfg.Listen.TCP)
		}
		if err != nil {
			panic(err)
		}
		s, err := server.New(ctx, cfg, client, cache, dur)
		if err != nil {
			panic(err)
		}
		if cfg.Listen.FastCGI {
			ch <- fcgi.Serve(l, s.Handler())
		} else {
			s.Use(func(next ljus.Handler, w *ljus.StatusWriter, r *http.Request) {
				addr := r.Header.Get("X-Real-Ip")
				if addr == "" {
					addr = r.Header.Get("X-Forwarded-For")
				}
				if addr == "" {
					addr = r.RemoteAddr
				}
				r.RemoteAddr = addr
				next(w, r)
			})
			ch <- http.Serve(l, s.Handler())
		}
	}()
	select {
	case <-ctx.Done():
		slog.Warn("received stop signal")
	case err = <-ch:
	}
	slog.Info("exiting")
	if err != nil {
		panic(err)
	}
}
