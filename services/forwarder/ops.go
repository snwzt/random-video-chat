package forwarder

import (
	"snwzt/rvc/data/interfaces"
	"sync"
)

type ForwarderOperations struct {
	handlers interfaces.ForwarderOperationsHandler
}

func NewForwarder(handlers interfaces.ForwarderOperationsHandler) *ForwarderOperations {
	return &ForwarderOperations{
		handlers: handlers,
	}
}

func (svc *ForwarderOperations) Run() {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		svc.handlers.CreateMatch()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		go svc.handlers.DeleteMatch()
	}()

	wg.Wait()
}
