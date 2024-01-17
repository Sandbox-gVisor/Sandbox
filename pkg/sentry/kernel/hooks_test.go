package kernel

import (
	"sync"
	"testing"
)

func testInitHookTable() HooksTable {
	return HooksTable{
		DependentHooks:   make(map[string]TaskDependentGoHook),
		IndependentHooks: make(map[string]TaskIndependentGoHook),
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

	_, ok := ht.DependentHooks[h.jsName()]
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

	_, ok := ht.IndependentHooks[h.jsName()]
	if !ok {
		t.Fatalf("dependent hook was't registered")
	}
}
