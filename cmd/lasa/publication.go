package main

import (
	"context"
	"fmt"
	"os"

	site "tangled.org/anhgelus.world/goat-site"
	"tangled.org/anhgelus.world/lasa"
	"tangled.org/anhgelus.world/lasa/cmd/internal"
	"tangled.org/anhgelus.world/xrpc"
	"tangled.org/anhgelus.world/xrpc/atproto"
)

func handlePublicationUsage() {
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
	if !help {
		os.Exit(1)
	}
}

func handlePublication(args []string) {
	if len(args) == 0 {
		handlePublicationUsage()
		return
	}
	did, err := lasa.Resolve(context.Background(), client.Directory(), args[0])
	if err != nil {
		panic(err)
	}
	if len(args) > 1 {
		handlePublicationSpecific(did, args[1:])
		return
	}
	pubs, _, err := xrpc.ListRecords[*site.Publication](context.Background(), client, did, 0, "", false)
	if err != nil {
		panic(err)
	}
	if len(pubs) == 0 {
		fmt.Println("No publication found for", args[0])
		return
	}
	for _, pub := range pubs {
		internal.DisplayPublication(context.Background(), client, did, pub.URI, pub.Value)
		fmt.Println()
	}
}

func handlePublicationSpecific(did *atproto.DID, args []string) {
	if len(args) != 1 {
		handlePublicationUsage()
		return
	}
	rkey, err := atproto.ParseRecordKey(args[0])
	if err != nil {
		return
	}
	pub, err := xrpc.GetRecord[*site.Publication](context.Background(), client, did, rkey, nil)
	if err != nil {
		panic(err)
	}
	internal.DisplayPublication(context.Background(), client, did, pub.URI, pub.Value)
	docs, err := lasa.ListDocuments(context.Background(), client, did, pub.URI)
	if err != nil {
		panic(err)
	}
	for _, doc := range docs {
		internal.DisplayDocument(context.Background(), client, did, doc.URI, pub.Value, doc.Value)
		fmt.Println("-----------------------------")
	}
}
