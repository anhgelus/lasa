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

type LimitManyRequests[T any] struct {
	content map[string]*limitRequests[T]
	mu      sync.RWMutex
}

func NewLimitManyRequests[T any]() *LimitManyRequests[T] {
	return &LimitManyRequests[T]{content: make(map[string]*limitRequests[T])}
}

func (lm *LimitManyRequests[T]) Do(key string, fn func() (T, error)) (T, error) {
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

	defer func() {
		lm.mu.Lock()
		delete(lm.content, key)
		lm.mu.Unlock()
	}()

	v, err := fn()
	if err != nil {
		return v, err
	}
	l.Send(v)
	return v, nil
}
