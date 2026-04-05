package lasa

import (
	"context"
	"errors"
	"strings"

	"tangled.org/anhgelus.world/xrpc/atproto"
)

func Resolve(ctx context.Context, dir atproto.Directory, arg string) (*atproto.DID, error) {
	var err error
	if strings.HasPrefix(arg, "did:") {
		var did *atproto.DID
		did, err = atproto.ParseDID(arg)
		if err == nil {
			return did, nil
		}
	}
	handle, e := atproto.ParseHandle(arg)
	if err == nil {
		err = e
	} else {
		err = errors.Join(err, e)
	}
	if err != nil {
		return nil, err
	}
	doc, err := dir.ResolveHandle(ctx, handle)
	if err != nil {
		return nil, err
	}
	return doc.DID, nil
}
