package main

import (
	"flag"
	"net"
	"net/http"

	"tangled.org/anhgelus.world/lasa"
	"tangled.org/anhgelus.world/lasa/cmd/internal"
	"tangled.org/anhgelus.world/xrpc"
)

var (
	help bool
)

func init() {
	flag.BoolVar(&help, "h", false, "show the help")
}

var commands = []internal.Command{
	{Name: "publication", Usage: "works with publication", Callback: handlePublication},
}

var client xrpc.Client

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 || help {
		handleHelp()
		return
	}
	client = lasa.NewClient(http.DefaultClient, net.DefaultResolver, nil, 0)
	command := args[0]
	var next []string
	if len(args) > 1 {
		next = args[1:]
	}
	for _, c := range commands {
		if c.Name == command {
			c.Callback(next)
			return
		}
	}
	handleHelp()
}

func handleHelp() {
	internal.Usage(
		`lasa <command>`,
		`Lasa is a CLI tool.`,
		commands,
		nil,
		[]string{
			"lasa publication anhgelus.world\t-\tdisplay publications of anhgelus.world",
		},
	)
}
