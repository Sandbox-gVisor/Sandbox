package kernel

import (
	"errors"
	"gvisor.dev/gvisor/pkg/sentry/arch"
	"sync"
)

type Callback interface {
	CallbackFunc(t *Task, sysno uintptr, args *arch.SyscallArguments) (*arch.SyscallArguments, error)
}

type CallbackTable struct {
	data  map[uintptr]Callback
	mutex sync.Mutex
}

func (ct *CallbackTable) registerCallback(sysno uintptr, f *Callback) error {
	if f == nil {
		return errors.New("callback func is nil")
	}
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	ct.data[sysno] = *f
	return nil
}

func (ct *CallbackTable) unregisterCallback(sysno uintptr) error {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	delete(ct.data, sysno)
	return nil
}

func (ct *CallbackTable) getCallback(sysno uintptr) Callback {
	ct.mutex.Lock()

	f, ok := ct.data[sysno]
	ct.mutex.Unlock()
	if ok {
		return f
	} else {
		return nil
	}
}

type SimplePrinter struct {
	counter int
	mu      sync.Mutex
}

func (s *SimplePrinter) CallbackFunc(t *Task, sysno uintptr, args *arch.SyscallArguments) (*arch.SyscallArguments, error) {
	s.mu.Lock()
	val := s.counter
	s.counter += 1
	s.mu.Unlock()
	t.Debugf("Bruh... %v", val)
	return args, nil
}
