package internal

import (
	"context"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	site "tangled.org/anhgelus.world/goat-site"
	"tangled.org/anhgelus.world/xrpc"
	"tangled.org/anhgelus.world/xrpc/atproto"
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

func DisplayPublication(
	ctx context.Context,
	client xrpc.Client,
	did *atproto.DID,
	raw atproto.RawURI,
	pub *site.Publication,
) {
	uri, err := raw.URI(context.Background(), client.Directory())
	if err != nil {
		panic(err)
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 2, 1, ' ', 0)
	fmt.Fprintln(w, "Name:\t", pub.Name)
	fmt.Fprintln(w, "URL:\t", pub.URL.String())
	fmt.Fprintln(w, "AT URL:\t", raw)
	fmt.Fprintln(w, "RecordKey:\t", *uri.RecordKey())
	if pub.Description != nil {
		fmt.Fprintln(w, "Description:\t", *pub.Description)
	}
	ok, err := pub.Verify(context.Background(), client.HTTP(), did, *uri.RecordKey())
	if err != nil {
		fmt.Fprintln(w, "Verification:\t error:", err)
		return
	}
	fmt.Fprintf(w, "Verification:\t %v\n", ok)
	err = w.Flush()
	if err != nil {
		panic(err)
	}
}
