package kernel

import (
	"gvisor.dev/gvisor/pkg/sentry/arch"
	util "gvisor.dev/gvisor/pkg/sentry/kernel/callbacks"
	"testing"
)

var sigMask2SigNamesWithNoArgs = `
	function cb() {
		hooks.signalMaskToNames()
	}
`

func TestSignalMaskToSignalNamesHook_withNoArgs_Fails(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()
	task := testCreateEmptyTask()
	args := arch.SyscallArguments{}
	cb := JsCallbackBefore{info: util.JsCallbackInfo{
		Sysno:          1,
		CallbackSource: sigMask2SigNamesWithNoArgs,
		CallbackBody:   sigMask2SigNamesWithNoArgs,
		CallbackArgs:   []string{},
		Type:           JsCallbackTypeBefore,
		EntryPoint:     "cb",
	}}

	_, _, err := RunAbstractCallback(&task, jsCallbackInvocationTemplate(&cb), &args, ScriptContextsBuilderOf().Build())
	if err == nil {
		t.Fatalf("no error when calling hook, which needs 1 arg, with no args")
	}
}

var sigMask2SigNamesWithNullArg = `
	function cb() {
		hooks.signalMaskToNames(null)
	}
`

func TestSignalMaskToSignalNamesHook_withNullArg_Fails(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()
	task := testCreateEmptyTask()
	args := arch.SyscallArguments{}
	cb := JsCallbackBefore{info: util.JsCallbackInfo{
		Sysno:          1,
		CallbackSource: sigMask2SigNamesWithNullArg,
		CallbackBody:   sigMask2SigNamesWithNullArg,
		CallbackArgs:   []string{},
		Type:           JsCallbackTypeBefore,
		EntryPoint:     "cb",
	}}

	_, _, err := RunAbstractCallback(&task, jsCallbackInvocationTemplate(&cb), &args, ScriptContextsBuilderOf().Build())
	if err == nil {
		t.Fatalf("no error when calling hook with null arg")
	}
}

var sigMask2SigNamesWithUndefinedArg = `
	function cb() {
		hooks.signalMaskToNames(undefined)
	}
`

func TestSignalMaskToSignalNamesHook_withUndefinedArg_Fails(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()
	task := testCreateEmptyTask()
	args := arch.SyscallArguments{}
	cb := JsCallbackBefore{info: util.JsCallbackInfo{
		Sysno:          1,
		CallbackSource: sigMask2SigNamesWithUndefinedArg,
		CallbackBody:   sigMask2SigNamesWithUndefinedArg,
		CallbackArgs:   []string{},
		Type:           JsCallbackTypeBefore,
		EntryPoint:     "cb",
	}}

	_, _, err := RunAbstractCallback(&task, jsCallbackInvocationTemplate(&cb), &args, ScriptContextsBuilderOf().Build())
	if err == nil {
		t.Fatalf("no error when calling hook with undefined arg")
	}
}

var sigMask2SigNamesWith3Args = `
	function cb() {
		hooks.signalMaskToNames(8, 0, 0)
	}
`

func TestSignalMaskToSignalNamesHook_with3Args_Fails(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()
	task := testCreateEmptyTask()
	args := arch.SyscallArguments{}
	cb := JsCallbackBefore{info: util.JsCallbackInfo{
		Sysno:          1,
		CallbackSource: sigMask2SigNamesWith3Args,
		CallbackBody:   sigMask2SigNamesWith3Args,
		CallbackArgs:   []string{},
		Type:           JsCallbackTypeBefore,
		EntryPoint:     "cb",
	}}

	_, _, err := RunAbstractCallback(&task, jsCallbackInvocationTemplate(&cb), &args, ScriptContextsBuilderOf().Build())
	if err == nil {
		t.Fatalf("no error when calling hook, which needs 1 arg, with 3 args")
	}
}

var sigMask2SigNamesWorks = `
	expectedSignals = ["SIGHUP", "SIGINT", "SIGQUIT"]

	function cb() {
		signals = hooks.signalMaskToNames(7)
		if (signals.length != expectedSignals.length) {
			return {
				"ret": -1,
				"errno": -1,
				"0": expectedSignals.length,
				"1": signals.length				
			}
		}

		for (i = 0; i < expectedSignals.length; i++) {
			if (signals.indexOf(expectedSignals[i]) == -1) {
				return {
					"ret": -1,
					"errno": i
				}
			}
		}

		return {
			"ret": 0,
			"errno": 0
		}
	}
`

func TestSignalMaskToSignalNamesHook_WorksExpected(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()
	task := testCreateEmptyTask()
	args := arch.SyscallArguments{}
	cb := JsCallbackBefore{info: util.JsCallbackInfo{
		Sysno:          1,
		CallbackSource: sigMask2SigNamesWorks,
		CallbackBody:   sigMask2SigNamesWorks,
		CallbackArgs:   []string{},
		Type:           JsCallbackTypeBefore,
		EntryPoint:     "cb",
	}}

	newArgs, rval, err := RunAbstractCallback(&task, jsCallbackInvocationTemplate(&cb), &args, ScriptContextsBuilderOf().Build())
	if err != nil {
		t.Fatalf("error when calling hook")
	}
	if rval == nil {
		t.Fatalf("return value expected")
	}
	if rval.returnValue != 0 {
		if int(rval.errno) < 0 {
			t.Fatalf("amount of got signal names (%v) and expected signal names (%v) do not match", newArgs[1].Value, newArgs[0].Value)
		}
		t.Fatalf("expected signal with index %v not in provided list", int(rval.errno))
	}
}
