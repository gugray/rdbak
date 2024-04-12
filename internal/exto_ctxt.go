package internal

import (
	"context"
	"sync"
	"time"
)

type ExtensibleTimeoutContext interface {
	Context() context.Context
	Extend()
	Cancel()
}

type exToCo struct {
	ctxt           context.Context
	cancel         context.CancelFunc
	tickerDone     chan bool
	timeoutSeconds int
	deadline       time.Time
	isCanceled     bool
	mu             sync.Mutex
}

func NewExtensibleTimeoutContext(timeoutSeconds int) ExtensibleTimeoutContext {

	etc := exToCo{
		tickerDone:     make(chan bool),
		timeoutSeconds: timeoutSeconds,
		deadline:       time.Now().Add(time.Duration(timeoutSeconds) * time.Second),
	}
	etc.ctxt, etc.cancel = context.WithCancel(context.Background())

	ticker := time.NewTicker(500 * time.Millisecond)
	go func() {
		for {
			select {
			case <-etc.tickerDone:
				return
			case t := <-ticker.C:
				var isExpired bool
				etc.mu.Lock()
				isExpired = t.After(etc.deadline)
				etc.mu.Unlock()
				if isExpired {
					etc.Cancel()
				}
			}
		}
	}()

	return &etc
}

func (etc *exToCo) Context() context.Context {
	return etc.ctxt
}

func (etc *exToCo) Cancel() {
	etc.mu.Lock()
	if etc.isCanceled {
		etc.mu.Unlock()
		return
	}
	go func() {
		etc.tickerDone <- true
	}()
	etc.isCanceled = true
	etc.mu.Unlock()
	etc.cancel()
}

func (etc *exToCo) Extend() {
	etc.mu.Lock()
	etc.deadline = time.Now().Add(time.Duration(etc.timeoutSeconds) * time.Second)
	etc.mu.Unlock()
}
