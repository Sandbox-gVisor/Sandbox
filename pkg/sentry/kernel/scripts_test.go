package kernel

import (
	"gvisor.dev/gvisor/pkg/sentry/arch"
	"gvisor.dev/gvisor/pkg/sentry/kernel/callbacks"
	"testing"
)

func testCreateMockTask() *Task {
	return &Task{}
}

func testInitJsRuntime() {
	jsRuntime = initJsRuntime()
}

func testDestroyJsRuntime() {
	jsRuntime = nil
}

var simpleScript = `
	function testF() {
		a = 0
		for (i = 0; i < 10; i++) {
			a = a + 1
		}
		return a
	}

	testF()
	`

func TestRunJsScript_RunsAndReturnsValue(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()

	jsVal, err := RunJsScript(jsRuntime.JsVM, simpleScript, []ScriptContext{})
	if err != nil {
		t.Fatalf("failed to execute script with err %s", err)
	}
	var val int64
	err = jsRuntime.JsVM.ExportTo(jsVal, &val)
	if err != nil {
		t.Fatalf("failed to convert return value %s", err)
	}
	if val != 10 {
		t.Fatalf("Wrong value. Got %v, expected 10", val)
	}
}

var simpleBadScript = `
	function testF() 
		a = 0
		for (i = 0; i < 10; i++) {
			a = a + 1
		}
		return a
	}

	testF()
	`

func TestRunJsScript_failsWithBadScript(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()

	_, err := RunJsScript(jsRuntime.JsVM, simpleBadScript, []ScriptContext{})
	if err == nil {
		t.Fatalf("unexpeted succes in executing incorrect script")
	}
}

var cbGetExpectedArgsChangeArgs = `
	function cb() {
		p = hooks.print
		arg = args.arg0
		localStorage = persistence.local
		globalStorage = persistence.glb

		return {
			"ret": 1,
			"errno": 1
		}
	}
	`

func testRunAbstractCallbackRunsAndReturnsSyscallRetValue(t *testing.T, cb JsCallback) {
	task := testCreateMockTask()
	args := arch.SyscallArguments{}
	newArgs, rval, err := RunAbstractCallback(
		task,
		jsCallbackInvocationTemplate(cb),
		&args,
		ScriptContextsBuilderOf().Build())
	if err != nil {
		t.Fatalf("failed to execute callback: %s err: %s", cb.callbackInfo().ToString(), err)
	}
	if len(newArgs) != len(args) {
		t.Fatalf("arguments count doesn't match")
	}
	if rval == nil {
		t.Fatalf("unexpected return value")
	}
	if rval.returnValue != 1 || rval.errno != 1 {
		t.Fatalf("bad return value. Has: returnValue = %v, errno = %v; Expected: returnValue = 1, errno = 1",
			rval.returnValue, rval.errno)
	}
}

func TestRunAbstractCallback_WithJsCallbackBefore_runsAndReturnsNewSyscallValue(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()

	cb := JsCallbackBefore{
		info: callbacks.JsCallbackInfo{
			Sysno:          1,
			EntryPoint:     "cb",
			CallbackSource: cbGetExpectedArgsChangeArgs,
			CallbackBody:   cbGetExpectedArgsChangeArgs,
			CallbackArgs:   []string{"count"},
			Type:           JsCallbackTypeBefore,
		},
	}

	testRunAbstractCallbackRunsAndReturnsSyscallRetValue(t, &cb)
}

func TestRunAbstractCallback_WithJsCallbackAfter_runsAndReturnsNewSyscallValue(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()

	cb := JsCallbackAfter{
		info: callbacks.JsCallbackInfo{
			Sysno:          1,
			EntryPoint:     "cb",
			CallbackSource: cbGetExpectedArgsChangeArgs,
			CallbackBody:   cbGetExpectedArgsChangeArgs,
			CallbackArgs:   []string{"count"},
			Type:           JsCallbackTypeAfter,
		},
	}

	testRunAbstractCallbackRunsAndReturnsSyscallRetValue(t, &cb)
}
