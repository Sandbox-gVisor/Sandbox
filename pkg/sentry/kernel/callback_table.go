package kernel

import (
	"errors"
	"gvisor.dev/gvisor/pkg/sentry/arch"
	"sync"
)

// CallbackBefore - interface which is used to observe and / or modify syscall arguments
type CallbackBefore interface {
	// CallbackBeforeFunc accepts Task, sysno and syscall arguments returns:
	//
	// TODO: return new args, new rval, new err, instead, error
	CallbackBeforeFunc(t *Task, sysno uintptr, args *arch.SyscallArguments) (*arch.SyscallArguments, error)
}

type CallbackAfter interface {
	// CallbackAfterFunc accepts Task, sysno and syscall arguments
	//
	// - new args
	//
	// - new rval
	//
	// - new err (should be converted to golang error)
	//
	// - error if something went wrong
	CallbackAfterFunc(t *Task, sysno uintptr, args *arch.SyscallArguments) (*arch.SyscallArguments, uintptr, uintptr, error)
}

type CallbackTable struct {
	// callbackBefore is a map of:
	//
	// key - sysno (uintptr)
	//
	// val - CallbackBefore
	callbackBefore map[uintptr]CallbackBefore

	// mutexBefore is sync.Mutex used to sync callbackBefore
	mutexBefore sync.Mutex

	// callbackAfter is a map of:
	//
	// key - sysno (uintptr)
	//
	// val - CallbackAfter
	callbackAfter map[uintptr]CallbackAfter

	// mutexAfter is sync.Mutex used to sync callbackAfter
	mutexAfter sync.Mutex
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

func (ct *CallbackTable) registerCallbackAfter(sysno uintptr, f CallbackAfter) error {
	if f == nil {
		return errors.New("callback func is nil")
	}
	ct.mutexAfter.Lock()
	defer ct.mutexAfter.Unlock()

	ct.callbackAfter[sysno] = f
	return nil
}

func (ct *CallbackTable) registerCallbackBeforeNoLock(sysno uintptr, f CallbackBefore) error {
	if f == nil {
		return errors.New("callback func is nil")
	}

	ct.callbackBefore[sysno] = f
	return nil
}

func (ct *CallbackTable) registerCallbackAfterNoLock(sysno uintptr, f CallbackAfter) error {
	if f == nil {
		return errors.New("callback func is nil")
	}

	ct.callbackAfter[sysno] = f
	return nil
}

func (ct *CallbackTable) unregisterCallbackBefore(sysno uintptr) error {
	ct.mutexBefore.Lock()
	defer ct.mutexBefore.Unlock()

	delete(ct.callbackBefore, sysno)
	return nil
}

func (ct *CallbackTable) unregisterCallbackAfter(sysno uintptr) error {
	ct.mutexAfter.Lock()
	defer ct.mutexAfter.Unlock()

	delete(ct.callbackAfter, sysno)
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

func (ct *CallbackTable) getCallbackAfter(sysno uintptr) CallbackAfter {
	ct.mutexAfter.Lock()

	f, ok := ct.callbackAfter[sysno]
	ct.mutexAfter.Unlock()
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

func (s *SimplePrinter) CallbackBeforeFunc(t *Task, sysno uintptr, args *arch.SyscallArguments) (*arch.SyscallArguments, error) {
	s.mu.Lock()
	val := s.counter
	s.counter += 1
	s.mu.Unlock()
	t.Debugf("sysno %v: Bruh... %v", sysno, val)
	return args, nil
}
