package kernel

import (
	"gvisor.dev/gvisor/pkg/sentry/arch"
	util "gvisor.dev/gvisor/pkg/sentry/kernel/callbacks"
	"testing"
)

var signalByNameWithNoArgs = `
	function cb() {
		hooks.nameToSignal()
	}
`

func TestSignalByNameHook_withNoArgs_Fails(t *testing.T) {
	testThatCbFailsWithErr(t, signalByNameWithNoArgs, "no error when calling hook, which needs 1 arg, with no args")
}

var signalByNameWith2Args = `
	function cb() {
		hooks.nameToSignal(1, 2)
	}
`

func TestSignalByNameHook_with3Args_Fails(t *testing.T) {
	testThatCbFailsWithErr(t, signalByNameWith2Args, "no error when calling hook, which needs 1 arg, with 2 args")
}

var signalByNameWithNullArg = `
	function cb() {
		hooks.nameToSignal(null)
	}
`

func TestSignalByNameHook_withNullArg_Fails(t *testing.T) {
	testThatCbFailsWithErr(t, signalByNameWithNullArg, "no error when calling hook with null arg")
}

var signalByNameWithUndefinedArg = `
	function cb() {
		hooks.nameToSignal(undefined)
	}
`

func TestSignalByNameHook_withUndefinedArg_Fails(t *testing.T) {
	testThatCbFailsWithErr(t, signalByNameWithUndefinedArg, "no error when calling hook with undefined arg")
}

var signalByNameWorks = `
	function cb() {
		sigNum = hooks.nameToSignal("SIGINT")
		if (sigNum != 2) {
			return {
				"ret": -1,
				"errno": sigNum
			}
		}

		return {
			"ret": 0,
			"errno": 0
		}
	}
`

func TestSignalByNameHook_Works(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()
	task := testCreateEmptyTask()
	args := arch.SyscallArguments{}
	cb := JsCallbackBefore{info: util.JsCallbackInfo{
		Sysno:          1,
		CallbackSource: signalByNameWorks,
		CallbackBody:   signalByNameWorks,
		CallbackArgs:   []string{},
		Type:           JsCallbackTypeBefore,
		EntryPoint:     "cb",
	}}

	_, rval, err := RunAbstractCallback(&task, jsCallbackInvocationTemplate(&cb), &args, ScriptContextsBuilderOf().Build())
	if err != nil {
		t.Fatalf("unexpected error while running callback")
	}
	if rval == nil {
		t.Fatalf("unexpected nil return value")
	}
	if int(rval.returnValue) != 0 {
		t.Fatalf("bad hook return value: got %v, expected 2", int(rval.errno))
	}
}
