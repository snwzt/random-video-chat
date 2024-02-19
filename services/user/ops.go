package user

import (
	"snwzt/rvc/data/interfaces"
	"sync"
)

type UserOperations struct {
	handlers interfaces.UserOperationsHandler
}

func NewUserOperations(handlers interfaces.UserOperationsHandler) *UserOperations {
	return &UserOperations{
		handlers: handlers,
	}
}

func (svc *UserOperations) Run() {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		svc.handlers.UserMatcher()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		svc.handlers.UserRemove()
	}()

	wg.Wait()
}
