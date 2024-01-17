package kernel

import (
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
		t.Fatalf("unexpected error while registering dependent hook: %s", err)
	}

	_, ok := ht.independentHooks[h.jsName()]
	if !ok {
		t.Fatalf("dependent hook was't registered")
	}
}
