package kernel

import (
	"errors"
	"fmt"
	"github.com/dop251/goja"
	"gvisor.dev/gvisor/pkg/sentry/arch"
	"gvisor.dev/gvisor/pkg/sentry/kernel/callbacks"
	"strconv"
	"strings"
	"sync"
)

type SyscallReturnValueSubstitution struct {
	returnValue uintptr
	errno       uintptr
}

// CallbackBefore - interface which is used to observe and / or modify syscall arguments
type CallbackBefore interface {
	// CallbackBeforeFunc accepts Task, sysno and syscall arguments returns:
	//
	// new args, returnValue/err if needed, error if something bad occurred
	CallbackBeforeFunc(t *Task, sysno uintptr, args *arch.SyscallArguments) (*arch.SyscallArguments, *SyscallReturnValueSubstitution, error)
}

// CallbackAfter - interface which is used to replace args / return value / errno of syscall
type CallbackAfter interface {
	// CallbackAfterFunc accepts Task, sysno, syscall arguments and returnValue, err after as result of gvisor syscall impl
	//
	// - new args
	//
	// - new returnValue
	//
	// - new err (should be converted to golang error)
	//
	// - error if something went wrong
	CallbackAfterFunc(t *Task, sysno uintptr, args *arch.SyscallArguments,
		substitution *SyscallReturnValueSubstitution) (*arch.SyscallArguments, *SyscallReturnValueSubstitution, error)
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

// js callback staff bellow

type JsCallback interface {
	callbackInfo() *callbacks.JsCallbackInfo
}

const JsCallbackTypeAfter = "after"

const JsCallbackTypeBefore = "before"

const HooksJsName = "hooks"

const ArgsJsName = "args"

// JsCallbackBefore implements CallbackBefore and JsCallback
type JsCallbackBefore struct {
	info callbacks.JsCallbackInfo
}

// JsCallbackAfter implements CallbackAfter and JsCallback
type JsCallbackAfter struct {
	info callbacks.JsCallbackInfo
}

func (cb *JsCallbackBefore) callbackInfo() *callbacks.JsCallbackInfo {
	return &cb.info
}

func (cb *JsCallbackAfter) callbackInfo() *callbacks.JsCallbackInfo {
	return &cb.info
}

func checkJsCallback(cb JsCallback) error {
	info := cb.callbackInfo()
	if info.CallbackSource == "" {
		return errors.New("js callback source is empty")
	}
	if info.EntryPoint == "" {
		return errors.New("js callback entry point is empty")
	}
	if info.Type != JsCallbackTypeBefore && info.Type != JsCallbackTypeAfter {
		return errors.New(fmt.Sprintf("incorrect js callback type: %s", info.Type))
	}

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
func jsCallbackInvocationTemplate(jsCallback JsCallback) string {
	info := jsCallback.callbackInfo()
	args := make([]string, len(arch.SyscallArguments{}))
	for i := range args {
		args[i] = fmt.Sprintf("%s.arg%d", ArgsJsName, i)
	}

	return fmt.Sprintf("%s; %s(%s)", info.CallbackSource, info.EntryPoint, strings.Join(args, ", "))
}

func extractArgsFromRetJsValue(
	inputArgs *arch.SyscallArguments, vm *goja.Runtime, value *goja.Value) (retArgs *arch.SyscallArguments, err error) {

	retArgs = &arch.SyscallArguments{}
	*retArgs = *inputArgs
	retObj := (*value).ToObject(vm)

	for _, key := range retObj.Keys() {
		var ind int
		ind, err = strconv.Atoi(key)
		if err != nil {
			continue
		}

		if ind < 0 || len(inputArgs) < ind {
			err = errors.New("invalid index of ret args")
			return
		}

		ptrVal := retObj.Get(key)
		var ptr uintptr
		ptr, err = callbacks.ExtractPtrFromValue(vm, ptrVal)
		if err != nil {
			return
		}
		retArgs[ind].Value = ptr
	}

	return retArgs, nil
}

func extractSubstitutionFromRetJsValue() {

}

// CallbackBeforeFunc execution of user callback for syscall on js VM with our hooks
func (cb *JsCallbackBefore) CallbackBeforeFunc(t *Task, sysno uintptr, args *arch.SyscallArguments) (*arch.SyscallArguments, *SyscallReturnValueSubstitution, error) {
	kernel := t.Kernel()
	kernel.GojaRuntime.Mutex.Lock()
	defer kernel.GojaRuntime.Mutex.Unlock()

	vm := kernel.GojaRuntime.JsVM
	hooksHolder := vm.NewObject()
	if err := kernel.hooksTable.addHooksToContextObject(hooksHolder, t); err != nil {
		return nil, nil, err
	}

	if err := vm.Set(HooksJsName, hooksHolder); err != nil {
		return nil, nil, err
	}

	argsHolder := vm.NewObject()
	if err := addSyscallArgsToContextObject(argsHolder, args); err != nil {
		return nil, nil, err
	}

	if err := vm.Set(ArgsJsName, argsHolder); err != nil {
		return nil, nil, err
	}

	val, err := vm.RunString(jsCallbackInvocationTemplate(cb))
	if err != nil {
		return nil, nil, err
	}

	retArgs, err_ := extractArgsFromRetJsValue(args, vm, &val)
	if err_ != nil {
		return nil, nil, err_
	}

	return retArgs, nil
}

func (cb *JsCallbackAfter) CallbackAfterFunc(t *Task, sysno uintptr, args *arch.SyscallArguments, rval uintptr, err error) (*arch.SyscallArguments, uintptr, uintptr, error) {
	//TODO implement me
	panic("implement me")
}
