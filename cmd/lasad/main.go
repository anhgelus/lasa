package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/nyttikord/logos"
	"tangled.org/anhgelus.world/lasa/cmd/internal"
	"tangled.org/anhgelus.world/lasa/cmd/lasad/config"
)

var (
	flags      *flag.FlagSet
	help       = false
	verbose    = false
	noColor    = false
	configPath = config.DefaultPath
)

func init() {
	flags = flag.NewFlagSet("default", flag.PanicOnError)
	flags.BoolVar(&help, "h", help, "display the help")
	flags.BoolVar(&verbose, "v", verbose, "increase verbosity")
	flags.BoolVar(&noColor, "no-color", noColor, "disable colors")
	flags.StringVar(&configPath, "c", configPath, "path to the config")
}

var commands = []internal.Command{
	{Name: "run", Usage: "run the daemon", Callback: handleRun},
	{Name: "gen-config", Usage: "generate the config file", Callback: handleGenConfig},
}

func main() {
	flags.Parse(os.Args[1:])
	if help {
		handleHelp()
		return
	}
	logOpts := &logos.Options{PrintStackTrace: true, DisableColor: noColor}
	if verbose {
		logOpts.Level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(logos.New(os.Stdout, logOpts)))
	args := flags.Args()
	command := "run"
	if len(args) > 0 {
		command = args[0]
	}
	var next []string
	if len(args) > 1 {
		next = args[1:]
	}
	internal.HandleCommands(command, next, commands, flags, handleHelp)
}

func handleHelp() {
	internal.Usage(
		`lasad <command>`,
		`Daemon running Lasa.`,
		commands,
		flags,
		[]string{
			"lasad\t-\trun the daemon",
			"lasad -c /foo/bar.toml\t-\trun the daemon with the config at /foo/bar.toml",
		},
	)
}
