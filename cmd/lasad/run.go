package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"tangled.org/anhgelus.world/lasa/cmd/internal"
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

	mux := http.NewServeMux()

	ch := make(chan error, 1)

	go func() {
		ch <- http.ListenAndServe(":8000", mux)
	}()
	var err error
	select {
	case <-ctx.Done():
		err = ctx.Err()
	case err = <-ch:
	}
	if err != nil {
		panic(err)
	}
}
