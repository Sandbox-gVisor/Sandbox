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

func (ct *CallbackTable) registerCallback(sysno uintptr, f Callback) error {
	if f == nil {
		return errors.New("callback func is nil")
	}
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	ct.data[sysno] = f
	return nil
}

func (ct *CallbackTable) registerCallbackWithoutLock(sysno uintptr, f Callback) error {
	if f == nil {
		return errors.New("callback func is nil")
	}

	ct.data[sysno] = f
	return nil
}

func (ct *CallbackTable) registerAllFromCollector(cc *CallbackCollector) {
	if cc == nil {
		return
	}
	pairs := cc.getAll()
	ct.mutex.Lock()
	defer ct.mutex.Unlock()
	for i := 0; i < len(pairs); i += 1 {
		ct.data[pairs[i].sysno] = pairs[i].callback
	}
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
	if ok && f != nil {
		return f
	} else {
		return nil
	}
}

type CallbackPair struct {
	sysno    uintptr
	callback Callback
}

type CallbackCollector struct {
	collectedCallbacks []CallbackPair
}

func (cc *CallbackCollector) collect(sysno uintptr, callback Callback) {
	cc.collectedCallbacks = append(cc.collectedCallbacks, CallbackPair{sysno, callback})
}

func (cc *CallbackCollector) getAll() []CallbackPair {
	return cc.collectedCallbacks
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
