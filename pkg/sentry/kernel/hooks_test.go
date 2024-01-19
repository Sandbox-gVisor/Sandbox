package kernel

import (
	"github.com/dop251/goja"
	"sync"
	"testing"
)

func testInitHookTable() HooksTable {
	return HooksTable{
		dependentHooks:   make(map[string]TaskDependentGoHook),
		independentHooks: make(map[string]TaskIndependentGoHook),
		mutex:            sync.Mutex{},
	}
}

func TestHooksTable_registerDependentHook(t *testing.T) {
	ht := testInitHookTable()

	h := stubDependentGoHook{}
	err := ht.registerDependentHook(&h)
	if err != nil {
		t.Fatalf("unexpected error while registering dependent hook: %s", err)
	}

	_, ok := ht.dependentHooks[h.jsName()]
	if !ok {
		t.Fatalf("dependent hook was't registered")
	}
}

func TestHooksTable_registerIndependentHook(t *testing.T) {
	ht := testInitHookTable()

	h := stubIndependentGoHook{}
	err := ht.registerIndependentHook(&h)
	if err != nil {
		t.Fatalf("unexpected error while registering independent hook: %s", err)
	}

	_, ok := ht.independentHooks[h.jsName()]
	if !ok {
		t.Fatalf("independent hook was't registered")
	}
}

func TestHooksTable_addIndependentHooksToContextObject(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()
	ht := testInitHookTable()
	h := stubIndependentGoHook{}
	obj := jsRuntime.JsVM.NewObject()

	_ = ht.registerIndependentHook(&h)
	err := ht.addIndependentHooksToContextObject(obj)
	if err != nil {
		t.Fatalf("failed to add independent hooks to context object")
	}
	val := obj.Get(h.jsName())
	if goja.IsNull(val) || goja.IsUndefined(val) {
		t.Fatalf("no value added to object")
	}
}

func TestHooksTable_addDependentHooksToContextObject(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()
	ht := testInitHookTable()
	task := testCreateEmptyTask()
	h := stubDependentGoHook{}
	obj := jsRuntime.JsVM.NewObject()

	_ = ht.registerDependentHook(&h)
	err := ht.addDependentHooksToContextObject(obj, &task)
	if err != nil {
		t.Fatalf("failed to add dependent hooks to context object")
	}
	val := obj.Get(h.jsName())
	if goja.IsNull(val) || goja.IsUndefined(val) {
		t.Fatalf("no value added to object")
	}
}
