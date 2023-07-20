package kernel

import (
	"errors"
	"gvisor.dev/gvisor/pkg/sentry/arch"
	"sync"
)

// CallbackBefore - interface which is used to observe and / or modify syscall arguments
type CallbackBefore interface {
	// CallbackFunc accepts Task, sysno and syscall arguments
	// returns:
	// - new syscall arguments
	// - error
	CallbackFunc(t *Task, sysno uintptr, args *arch.SyscallArguments) (*arch.SyscallArguments, error)
}

type CallbackTable struct {
	// callbackBefore is a map of:
	// key - sysno
	// val - CallbackBefore
	callbackBefore map[uintptr]CallbackBefore

	// mutexBefore is sync.Mutex used to sync callbackBefore
	mutexBefore sync.Mutex
}

func (ct *CallbackTable) registerCallbackBefore(sysno uintptr, f CallbackBefore) error {
	if f == nil {
		return errors.New("callback func is nil")
	}
	ct.mutexBefore.Lock()
	defer ct.mutexBefore.Unlock()

	ct.callbackBefore[sysno] = f
	return nil
}

func (ct *CallbackTable) registerCallbackBeforeNoLock(sysno uintptr, f CallbackBefore) error {
	if f == nil {
		return errors.New("callback func is nil")
	}

	ct.callbackBefore[sysno] = f
	return nil
}

func (ct *CallbackTable) unregisterCallbackBefore(sysno uintptr) error {
	ct.mutexBefore.Lock()
	defer ct.mutexBefore.Unlock()

	delete(ct.callbackBefore, sysno)
	return nil
}

func (ct *CallbackTable) getCallbackBefore(sysno uintptr) CallbackBefore {
	ct.mutexBefore.Lock()

	f, ok := ct.callbackBefore[sysno]
	ct.mutexBefore.Unlock()
	if ok && f != nil {
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
	t.Debugf("sysno %v: Bruh... %v", sysno, val)
	return args, nil
}
