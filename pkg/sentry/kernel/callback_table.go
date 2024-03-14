package kernel

import (
	"errors"
	"fmt"
	"sync"
)

// CallbackTable is a storage of functions which can be called before and/ or after syscall execution
// TODO incapsulate the mutex (exposing mutex - straight way to deadlock or other memes)
type CallbackTable struct {
	// callbackBefore is a map of:
	//	key - sysno (uintptr)
	//	val - CallbackBefore
	callbackBefore map[uintptr]CallbackBefore

	// rwLockBefore is sync.Mutex used to sync callbackBefore
	rwLockBefore sync.RWMutex

	// callbackAfter is a map of:
	//	key - sysno (uintptr)
	//	val - CallbackAfter
	callbackAfter map[uintptr]CallbackAfter

	// rwLockAfter is sync.Mutex used to sync callbackAfter
	rwLockAfter sync.RWMutex
}

func (ct *CallbackTable) registerCallbackBefore(sysno uintptr, f CallbackBefore) error {
	if f == nil {
		return errors.New("callback func is nil")
	}
	ct.rwLockBefore.Lock()
	defer ct.rwLockBefore.Unlock()

	ct.callbackBefore[sysno] = f
	return nil
}

func (ct *CallbackTable) registerCallbackAfter(sysno uintptr, f CallbackAfter) error {
	if f == nil {
		return errors.New("callback func is nil")
	}
	ct.rwLockAfter.Lock()
	defer ct.rwLockAfter.Unlock()

	ct.callbackAfter[sysno] = f
	return nil
}

func (ct *CallbackTable) UnregisterAll() {
	ct.rwLockBefore.Lock()
	ct.rwLockAfter.Lock()

	defer ct.rwLockAfter.Unlock()
	defer ct.rwLockBefore.Unlock()

	ct.callbackAfter = map[uintptr]CallbackAfter{}
	ct.callbackBefore = map[uintptr]CallbackBefore{}
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
	ct.rwLockBefore.Lock()
	defer ct.rwLockBefore.Unlock()

	_, ok := ct.callbackBefore[sysno]
	if !ok {
		return errors.New(fmt.Sprintf("before-callback with sysno %v not exist", sysno))
	}

	delete(ct.callbackBefore, sysno)
	return nil
}

func (ct *CallbackTable) unregisterCallbackAfter(sysno uintptr) error {
	ct.rwLockAfter.Lock()
	defer ct.rwLockAfter.Unlock()

	_, ok := ct.callbackAfter[sysno]
	if !ok {
		return errors.New(fmt.Sprintf("after-callback with sysno %v not exist", sysno))
	}

	delete(ct.callbackAfter, sysno)
	return nil
}

func (ct *CallbackTable) getCallbackBefore(sysno uintptr) CallbackBefore {
	ct.rwLockBefore.RLock()

	f, ok := ct.callbackBefore[sysno]
	ct.rwLockBefore.RUnlock()
	if ok && f != nil {
		return f
	} else {
		return nil
	}
}

func (ct *CallbackTable) getCallbackAfter(sysno uintptr) CallbackAfter {
	ct.rwLockAfter.RLock()

	f, ok := ct.callbackAfter[sysno]
	ct.rwLockAfter.RUnlock()
	if ok && f != nil {
		return f
	} else {
		return nil
	}
}
