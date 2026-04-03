package main

import (
	"context"
	"os"

	site "tangled.org/anhgelus.world/goat-site"
	"tangled.org/anhgelus.world/lasa"
	"tangled.org/anhgelus.world/lasa/cmd/internal"
	"tangled.org/anhgelus.world/xrpc"
	"tangled.org/anhgelus.world/xrpc/atproto"
)

func handleRSSUsage() {
	internal.Usage(
		`lasa rss <identifier> <rkey>`,
		`Generate the RSS for the given publication.`,
		nil,
		flags,
		[]string{
			"lasa publication did:web:example.org fooBar\t-\tgenerate RSS publication of did:web:example.org referenced by fooBar",
		},
	)
	if !help {
		os.Exit(1)
	}
}

func handleRSS(args []string) {
	if len(args) != 2 || help {
		handleRSSUsage()
		return
	}
	did, err := lasa.Resolve(context.Background(), client.Directory(), args[0])
	if err != nil {
		panic(err)
	}
	rkey, err := atproto.ParseRecordKey(args[1])
	if err != nil {
		return
	}
	pub, err := xrpc.GetRecord[*site.Publication](context.Background(), client, did, rkey, nil)
	if err != nil {
		panic(err)
	}
	err = lasa.GenerateRSS(context.Background(), client, os.Stdout, did, pub)
	if err != nil {
		panic(err)
	}
}
