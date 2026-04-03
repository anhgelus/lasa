package lasa

import (
	"context"
	"embed"
	"html"
	"io"
	"reflect"
	"text/template"
	"time"

	site "tangled.org/anhgelus.world/goat-site"
	"tangled.org/anhgelus.world/xrpc"
	"tangled.org/anhgelus.world/xrpc/atproto"
)

//go:embed rss.xml
var rssTemplate embed.FS

type RSSItem struct {
	Title       string
	Link        string
	Description string
	PubDate     string
	Author      string
	Categories  []string
}

type RSSData struct {
	// required
	Title       string
	Link        string
	Description string
	// optional
	LastBuildDate string
	Items         []RSSItem
}

type ErrCannotGenerateRSS struct {
	v string
}

func (err ErrCannotGenerateRSS) Error() string {
	return "cannot generate RSS: " + err.v
}

func GenerateRSS(
	ctx context.Context,
	client xrpc.Client,
	w io.Writer,
	author *atproto.DID,
	pub xrpc.RecordStored[*site.Publication],
) error {
	if pub.Value.Description == nil {
		return ErrCannotGenerateRSS{"description is not set"}
	}
	data := RSSData{
		Title:       html.EscapeString(pub.Value.Name),
		Link:        pub.Value.URL.String(),
		Description: html.EscapeString(*pub.Value.Description),
	}
	items, err := ListDocuments(ctx, client, author, pub.URI)
	if err != nil {
		return err
	}
	doc, err := client.Directory().ResolveDID(ctx, author)
	if err != nil {
		return err
	}
	data.Items = make([]RSSItem, len(items))
	handle, ok := doc.Handle()
	for i, item := range items {
		if i == 0 {
			data.LastBuildDate = item.Value.PublishedAt.Format(time.RFC1123)
		}
		url := pub.Value.URL
		if item.Value.Path == nil {
			return ErrCannotGenerateRSS{"path is not set for " + item.Value.Title}
		}
		url.Path = *item.Value.Path
		for i, v := range item.Value.Tags {
			item.Value.Tags[i] = html.EscapeString(v)
		}
		d := RSSItem{
			Link:       url.String(),
			Title:      html.EscapeString(item.Value.Title),
			PubDate:    item.Value.PublishedAt.Format(time.RFC1123),
			Categories: item.Value.Tags,
		}
		if ok {
			d.Author = "@" + handle.String()
		}
		if item.Value.Description != nil {
			d.Description = html.EscapeString(*item.Value.Description)
		}
		data.Items[i] = d
	}
	return template.Must(template.New("rss").Funcs(map[string]any{
		"isSet": isSet,
	}).ParseFS(rssTemplate, "rss.xml")).ExecuteTemplate(w, "rss.xml", data)
}

func isSet(v any) bool {
	return !reflect.ValueOf(v).IsZero()
}
