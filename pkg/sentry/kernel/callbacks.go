package kernel

import (
	"fmt"
	"github.com/dop251/goja"
	"gvisor.dev/gvisor/pkg/sentry/arch"
	"gvisor.dev/gvisor/pkg/sentry/kernel/callbacks"
)

// CallbackBefore - interface which is used to observe and / or modify syscall arguments
type CallbackBefore interface {
	// CallbackBeforeFunc accepts:
	//	- Task
	//	- sysno
	//	- syscall arguments
	//
	// returns:
	//	- new args
	//	- SyscallReturnValue
	//	- error if something bad occurred
	CallbackBeforeFunc(t *Task, sysno uintptr, args *arch.SyscallArguments) (*arch.SyscallArguments, *SyscallReturnValue, error)

	// Info about this callback
	Info() callbacks.JsCallbackInfo
}

// CallbackAfter - interface which is used to replace args / return value / errno of syscall
type CallbackAfter interface {
	// CallbackAfterFunc accepts:
	//	- Task
	//	- sysno
	//	- syscall arguments
	//	- returnValue (return value of gvisor syscall impl)
	//	- err (err of gvisor syscall impl)
	//
	// returns
	//	- new args
	//	- SyscallReturnValue
	//	- error if something went wrong
	CallbackAfterFunc(t *Task, sysno uintptr, args *arch.SyscallArguments,
		ret uintptr, err error) (*arch.SyscallArguments, *SyscallReturnValue, error)

	// Info about this callback
	Info() callbacks.JsCallbackInfo
}

type SyscallReturnValue struct {
	returnValue uintptr
	errno       uintptr
}

func (s SyscallReturnValue) addSelfToContextObject(object *goja.Object) error {
	err := object.Set(JsSyscallReturnValue, int64(s.returnValue))
	if err != nil {
		return err
	}

	err = object.Set(JsSyscallErrno, int64(s.errno))
	if err != nil {
		return err
	}

	return nil
}

type SyscallReturnValueWithError struct {
	returnValue uintptr
	errno       error
}

func (s SyscallReturnValueWithError) addSelfToContextObject(object *goja.Object) error {
	err := object.Set(JsSyscallReturnValue, int64(s.returnValue))
	if err != nil {
		return err
	}

	err = object.Set(JsSyscallErrno, s.errno)
	if err != nil {
		return err
	}

	return nil
}

type SyscallArgsAddableAdapter struct {
	Args *arch.SyscallArguments
}

func (s *SyscallArgsAddableAdapter) addSelfToContextObject(object *goja.Object) error {
	return addSyscallArgsToContextObject(object, s.Args)
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
