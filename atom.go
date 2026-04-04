package lasa

import (
	"context"
	"embed"
	"io"
	"text/template"

	site "tangled.org/anhgelus.world/goat-site"
	"tangled.org/anhgelus.world/xrpc"
	"tangled.org/anhgelus.world/xrpc/atproto"
)

//go:embed atom.xml
var atomTemplate embed.FS

func GenerateAtom(
	ctx context.Context,
	client xrpc.Client,
	w io.Writer,
	author *atproto.DID,
	pub xrpc.RecordStored[*site.Publication],
) error {
	data, err := genFeedData(ctx, client, author, pub)
	if err != nil {
		return err
	}
	return template.Must(template.New("rss").Funcs(map[string]any{
		"isSet": isSet,
	}).ParseFS(atomTemplate, "atom.xml")).ExecuteTemplate(w, "atom.xml", data)
}
