package lasa

import (
	"context"
	"errors"
	"strings"

	"tangled.org/anhgelus.world/xrpc/atproto"
)

func Resolve(ctx context.Context, dir atproto.Directory, arg string) (did *atproto.DID, err error) {
	if strings.HasPrefix(arg, "did:") {
		did, err = atproto.ParseDID(arg)
		if err == nil {
			return
		}
	}
	handle, e := atproto.ParseHandle(arg)
	if err == nil {
		err = e
	} else {
		err = errors.Join(err, e)
	}
	if err != nil {
		return
	}
	return handle.DID(ctx, dir)
}
