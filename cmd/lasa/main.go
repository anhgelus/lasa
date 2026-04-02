package main

import (
	"flag"
)

var (
	help bool
)

func init() {
	flag.BoolVar(&help, "h", false, "show the help")
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
	switch command {
	case "help":
		handleHelp()
	case "publication":
		handlePublication(next)
	}
}

func handleHelp() {
	usage(
		`lasa <command>`,
		`Lasa is a CLI tool.`,
		[]string{
			"lasa publication\t-\tworks with publication",
		},
		nil,
	)()
}

func handlePublication(args []string) {
	if len(args) == 0 {
		usage(
			`lasa publication <identifier> [rkey]`,
			`List publications of identifier (can be a DID or an Handle) or display a specific publication referenced by its rkey`,
			[]string{
				"lasa publication anhgelus.world\t-\tdisplay publications of anhgelus.world",
				"lasa publication did:plc:123\t-\tdisplay publications of did:plc:123",
				"lasa publication did:web:example.org fooBar\t-\tdisplay publication of did:web:example.org referenced by fooBar",
			},
			nil,
		)()
		return
	}
}
