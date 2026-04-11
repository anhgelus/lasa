package internal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"time"
)

type StatusWriter struct {
	http.ResponseWriter
	Code int
}

func (w *StatusWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.Code = statusCode
}

type Handler func(*StatusWriter, *http.Request)

type Middleware func(Handler, *StatusWriter, *http.Request)

type Mux struct {
	middlewares []func(Handler) Handler
	handler     Handler
}

func NewMux(base *http.ServeMux) *Mux {
	return &Mux{handler: func(w *StatusWriter, r *http.Request) {
		base.ServeHTTP(w, r)
	}}
}

func (m *Mux) Handle() http.Handler {
	slices.Reverse(m.middlewares)
	for _, middle := range m.middlewares {
		m.handler = middle(m.handler)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.handler(&StatusWriter{ResponseWriter: w}, r)
	})
}

func (m *Mux) Use(middlewares ...Middleware) {
	for _, middle := range middlewares {
		m.middlewares = append(m.middlewares, func(next Handler) Handler {
			return func(w *StatusWriter, r *http.Request) {
				middle(next, w, r)
			}
		})
	}
}

func MiddlewareLog(cancelCause func(context.Context) context.CancelCauseFunc, logNotFound, logBadRequest bool) Middleware {
	return func(next Handler, w *StatusWriter, r *http.Request) {
		log := slog.With("uri", r.RequestURI)
		now := time.Now()
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Error("panic!", "error", err, "duration", time.Since(now))
				switch e := err.(type) {
				case error:
					cancelCause(r.Context())(e)
				case string:
					cancelCause(r.Context())(errors.New(e))
				default:
					log.Warn(
						"cannot set cancel cause, because error type is not supported",
						"type", fmt.Sprintf("%T", e),
					)
				}
			}
		}()

		next(w, r)

		log = log.With("status", w.Code, "duration", time.Since(now))
		if w.Code < 400 {
			log.Debug("handled")
		} else if w.Code < 500 {
			level := slog.LevelDebug
			if (w.Code == http.StatusNotFound && logNotFound) ||
				(w.Code == http.StatusBadRequest && logBadRequest) ||
				(w.Code != http.StatusNotFound && w.Code != http.StatusBadRequest) {
				level = slog.LevelWarn
			}
			log.Log(context.Background(), level, "invalid request")
		} else {
			log.Error("error while handling request")
		}
	}
}
