package kernel

import (
	"errors"
	"gvisor.dev/gvisor/pkg/sentry/arch"
	"sync"
)

type CallbackFunc interface {
	Callback(t *Task, sysno uintptr, args *arch.SyscallArguments) (*arch.SyscallArguments, error)
}

type CallbackTable struct {
	data  map[uintptr]CallbackFunc
	mutex sync.Mutex
}

func (ct *CallbackTable) registerCallback(sysno uintptr, f *CallbackFunc) error {
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

func (ct *CallbackTable) getCallback(sysno uintptr) CallbackFunc {
	ct.mutex.Lock()

	f, ok := ct.data[sysno]
	ct.mutex.Unlock()
	if ok {
		return f
	} else {
		return nil
	}
}
