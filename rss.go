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

type FeedItem struct {
	Title       string
	Link        string
	Description string
	PubDate     string
	UpdatedDate string
	Author      string
	Categories  []string
}

type FeedData struct {
	// required
	Title       string
	Link        string
	Description string
	// optional
	LastBuildDate string
	Items         []FeedItem
	Author        string
}

type ErrCannotGenerateFeed struct {
	v string
}

func (err ErrCannotGenerateFeed) Error() string {
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
		return ErrCannotGenerateFeed{"description is not set"}
	}
	data, err := genFeedData(ctx, client, author, pub)
	if err != nil {
		return err
	}
	return template.Must(template.New("rss").Funcs(map[string]any{
		"isSet": IsSet,
	}).ParseFS(rssTemplate, "rss.xml")).ExecuteTemplate(w, "rss.xml", data)
}

func genFeedData(
	ctx context.Context,
	client xrpc.Client,
	author *atproto.DID,
	pub xrpc.RecordStored[*site.Publication],
) (FeedData, error) {
	data := FeedData{
		Title: html.EscapeString(pub.Value.Name),
		Link:  pub.Value.URL.String(),
	}
	if pub.Value.Description != nil {
		data.Description = *pub.Value.Description
	}
	items, err := ListDocuments(ctx, client, author, pub.URI)
	if err != nil {
		return data, err
	}
	doc, err := client.Directory().ResolveDID(ctx, author)
	if err != nil {
		return data, err
	}
	handle, ok := doc.Handle()
	if ok {
		data.Author = handle.String()
	}
	data.Items = make([]FeedItem, len(items))
	for i, item := range items {
		url := pub.Value.URL
		if item.Value.Path == nil {
			return data, ErrCannotGenerateFeed{"path is not set for " + item.Value.Title}
		}
		url.Path = *item.Value.Path
		for i, v := range item.Value.Tags {
			item.Value.Tags[i] = html.EscapeString(v)
		}
		d := FeedItem{
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
		if item.Value.UpdatedAt != nil {
			d.UpdatedDate = item.Value.UpdatedAt.Format(time.RFC1123)
		} else {
			d.UpdatedDate = d.PubDate
		}
		if i == 0 {
			data.LastBuildDate = d.UpdatedDate
		}
		data.Items[i] = d
	}
	return data, nil
}

func IsSet(v any) bool {
	return !reflect.ValueOf(v).IsZero()
}
