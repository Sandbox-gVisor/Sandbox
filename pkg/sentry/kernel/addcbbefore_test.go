package kernel

import (
	"github.com/dop251/goja"
	"testing"
)

var simpleAddCbBefore = `
	function cb() {}

	hooks.AddCbBefore(1, cb)
 `

func TestAddCbBeforeHook_registersCallback(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()
	contexts := testBuildContexts()

	val, err := RunJsScript(jsRuntime.JsVM, simpleAddCbBefore, contexts)
	if err != nil {
		t.Fatalf("unexpected error while registering callback")
	}
	if !goja.IsNull(val) {
		t.Fatalf("unexpected return value")
	}
	cb := jsRuntime.callbackTable.getCallbackBefore(1)
	if cb == nil {
		t.Fatalf("callback wasn't registered")
	}
	info := cb.Info()
	if info.Sysno != 1 {
		t.Fatalf("bad sysno in info got %v expected 1", info.Sysno)
	}
	if len(info.CallbackArgs) != 0 {
		t.Fatalf("Callback args should be empty for this callback")
	}
}

var addCbBeforeWith1arg = `
	hooks.AddCbBefore(1)
`

func TestAddCbBeforeHook_fails_with1arg(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()
	contexts := testBuildContexts()

	_, err := RunJsScript(jsRuntime.JsVM, addCbBeforeWith1arg, contexts)
	if err == nil {
		t.Fatalf("no error in callback which 1 argument was given instead of 2")
	}
}

var addCbBeforeWith3arg = `
	function cb() {}

	hooks.AddCbBefore(1, cb, 3)
`

func TestAddCbBeforeHook_fails_with3arg(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()
	contexts := testBuildContexts()

	_, err := RunJsScript(jsRuntime.JsVM, addCbBeforeWith3arg, contexts)
	if err == nil {
		t.Fatalf("no error in callback which 3 argument was given instead of 2")
	}
}

var addCbBeforeWithNullSysno = `
	function cb() {}

	hooks.AddCbBefore(null, cb)
`

func TestAddCbBeforeHook_fails_withNullSysno(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()
	contexts := testBuildContexts()

	_, err := RunJsScript(jsRuntime.JsVM, addCbBeforeWithNullSysno, contexts)
	if err == nil {
		t.Fatalf("no error in callback which 1 argument is null")
	}
}

var addCbBeforeWithUndefinedSysno = `
	function cb() {}

	hooks.AddCbBefore(undefined, cb)
`

func TestAddCbBeforeHook_fails_withUndefinedSysno(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()
	contexts := testBuildContexts()

	_, err := RunJsScript(jsRuntime.JsVM, addCbBeforeWithUndefinedSysno, contexts)
	if err == nil {
		t.Fatalf("no error in callback which 1 argument is undefined")
	}
}

var addCbBeforeWithNullCb = `
	hooks.AddCbBefore(1, null)
`

func TestAddCbBeforeHook_fails_withNullCb(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()
	contexts := testBuildContexts()

	_, err := RunJsScript(jsRuntime.JsVM, addCbBeforeWithNullCb, contexts)
	if err == nil {
		t.Fatalf("no error in callback which 2 argument is null")
	}
}

var addCbBeforeWithUndefinedCb = `
	hooks.AddCbBefore(1, undefined)
`

func TestAddCbBeforeHook_fails_withUndefinedCb(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()
	contexts := testBuildContexts()

	_, err := RunJsScript(jsRuntime.JsVM, addCbBeforeWithUndefinedCb, contexts)
	if err == nil {
		t.Fatalf("no error in callback which 2 argument is undefined")
	}
}

var addCbBeforeWithNoExistedCb = `
	hooks.AddCbBefore(1, cb)
`

func TestAddCbBeforeHook_fails_withNotExistedCb(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()
	contexts := testBuildContexts()

	_, err := RunJsScript(jsRuntime.JsVM, addCbBeforeWithNoExistedCb, contexts)
	if err == nil {
		t.Fatalf("no error in callback which 2 argument is not existing object")
	}
}
