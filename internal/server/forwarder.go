package server

import (
	"snwzt/random-video-chat/data/interfaces"
	"sync"
)

type Forwarder struct {
	handlers interfaces.ForwarderHandler
}

func NewForwarder(handlers interfaces.ForwarderHandler) *Forwarder {
	return &Forwarder{
		handlers: handlers,
	}
}

func (f *Forwarder) Run() {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		f.handlers.CreateMatch()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		go f.handlers.DeleteMatch()
	}()

	wg.Wait()
}
