package lasa

import (
	"sync"
	"sync/atomic"
)

type limitRequests[T any] struct {
	n  atomic.Uint32
	ch chan T
}

func newLimitRequests[T any]() *limitRequests[T] {
	return &limitRequests[T]{ch: make(chan T)}
}

func (l *limitRequests[T]) Sub() T {
	l.n.Add(1)
	return <-l.ch
}

func (l *limitRequests[T]) Send(v T) {
	for range l.n.Load() {
		l.ch <- v
	}
}

type limitManyRequests[T any] struct {
	content map[string]*limitRequests[T]
	mu      sync.RWMutex
}

func newLimitManyRequests[T any]() *limitManyRequests[T] {
	return &limitManyRequests[T]{content: make(map[string]*limitRequests[T])}
}

func (lm *limitManyRequests[T]) Do(key string, fn func() (T, error)) (T, error) {
	lm.mu.RLock()
	l, ok := lm.content[key]
	if ok {
		lm.mu.RUnlock()
		return l.Sub(), nil
	}
	lm.mu.RUnlock()

	l = newLimitRequests[T]()
	lm.mu.Lock()
	lm.content[key] = l
	lm.mu.Unlock()

	v, err := fn()
	if err != nil {
		return v, err
	}
	l.Send(v)
	return v, nil
}
