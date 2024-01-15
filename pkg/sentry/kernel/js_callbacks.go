package kernel

import (
	"errors"
	"fmt"
	"gvisor.dev/gvisor/pkg/sentry/arch"
	"gvisor.dev/gvisor/pkg/sentry/kernel/callbacks"
)

type JsCallback interface {
	callbackInfo() *callbacks.JsCallbackInfo

	registerAtCallbackTable(ct *CallbackTable) error
}

// JsCallbackByInfo returns suitable JsCallback (JsCallbackAfter or JsCallbackBefore)
// according to callbacks.JsCallbackInfo
func JsCallbackByInfo(info callbacks.JsCallbackInfo) (JsCallback, error) {
	if info.Type == JsCallbackTypeAfter {
		cb := &JsCallbackAfter{info: info}
		return cb, checkJsCallback(cb)
	}
	if info.Type == JsCallbackTypeBefore {
		cb := &JsCallbackBefore{info: info}
		return cb, checkJsCallback(cb)
	}

	return nil, errors.New("incorrect callback type " + info.Type)
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

// JsCallbackBefore implements CallbackBefore and JsCallback
type JsCallbackBefore struct {
	info callbacks.JsCallbackInfo
}

func (cb *JsCallbackBefore) callbackInfo() *callbacks.JsCallbackInfo {
	return &cb.info
}

func (cb *JsCallbackBefore) Info() callbacks.JsCallbackInfo {
	return cb.info
}

func (cb *JsCallbackBefore) registerAtCallbackTable(ct *CallbackTable) error {
	return ct.registerCallbackBefore(uintptr(cb.info.Sysno), cb)
}

// CallbackBeforeFunc execution of user callback for syscall on js VM with our DependentHooks
func (cb *JsCallbackBefore) CallbackBeforeFunc(t *Task, _ uintptr,
	args *arch.SyscallArguments) (*arch.SyscallArguments, *SyscallReturnValue, error) {

	return RunAbstractCallback(t, jsCallbackInvocationTemplate(cb), args, ScriptContextsBuilderOf().Build())
}

// JsCallbackAfter implements CallbackAfter and JsCallback
type JsCallbackAfter struct {
	info callbacks.JsCallbackInfo
}

func (cb *JsCallbackAfter) callbackInfo() *callbacks.JsCallbackInfo {
	return &cb.info
}

func (cb *JsCallbackAfter) Info() callbacks.JsCallbackInfo {
	return cb.info
}

func (cb *JsCallbackAfter) registerAtCallbackTable(ct *CallbackTable) error {
	return ct.registerCallbackAfter(uintptr(cb.info.Sysno), cb)
}

func (cb *JsCallbackAfter) CallbackAfterFunc(t *Task, _ uintptr, args *arch.SyscallArguments,
	ret uintptr, inputErr error) (*arch.SyscallArguments, *SyscallReturnValue, error) {

	context := ScriptContextsBuilderOf().AddContext3(ArgsJsName,
		SyscallReturnValueWithError{returnValue: ret, errno: inputErr}).Build()

	return RunAbstractCallback(t, jsCallbackInvocationTemplate(cb), args, context)
}
