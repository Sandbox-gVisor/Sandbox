package kernel

import (
	"fmt"
	"github.com/dop251/goja"
	"gvisor.dev/gvisor/pkg/sentry/arch"
	"gvisor.dev/gvisor/pkg/sentry/kernel/callbacks"
	"testing"
)

func testCreateEmptyTask() Task {
	return Task{}
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

var cbNewSyscallReturnValue = `
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
	task := testCreateEmptyTask()
	args := arch.SyscallArguments{}
	newArgs, rval, err := RunAbstractCallback(
		&task,
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
		t.Fatalf("nil return value")
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
			CallbackSource: cbNewSyscallReturnValue,
			CallbackBody:   cbNewSyscallReturnValue,
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
			CallbackSource: cbNewSyscallReturnValue,
			CallbackBody:   cbNewSyscallReturnValue,
			CallbackArgs:   []string{"count"},
			Type:           JsCallbackTypeAfter,
		},
	}

	testRunAbstractCallbackRunsAndReturnsSyscallRetValue(t, &cb)
}

var cbGetExpectedArgsReturnNewSyscallArgs = `
	function cb(testArg0, testArg1, testArg2, testArg3, testArg4, testArg5) {
		if (args.arg0 != testArg0 || testArg0 != 0) {
			return {
				"ret": -1,
				"errno": 1
			}
		}

		if (args.arg1 != testArg1 || testArg1 != 1) {
			return {
				"ret": -1,
				"errno": 2
			}
		}

		if (args.arg2 != testArg2 || testArg2 != 2) {
			return {
				"ret": -1,
				"errno": 3
			}
		}

		if (args.arg3 != testArg3 || testArg3 != 3) {
			return {
				"ret": -1,
				"errno": 4
			}
		}

		if (args.arg4 != testArg4 || testArg4 != 4) {
			return {
				"ret": -1,
				"errno": 5
			}
		}

		if (args.arg5 != testArg5 || testArg5 != 5) {
			return {
				"ret": -1,
				"errno": 6
			}
		}

		return {
			"0": 20,
			"1": 21,
			"2": 22,
			"3": 23,
			"4": 24,
			"5": 25
		}
	}
`

func testRunAbstractCallbackGetCorrectArguments(t *testing.T, cb JsCallback) {
	task := testCreateEmptyTask()
	args := arch.SyscallArguments{}
	for i := 0; i < len(args); i++ {
		args[i] = arch.SyscallArgument{Value: uintptr(i)}
	}
	newArgs, rval, err := RunAbstractCallback(
		&task,
		jsCallbackInvocationTemplate(cb),
		&args,
		ScriptContextsBuilderOf().Build())
	if err != nil {
		t.Fatalf("failed to execute callback: %s err: %s", cb.callbackInfo().ToString(), err)
	}
	if rval != nil {
		t.Fatalf("unexpected return value. Check argument failed on argument: %v", rval.errno-1)
	}
	if len(newArgs) != len(args) && len(newArgs) == 6 {
		t.Fatalf("argument count doesn't match")
	}
	for i := 0; i < len(newArgs); i++ {
		val := 20 + i
		if newArgs[i].Value != uintptr(val) {
			t.Fatalf("argument doesn't match got %v expected %v", newArgs[i].Value, val)
		}
	}
}

func TestRunAbstractCallback_WithJsCallbackBefore_hasCorrectArgsAndReturnNewArgs(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()

	cb := JsCallbackBefore{
		info: callbacks.JsCallbackInfo{
			Sysno:          1,
			EntryPoint:     "cb",
			CallbackSource: cbGetExpectedArgsReturnNewSyscallArgs,
			CallbackBody:   cbGetExpectedArgsReturnNewSyscallArgs,
			CallbackArgs:   make([]string, 0),
			Type:           JsCallbackTypeBefore,
		},
	}

	for i := 0; i < 6; i++ {
		cb.info.CallbackArgs = append(cb.info.CallbackArgs, fmt.Sprintf("testArg%v", i))
	}

	testRunAbstractCallbackGetCorrectArguments(t, &cb)
}

func TestRunAbstractCallback_WithJsCallbackAfter_hasCorrectArgsAndReturnNewArgs(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()

	cb := JsCallbackAfter{
		info: callbacks.JsCallbackInfo{
			Sysno:          1,
			EntryPoint:     "cb",
			CallbackSource: cbGetExpectedArgsReturnNewSyscallArgs,
			CallbackBody:   cbGetExpectedArgsReturnNewSyscallArgs,
			CallbackArgs:   make([]string, 0),
			Type:           JsCallbackTypeAfter,
		},
	}

	for i := 0; i < 6; i++ {
		cb.info.CallbackArgs = append(cb.info.CallbackArgs, fmt.Sprintf("testArg%v", i))
	}

	testRunAbstractCallbackGetCorrectArguments(t, &cb)
}

type stubDependentGoHook struct {
	createCount int
	callCount   int
}

func (*stubDependentGoHook) description() HookInfoDto {
	return HookInfoDto{}
}

func (*stubDependentGoHook) jsName() string {
	return "stubD"
}

func (h *stubDependentGoHook) createCallback(t *Task) HookCallback {
	h.createCount += 1
	return func(args ...goja.Value) (interface{}, error) {
		h.callCount += 1
		return nil, nil
	}
}

type stubIndependentGoHook struct {
	createCount int
	callCount   int
}

func (*stubIndependentGoHook) description() HookInfoDto {
	return HookInfoDto{}
}

func (*stubIndependentGoHook) jsName() string {
	return "stubI"
}

func (h *stubIndependentGoHook) createCallback() HookCallback {
	h.createCount += 1
	return func(args ...goja.Value) (interface{}, error) {
		h.callCount += 1
		return nil, nil
	}
}

var cbHooksIsCalled = `
	function cb() {
		for (i = 0; i < 10; i++) {
			hooks.stubD()
			hooks.stubI()
		}
	}
`

func testScriptRunsWithHookUsage(t *testing.T, cb JsCallback) {
	task := testCreateEmptyTask()
	args := arch.SyscallArguments{}
	newArgs, rval, err := RunAbstractCallback(
		&task,
		jsCallbackInvocationTemplate(cb),
		&args,
		ScriptContextsBuilderOf().Build())
	if err != nil {
		t.Fatalf("unexpected error while executing callback: %s", err)
	}
	if len(newArgs) == len(args) {
		for i := 0; i < len(args); i++ {
			if newArgs[i].Value != args[i].Value {
				t.Fatalf("unexpected new syscall argument %v", i)
			}
		}
	}
	if rval != nil {
		t.Fatalf("unexpected new return value")
	}
}

func TestRunAbstractCallback_usesInDependentGoHook(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()

	ht := testInitHookTable()
	jsRuntime.hooksTable = &ht
	dHook := stubDependentGoHook{}
	err := jsRuntime.hooksTable.registerDependentHook(&dHook)
	if err != nil {
		t.Fatalf("unexpected error while registering dependent hook to HooksTable: %s", err)
	}
	iHook := stubIndependentGoHook{}
	err = jsRuntime.hooksTable.registerIndependentHook(&iHook)
	if err != nil {
		t.Fatalf("unexpected error while registering independent hook to HooksTable: %s", err)
	}

	cb := JsCallbackBefore{
		info: callbacks.JsCallbackInfo{
			Sysno:          1,
			EntryPoint:     "cb",
			CallbackSource: cbHooksIsCalled,
			CallbackBody:   cbHooksIsCalled,
			CallbackArgs:   make([]string, 0),
			Type:           JsCallbackTypeBefore,
		},
	}

	testScriptRunsWithHookUsage(t, &cb)
	if dHook.createCount != 1 {
		t.Fatalf("dependent hook created: %v times, expected 1", dHook.createCount)
	}
	if iHook.createCount != 1 {
		t.Fatalf("independent hook created: %v times, expected 1", iHook.createCount)
	}
	if dHook.callCount != 10 {
		t.Fatalf("dependent hook called: %v times, expected 10", dHook.callCount)
	}
	if iHook.callCount != 10 {
		t.Fatalf("independent hook called: %v times, expected 10", iHook.callCount)
	}
}

var cbWithNotRegisteredHook = `
	function cb() {
		hooks.abracadabra()
	}
`

func TestRunAbstractCallback_withNotRegisteredHook(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()

	cb := JsCallbackBefore{
		info: callbacks.JsCallbackInfo{
			Sysno:          1,
			EntryPoint:     "cb",
			CallbackSource: cbWithNotRegisteredHook,
			CallbackBody:   cbWithNotRegisteredHook,
			CallbackArgs:   make([]string, 0),
			Type:           JsCallbackTypeBefore,
		},
	}
	task := testCreateEmptyTask()
	args := arch.SyscallArguments{}
	_, _, err := RunAbstractCallback(
		&task,
		jsCallbackInvocationTemplate(&cb),
		&args,
		ScriptContextsBuilderOf().Build())
	if err == nil {
		t.Fatalf("unexpected successful execution of callback with unregistered hook")
	}
}

var cbCheckLocalStorageBefore = `
	function cbBefore() {
		const storage = persistence.local
		storage.savedStr = "hello world"
	}
`

var cbCheckLocalStorageAfter = `
	function cbAfter() {
		const storage = persistence.local
		if (storage.savedStr != undefined && storage.savedStr === "hello world") {
			return {
				"ret": 0,
				"errno": 0
			}
		}
		return {
			"ret": -1,
			"errno": 1
		}
	}
`

func testStorage(t *testing.T, cbBefore JsCallback, cbAfter JsCallback, local bool) {
	task := testCreateEmptyTask()
	args := arch.SyscallArguments{}
	beforeArgs, _, err := RunAbstractCallback(
		&task,
		jsCallbackInvocationTemplate(cbBefore),
		&args,
		ScriptContextsBuilderOf().Build())
	if err != nil {
		t.Fatalf("unexpected error while executing callback before: %s", err)
	}
	if !local {
		task = testCreateEmptyTask()
	}
	_, afterRval, err := RunAbstractCallback(
		&task,
		jsCallbackInvocationTemplate(cbAfter),
		beforeArgs,
		ScriptContextsBuilderOf().Build())
	if err != nil {
		t.Fatalf("unexpected error while executing callback after: %s", err)
	}
	if afterRval == nil {
		t.Fatalf("unexpected nil rval from callback")
	}
	if afterRval.returnValue != 0 || afterRval.errno != 0 {
		t.Fatalf("callback executed with failure")
	}
}

func TestRunAbstractCallback_withCheckingLocalStorage(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()

	cbBefore := JsCallbackBefore{
		info: callbacks.JsCallbackInfo{
			Sysno:          1,
			EntryPoint:     "cbBefore",
			CallbackSource: cbCheckLocalStorageBefore,
			CallbackBody:   cbCheckLocalStorageBefore,
			CallbackArgs:   make([]string, 0),
			Type:           JsCallbackTypeBefore,
		},
	}

	cbAfter := JsCallbackAfter{
		info: callbacks.JsCallbackInfo{
			Sysno:          1,
			EntryPoint:     "cbAfter",
			CallbackSource: cbCheckLocalStorageAfter,
			CallbackBody:   cbCheckLocalStorageAfter,
			CallbackArgs:   make([]string, 0),
			Type:           JsCallbackTypeAfter,
		},
	}

	testStorage(t, &cbBefore, &cbAfter, true)
}

var cbCheckGlobalStorageBefore = `
	function cbBefore() {
		const storage = persistence.glb
		storage.savedStr = "hello world"
	}
`

var cbCheckGlobalStorageAfter = `
	function cbAfter() {
		const storage = persistence.glb
		if (storage.savedStr != undefined && storage.savedStr === "hello world") {
			return {
				"ret": 0,
				"errno": 0
			}
		}
		return {
			"ret": -1,
			"errno": 1
		}
	}
`

func TestRunAbstractCallback_withCheckingGlobalStorage(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()

	cbBefore := JsCallbackBefore{
		info: callbacks.JsCallbackInfo{
			Sysno:          1,
			EntryPoint:     "cbBefore",
			CallbackSource: cbCheckGlobalStorageBefore,
			CallbackBody:   cbCheckGlobalStorageBefore,
			CallbackArgs:   make([]string, 0),
			Type:           JsCallbackTypeBefore,
		},
	}

	cbAfter := JsCallbackAfter{
		info: callbacks.JsCallbackInfo{
			Sysno:          1,
			EntryPoint:     "cbAfter",
			CallbackSource: cbCheckGlobalStorageAfter,
			CallbackBody:   cbCheckGlobalStorageAfter,
			CallbackArgs:   make([]string, 0),
			Type:           JsCallbackTypeAfter,
		},
	}

	testStorage(t, &cbBefore, &cbAfter, false)
}
