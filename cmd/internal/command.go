package internal

import (
	"flag"
	"os"
)

type Command struct {
	Name     string
	Usage    string
	Callback func([]string)
}

func HandleCommands(command string, args []string, cmds []Command, flags *flag.FlagSet, help func()) {
	for _, c := range cmds {
		if c.Name == command {
			flags.Parse(args)
			args = flags.Args()
			c.Callback(args)
			return
		}
	}
	help()
	os.Exit(1)
}
