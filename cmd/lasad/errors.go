package main

import (
	"errors"
	"net/http"

	"tangled.org/anhgelus.world/xrpc/atproto"
)

func HandleErrors(w http.ResponseWriter, err error) {
	w.Header().Add("Content-Type", "text/plain")
	if atproto.IsErrCannotParse(err) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	} else if errors.Is(err, atproto.ErrHandleNotFound) || errors.Is(err, atproto.ErrDIDNotFound{}) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}
	panic(err)
}
