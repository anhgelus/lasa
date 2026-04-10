package server

import (
	"errors"
	"net/http"

	"tangled.org/anhgelus.world/xrpc"
	"tangled.org/anhgelus.world/xrpc/atproto"
)

func HandleErrors(w http.ResponseWriter, err error) {
	var status int
	if atproto.IsErrCannotParse(err) {
		status = http.StatusBadRequest
	} else if errors.Is(err, atproto.ErrHandleNotFound) {
		status = http.StatusNotFound
	} else if _, ok := errors.AsType[atproto.ErrDIDNotFound](err); ok {
		status = http.StatusNotFound
	} else if errors.Is(err, xrpc.ErrRecordNotFound) {
		status = http.StatusNotFound
	}
	if status > 0 {
		http.Error(w, err.Error(), status)
		return
	}
	panic(err)

}
