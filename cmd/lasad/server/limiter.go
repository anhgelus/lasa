package server

import (
	"log/slog"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"
)

type limited struct {
	time  time.Time
	times uint
}

type Limiter struct {
	mu      sync.RWMutex
	limited map[string]*limited
}

func exp(times, level uint) time.Duration {
	if int(times) < 4 {
		return 0
	}
	// max at 10800 seconds (3 hours)
	return min(time.Duration(math.Pow(2, float64(times+level))), 10800) * time.Second
}

func (l *Limiter) isLimited(r *http.Request) bool {
	ip, _, _ := strings.Cut(r.RemoteAddr, ":")
	l.mu.RLock()
	tm, ok := l.limited[ip]
	l.mu.RUnlock()
	return ok && tm.time.Unix() > time.Now().Unix()
}

func (l *Limiter) handle(w *statusWriter, r *http.Request, log *slog.Logger) {
	ip, _, _ := strings.Cut(r.RemoteAddr, ":")
	if w.code != http.StatusNotFound && w.code != http.StatusBadRequest && w.code != http.StatusTooManyRequests {
		return
	}
	var level uint = 0
	if w.code == http.StatusBadRequest {
		level = 1
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	tm, ok := l.limited[ip]
	if !ok {
		l.limited[ip] = &limited{times: 1}
		return
	}
	tm.times++
	dur := exp(tm.times, level)
	if dur == 0 {
		return
	}
	log.Info("rate limiting", "ip", ip, "for", dur)
	tm.time = time.Now().Add(dur)

	go func(l *Limiter, dur time.Duration, tm *limited, ip string) {
		time.Sleep(dur)
		l.mu.Lock()
		defer l.mu.Unlock()
		if time.Now().Unix() < tm.time.Unix() {
			return
		}
		delete(l.limited, ip)
	}(l, dur, tm, ip)
}
