package main

import (
	"flag"

	"tangled.org/anhgelus.world/lasa/cmd/internal"
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

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 || help {
		handleHelp()
		return
	}
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

func handlePublication(args []string) {
	if len(args) == 0 {
		internal.Usage(
			`lasa publication <identifier> [rkey]`,
			`List publications of identifier (can be a DID or an Handle) or display a specific publication referenced by its rkey`,
			nil,
			nil,
			[]string{
				"lasa publication anhgelus.world\t-\tdisplay publications of anhgelus.world",
				"lasa publication did:plc:123\t-\tdisplay publications of did:plc:123",
				"lasa publication did:web:example.org fooBar\t-\tdisplay publication of did:web:example.org referenced by fooBar",
			},
		)
		return
	}
}
