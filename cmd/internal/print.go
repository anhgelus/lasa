package internal

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
)

func Usage(syntax, usage string, commands []Command, flags *flag.FlagSet, examples []string) {
	fmt.Println("Usage:", syntax)
	fmt.Println(usage)
	w := tabwriter.NewWriter(os.Stdout, 0, 2, 1, ' ', 0)
	if commands != nil {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Commands:")
		for _, c := range commands {
			fmt.Fprintln(w, "\t", c.Name, "\t-\t", c.Usage)
		}
	}
	if flags != nil {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Options:")
		flags.VisitAll(func(f *flag.Flag) {
			fmt.Fprintln(w, "\t-", f.Name, "\t", f.Usage, "\t(default:", f.DefValue, ")")
		})
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	for _, s := range examples {
		fmt.Fprintln(w, "\t", s)
	}
	err := w.Flush()
	if err != nil {
		panic(err)
	}
}
