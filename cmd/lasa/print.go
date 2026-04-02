package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
)

func usage(syntax, usage string, examples []string, flags *flag.FlagSet) func() {
	return func() {
		fmt.Println("Usage:", syntax)
		fmt.Println(usage)
		w := tabwriter.NewWriter(os.Stdout, 0, 2, 1, ' ', 0)
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
}
