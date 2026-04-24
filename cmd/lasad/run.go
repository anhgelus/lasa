package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"tangled.org/anhgelus.world/lasa"
	"tangled.org/anhgelus.world/lasa/cmd/internal"
	"tangled.org/anhgelus.world/lasa/cmd/lasad/config"
	"tangled.org/anhgelus.world/lasa/cmd/lasad/server"
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
		ch <- server.Run(ctx, cfg, client, cache, dur)
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
