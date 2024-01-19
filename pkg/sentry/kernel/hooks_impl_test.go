package kernel

import (
	"github.com/dop251/goja"
	"testing"
)

var simpleAddCbBefore = `
	function cb() {}

	hooks.AddCbBefore(1, cb)
 `

func TestAddCbBeforeHook(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()
	builder := ScriptContextsBuilderOf()
	builder = builder.AddContext3(HooksJsName, &IndependentHookAddableAdapter{ht: jsRuntime.hooksTable})
	builder = builder.AddContext3(JsPersistenceContextName,
		&ObjectAddableAdapter{name: JsGlobalPersistenceObject, object: jsRuntime.Global})
	contexts := builder.Build()

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
}
