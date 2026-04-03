package main

import (
	"flag"
	"os"

	"tangled.org/anhgelus.world/lasa/cmd/internal"
	"tangled.org/anhgelus.world/lasa/cmd/lasad/config"
)

var (
	flags      *flag.FlagSet
	help       = false
	configPath = config.DefaultPath
)

func init() {
	flags = flag.NewFlagSet("default", flag.PanicOnError)
	flags.BoolVar(&help, "h", help, "display the help")
	flags.StringVar(&configPath, "c", configPath, "path to the config")
}

var commands = []internal.Command{
	{Name: "run", Usage: "run the daemon", Callback: handleRun},
}

func main() {
	flags.Parse(os.Args[1:])
	args := flags.Args()
	command := "run"
	if len(args) > 0 {
		command = args[0]
	}
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
	handleHelp(args)
	os.Exit(1)
}

func handleHelp([]string) {
	internal.Usage(
		`lasad <command>`,
		`Daemon running Lasa.`,
		nil,
		flags,
		[]string{
			"lasad\t-\trun the daemon",
			"lasad -c /foo/bar.toml\t-\trun the daemon with the config at /foo/bar.toml",
		},
	)
}
