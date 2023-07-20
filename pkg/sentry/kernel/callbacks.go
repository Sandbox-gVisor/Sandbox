package kernel

import (
	"errors"
	"fmt"
	"github.com/dop251/goja"
	"gvisor.dev/gvisor/pkg/sentry/arch"
	"gvisor.dev/gvisor/pkg/sentry/kernel/callbacks"
	"strings"
	"sync"
)

type SyscallArgumentSubstitutionBeforeExecution struct {
	rval uintptr
	err  error
}

// CallbackBefore - interface which is used to observe and / or modify syscall arguments
type CallbackBefore interface {
	// CallbackBeforeFunc accepts Task, sysno and syscall arguments returns:
	//
	// new args, rval/err if needed, error if something bad occurred
	CallbackBeforeFunc(t *Task, sysno uintptr, args *arch.SyscallArguments) (*arch.SyscallArguments, *SyscallArgumentSubstitutionBeforeExecution, error)
}

type CallbackAfter interface {
	// CallbackAfterFunc accepts Task, sysno, syscall arguments and rval, err after as result of gvisor syscall impl
	//
	// - new args
	//
	// - new rval
	//
	// - new err (should be converted to golang error)
	//
	// - error if something went wrong
	CallbackAfterFunc(t *Task, sysno uintptr, args *arch.SyscallArguments, rval uintptr, err error) (*arch.SyscallArguments, uintptr, uintptr, error)
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

// JsCallbackBefore implements CallbackBefore
type JsCallbackBefore struct {
	source     string
	entryPoint string
	sysno      uintptr
}

type JsCallbackAfter struct {
}

func (cb *JsCallbackBefore) fromDto(dto *callbacks.CallbackDto) error {
	if dto.EntryPoint == "" || dto.CallbackSource == "" || dto.Type != "before" {
		return errors.New("invalid before callback dto")
	}

	cb.sysno = uintptr(dto.Sysno)
	cb.source = dto.CallbackSource
	cb.entryPoint = dto.EntryPoint
	return nil
}

// addSyscallArgsToContextObject from this context object user`s callback will take syscall args
func addSyscallArgsToContextObject(object *goja.Object, arguments *arch.SyscallArguments) error {
	for i, arg := range arguments {
		err := object.Set(fmt.Sprintf("arg%d", i), int64(arg.Value))

		if err != nil {
			return err
		}
	}

	return nil
}

// callbackInvocationTemplate generate string that represent user callback script + invocation of it with injected args
func (cb *JsCallbackBefore) callbackInvocationTemplate() string {
	args := make([]string, len(arch.SyscallArguments{}))
	for i := range args {
		args[i] = fmt.Sprintf("args.arg%d", i)
	}

	return fmt.Sprintf("%s; %s(%s)", cb.source, cb.entryPoint, strings.Join(args, ", "))
}

// CallbackBeforeFunc execution of user callback for syscall on js VM with our hooks
func (cb *JsCallbackBefore) CallbackBeforeFunc(t *Task, _ uintptr, args *arch.SyscallArguments) (*arch.SyscallArguments, error) {
	kernel := t.Kernel()
	kernel.GojaRuntime.Mutex.Lock()
	defer kernel.GojaRuntime.Mutex.Unlock()

	vm := kernel.GojaRuntime.JsVM
	hooksHolder := vm.NewObject()
	if err := kernel.hooksTable.addHooksToContextObject(hooksHolder, t); err != nil {
		return nil, err
	}

	if err := vm.Set("hooks", hooksHolder); err != nil {
		return nil, err
	}

	argsHolder := vm.NewObject()
	if err := addSyscallArgsToContextObject(argsHolder, args); err != nil {
		return nil, err
	}

	if err := vm.Set("args", argsHolder); err != nil {
		return nil, err
	}

	val, err := vm.RunString(cb.callbackInvocationTemplate())
	if err != nil {
		return nil, err
	}

	ret, err_ := callbacks.ExtractArgsFromRetJsValue(args, vm, &val)
	if err_ != nil {
		return nil, err_
	}

	return ret, nil
}
