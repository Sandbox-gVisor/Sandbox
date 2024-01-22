package kernel

import (
	"fmt"
	"github.com/dop251/goja"
	"testing"
)

var simpleAddCbAfter = `
	function cb(  
	) {}

	hooks.AddCbAfter(1, cb)
`

func TestAddCbAfterHook_registersCallback(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()
	contexts := testBuildContexts()

	val, err := RunJsScript(jsRuntime.JsVM, simpleAddCbAfter, contexts)
	if err != nil {
		t.Fatalf("unexpected error while registering callback")
	}
	if !goja.IsNull(val) {
		t.Fatalf("unexpected return value")
	}
	cb := jsRuntime.callbackTable.getCallbackAfter(1)
	if cb == nil {
		t.Fatalf("callback wasn't registered")
	}
	info := cb.Info()
	if info.Sysno != 1 {
		fmt.Println(info.ToString())
		t.Fatalf("bad sysno in info: got %v expected 1", info.Sysno)
	}
	if info.EntryPoint != "cb" {
		fmt.Println(info.ToString())
		t.Fatalf("bad entry point: got '%v', expected 'cb'")
	}
	if len(info.CallbackArgs) != 0 {
		fmt.Println(info.ToString())
		t.Fatalf("Callback args should be empty for this callback: got %v", info.CallbackArgs)
	}
	if info.Type != JsCallbackTypeAfter {
		fmt.Println(info.ToString())
		t.Fatalf("wrong cb type: got %s, expected %s", info.Type, JsCallbackTypeAfter)
	}
	if info.CallbackBody != simpleCb {
		fmt.Println(info.ToString())
		t.Fatalf("wrong cb body: got\n%s, expected\n%s\n", info.CallbackBody, simpleCb)
	}
	if info.CallbackSource != simpleCb {
		fmt.Println(info.ToString())
		t.Fatalf("wrong cb source: got\n%s, expected\n%s\n", info.CallbackSource, simpleCb)
	}
}

var addCbAfterWith1arg = `
	hooks.AddCbAfter(1)
`

func TestAddCbAfterHook_fails_with1arg(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()
	contexts := testBuildContexts()

	_, err := RunJsScript(jsRuntime.JsVM, addCbAfterWith1arg, contexts)
	if err == nil {
		t.Fatalf("no error in callback which 1 argument was given instead of 2")
	}
}

var addCbAfterWith3arg = `
	function cb() {}

	hooks.AddCbAfter(1, cb, 3)
`

func TestAddCbAfterHook_fails_with3arg(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()
	contexts := testBuildContexts()

	_, err := RunJsScript(jsRuntime.JsVM, addCbAfterWith3arg, contexts)
	if err == nil {
		t.Fatalf("no error in callback which 3 argument was given instead of 2")
	}
}

var addCbAfterWithNullSysno = `
	function cb() {}

	hooks.AddCbAfter(null, cb)
`

func TestAddCbAfterHook_fails_withNullSysno(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()
	contexts := testBuildContexts()

	_, err := RunJsScript(jsRuntime.JsVM, addCbAfterWithNullSysno, contexts)
	if err == nil {
		t.Fatalf("no error in callback which 1 argument is null")
	}
}

var addCbAfterWithUndefinedSysno = `
	function cb() {}

	hooks.AddCbAfter(undefined, cb)
`

func TestAddCbAfterHook_fails_withUndefinedSysno(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()
	contexts := testBuildContexts()

	_, err := RunJsScript(jsRuntime.JsVM, addCbAfterWithUndefinedSysno, contexts)
	if err == nil {
		t.Fatalf("no error in callback which 1 argument is undefined")
	}
}

var addCbAfterWithNullCb = `
	hooks.AddCbAfter(1, null)
`

func TestAddCbAfterHook_fails_withNullCb(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()
	contexts := testBuildContexts()

	_, err := RunJsScript(jsRuntime.JsVM, addCbAfterWithNullCb, contexts)
	if err == nil {
		t.Fatalf("no error in callback which 2 argument is null")
	}
}

var addCbAfterWithUndefinedCb = `
	hooks.AddCbAfter(1, undefined)
`

func TestAddCbAfterHook_fails_withUndefinedCb(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()
	contexts := testBuildContexts()

	_, err := RunJsScript(jsRuntime.JsVM, addCbAfterWithUndefinedCb, contexts)
	if err == nil {
		t.Fatalf("no error in callback which 2 argument is undefined")
	}
}

var addCbAfterWithNoExistedCb = `
	hooks.AddCbAfter(1, cb)
`

func TestAddCbAfterHook_fails_withNotExistedCb(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()
	contexts := testBuildContexts()

	_, err := RunJsScript(jsRuntime.JsVM, addCbAfterWithNoExistedCb, contexts)
	if err == nil {
		t.Fatalf("no error in callback which 2 argument is not existing object")
	}
}
