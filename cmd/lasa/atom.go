package main

import (
	"context"
	"os"

	"tangled.org/anhgelus.world/lasa"
	"tangled.org/anhgelus.world/lasa/cmd/internal"
)

func handleAtomUsage() {
	internal.Usage(
		`lasa atom <identifier> <rkey>`,
		`Generate the Atom feed for the given publication.`,
		nil,
		flags,
		[]string{
			"lasa atom did:web:example.org fooBar\t-\tgenerate Atom feed of did:web:example.org referenced by fooBar",
		},
	)
	if !help {
		os.Exit(1)
	}
}

func handleAtom(args []string) {
	if len(args) != 2 || help {
		handleAtomUsage()
		return
	}
	did, pub := handleFeed(args)
	err := lasa.GenerateAtom(context.Background(), client, os.Stdout, did, pub)
	if err != nil {
		panic(err)
	}
}
