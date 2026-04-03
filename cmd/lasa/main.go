package main

import (
	"flag"
	"net"
	"net/http"
	"os"

	"tangled.org/anhgelus.world/lasa"
	"tangled.org/anhgelus.world/lasa/cmd/internal"
	"tangled.org/anhgelus.world/xrpc"
)

var (
	flags *flag.FlagSet
	help  bool
)

func init() {
	flags = flag.NewFlagSet("default", flag.PanicOnError)
	flags.BoolVar(&help, "h", false, "show the help")
}

var commands = []internal.Command{
	{Name: "publication", Usage: "display publications", Callback: handlePublication},
	{Name: "rss", Usage: "generate RSS", Callback: handleRSS},
}

var client xrpc.Client

func main() {
	flags.Parse(os.Args[1:])
	args := flags.Args()
	if len(args) == 0 {
		handleHelp()
		return
	}
	client = lasa.NewClient(http.DefaultClient, net.DefaultResolver, nil, 0, "local")
	command := args[0]
	var next []string
	if len(args) > 1 {
		next = args[1:]
	}
	for _, c := range commands {
		if c.Name == command {
			flags.Parse(next)
			next = flags.Args()
			c.Callback(next)
			return
		}
	}
	handleHelp()
	os.Exit(1)
}

func handleHelp() {
	internal.Usage(
		`lasa <command>`,
		`Lasa is a CLI tool.`,
		commands,
		flags,
		[]string{
			"lasa publication anhgelus.world\t-\tdisplay publications of anhgelus.world",
		},
	)
}
