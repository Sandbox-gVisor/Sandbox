package kernel

import (
	"cmp"
	"github.com/dop251/goja"
	"slices"
	"strings"
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
	if h.createCount != 1 {
		t.Fatalf("hook was not created")
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
	if h.createCount != 1 {
		t.Fatalf("hook was not created")
	}
}

// prepareSetOfDependentHooks should register same dependent hooks as RegisterHooks
func prepareSetOfDependentHooks() map[string]struct{} {
	set := make(map[string]struct{})

	set[(&ReadBytesHook{}).jsName()] = struct{}{}
	set[(&WriteBytesHook{}).jsName()] = struct{}{}
	set[(&ReadStringHook{}).jsName()] = struct{}{}
	set[(&WriteStringHook{}).jsName()] = struct{}{}
	set[(&EnvvGetterHook{}).jsName()] = struct{}{}
	set[(&MmapGetterHook{}).jsName()] = struct{}{}
	set[(&ArgvHook{}).jsName()] = struct{}{}
	set[(&SignalInfoHook{}).jsName()] = struct{}{}
	set[(&PidInfoHook{}).jsName()] = struct{}{}
	set[(&FDHook{}).jsName()] = struct{}{}
	set[(&FDsHook{}).jsName()] = struct{}{}
	set[(&UserJSONLogHook{}).jsName()] = struct{}{}
	set[(&AnonMmapHook{}).jsName()] = struct{}{}
	set[(&MunmapHook{}).jsName()] = struct{}{}
	set[(&SignalSendingHook{}).jsName()] = struct{}{}
	set[(&ThreadsStoppingHook{}).jsName()] = struct{}{}
	set[(&ThreadsResumingHook{}).jsName()] = struct{}{}
	set[(&ThreadInfoHook{}).jsName()] = struct{}{}

	return set
}

// prepareSetOfIndependentHooks should register same independent hooks as RegisterHooks
func prepareSetOfIndependentHooks() map[string]struct{} {
	set := make(map[string]struct{})

	set[(&PrintHook{}).jsName()] = struct{}{}
	set[(&AddCbBeforeHook{}).jsName()] = struct{}{}
	set[(&AddCbAfterHook{}).jsName()] = struct{}{}
	set[(&SignalByNameHook{}).jsName()] = struct{}{}
	set[(&SignalMaskToSignalNamesHook{}).jsName()] = struct{}{}

	return set
}

func findUnregistered[H GoHook](prepared map[string]struct{}, registered map[string]H) []string {
	notRegistered := make([]string, 0)
	for hookName := range registered {
		_, ok := prepared[hookName]
		if !ok {
			notRegistered = append(notRegistered, hookName)
		}
	}
	for hookName := range prepared {
		_, ok := registered[hookName]
		if !ok {
			notRegistered = append(notRegistered, hookName)
		}
	}
	slices.Sort(notRegistered)
	return notRegistered
}

func TestRegisterHooks(t *testing.T) {
	ht := testInitHookTable()
	setOfDependentHooks := prepareSetOfDependentHooks()
	setOfIndependentHooks := prepareSetOfIndependentHooks()

	err := RegisterHooks(&ht)
	if err != nil {
		t.Fatalf("unexpected error while registering hooks %s", err)
	}

	// testing dependent hooks
	if len(setOfDependentHooks) != len(ht.dependentHooks) {
		t.Fatalf("mismatch of dependent hooks count in tests (%v) and registered count (%v).\nThese hooks are not rergistered:\n%s",
			len(setOfDependentHooks),
			len(ht.dependentHooks),
			strings.Join(findUnregistered(setOfDependentHooks, ht.dependentHooks), "\n"))
	}

	for hookName := range setOfDependentHooks {
		_, ok := ht.dependentHooks[hookName]
		if !ok {
			t.Fatalf("dependent hook %v is missed.\nAlso check registered hooks and hooks in tests because amount is the same.", hookName)
		}
	}

	// testing independent hooks
	if len(setOfIndependentHooks) != len(ht.independentHooks) {
		t.Fatalf("mismatch of independent hooks count in tests (%v) and registered count (%v).\nThese hooks are not rergistered:\n%s",
			len(setOfIndependentHooks),
			len(ht.independentHooks),
			strings.Join(findUnregistered(setOfIndependentHooks, ht.independentHooks), "\n"))
	}

	for hookName := range setOfIndependentHooks {
		_, ok := ht.independentHooks[hookName]
		if !ok {
			t.Fatalf("independent hook %v is missed.\nAlso check registered hooks and hooks in tests because amount is the same.", hookName)
		}
	}
}

func TestHooksTable_getCurrentHooks(t *testing.T) {
	ht := testInitHookTable()
	setOfDependentHooks := prepareSetOfDependentHooks()
	setOfIndependentHooks := prepareSetOfIndependentHooks()

	_ = RegisterHooks(&ht)
	allHooks := ht.getCurrentHooks()

	if testCount := len(setOfIndependentHooks) + len(setOfDependentHooks); len(allHooks) != testCount {
		t.Fatalf("mismatch for getting current hooks. Test have %v, method returns %v", testCount, len(allHooks))
	}
	for _, h := range allHooks {
		_, okD := setOfDependentHooks[h.jsName()]
		_, okI := setOfIndependentHooks[h.jsName()]
		if okD && okI {
			t.Fatalf("hook %s registered as dependent and independent at the same time", h.jsName())
		}
		if !okD && !okI {
			t.Fatalf("hook %s not in set of existing hooks", h.jsName())
		}
	}

	if !slices.IsSortedFunc[[]GoHook, GoHook](allHooks,
		func(a GoHook, b GoHook) int {
			return cmp.Compare[string](
				strings.ToLower(a.jsName()),
				strings.ToLower(b.jsName()))
		}) {
		t.Fatalf("hooks are not sorted")
	}
}

func TestRegisterHooks_checkIfThereIsNoSuchObjectsInJS(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()

	vm := jsRuntime.JsVM
	obj := vm.NewObject()
	err := vm.Set(HooksJsName, obj)
	if err != nil {
		t.Fatalf("failed to set hooks object to vm")
	}

	for k := range jsRuntime.hooksTable.dependentHooks {
		val := vm.Get(HooksJsName).ToObject(vm).Get(k)
		if val != nil && (!goja.IsUndefined(val) || !goja.IsNull(val)) {
			t.Fatalf("object with name %s is already defined", k)
		}
	}
	for k := range jsRuntime.hooksTable.independentHooks {
		val := vm.Get(HooksJsName).ToObject(vm).Get(k)
		if val != nil && (!goja.IsUndefined(val) || !goja.IsNull(val)) {
			t.Fatalf("object with name %s is already defined", k)
		}
	}
}
