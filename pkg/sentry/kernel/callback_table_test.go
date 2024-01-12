package kernel

import (
	"gvisor.dev/gvisor/pkg/sentry/arch"
	"gvisor.dev/gvisor/pkg/sentry/kernel/callbacks"
	"testing"
)

type testCbBefore struct{}

func (testCbBefore) CallbackBeforeFunc(
	t *Task,
	sysno uintptr,
	args *arch.SyscallArguments,
) (*arch.SyscallArguments, *SyscallReturnValue, error) {
	return args, nil, nil
}

func (testCbBefore) Info() callbacks.JsCallbackInfo {
	return callbacks.JsCallbackInfo{}
}

func TestCallbackTable_registerCallbackBefore(t *testing.T) {
	cbt := initCallbackTable()

	f := testCbBefore{}

	err := cbt.registerCallbackBefore(1, f)
	if err != nil {
		t.Fatalf("unexpected error while adding non nil callbackBefore")
	}

	val, ok := cbt.callbackBefore[1]
	if !ok {
		t.Fatalf("callbackBefore not registored")
	}

	if val != f {
		t.Fatalf("not same callbackBefore registered")
	}

	err = cbt.registerCallbackBefore(2, nil)
	if err == nil {
		t.Fatalf("registering nil callbackBefore")
	}
}

type testCbAfter struct{}

func (testCbAfter) CallbackAfterFunc(
	t *Task,
	sysno uintptr,
	args *arch.SyscallArguments,
	ret uintptr,
	err error,
) (*arch.SyscallArguments, *SyscallReturnValue, error) {
	return args, nil, nil
}

func (testCbAfter) Info() callbacks.JsCallbackInfo {
	return callbacks.JsCallbackInfo{}
}

func initCallbackTable() CallbackTable {
	return CallbackTable{
		callbackBefore: make(map[uintptr]CallbackBefore),
		callbackAfter:  make(map[uintptr]CallbackAfter),
	}
}

func TestCallbackTable_registerCallbackAfter(t *testing.T) {
	cbt := initCallbackTable()

	f := testCbAfter{}

	err := cbt.registerCallbackAfter(1, f)
	if err != nil {
		t.Fatalf("unexpected error while adding non nil callbackAfter")
	}

	val, ok := cbt.callbackAfter[1]
	if !ok {
		t.Fatalf("callbackAfter not registored")
	}

	if val != f {
		t.Fatalf("not same callbackAfter registered")
	}

	err = cbt.registerCallbackAfter(2, nil)
	if err == nil {
		t.Fatalf("registering nil callbackAfter")
	}
}

func TestCallbackTable_UnregisterAll(t *testing.T) {
	cbt := initCallbackTable()

	for i := 0; i < 100; i++ {
		err := cbt.registerCallbackBefore(uintptr(i), testCbBefore{})
		if err != nil {
			t.Fatalf("unexpected error while adding non nil callbackBefore")
		}
		err = cbt.registerCallbackAfter(uintptr(i), testCbAfter{})
		if err != nil {
			t.Fatalf("unexpected error while adding non nil callbackAfter")
		}
	}

	cbCountTotal := len(cbt.callbackAfter) + len(cbt.callbackBefore)
	if cbCountTotal != 200 {
		t.Fatalf("Callback amout mismatch has %v expected 200", cbCountTotal)
	}

	cbt.UnregisterAll()
	cbCountTotal = len(cbt.callbackAfter) + len(cbt.callbackBefore)
	if cbCountTotal != 0 {
		t.Fatalf("Not all callbacks were unregister. Callbacks left: %d", cbCountTotal)
	}
}

func TestCallbackTable_unregisterCallbackBefore(t *testing.T) {
	cbt := initCallbackTable()

	err := cbt.unregisterCallbackBefore(1)
	if err == nil {
		t.Fatalf("unregistered not existed callbackBefore")
	}

	_ = cbt.registerCallbackBefore(1, testCbBefore{})

	err = cbt.unregisterCallbackBefore(1)
	if err != nil {
		t.Fatalf("unexpected failure of unregistering callbackBefore")
	}

	cbCount := len(cbt.callbackBefore)
	if cbCount != 0 {
		t.Fatalf("Bad unregistering. CallbackBefore is still there.")
	}
}

func TestCallbackTable_unregisterCallbackAfter(t *testing.T) {
	cbt := initCallbackTable()

	err := cbt.unregisterCallbackAfter(1)
	if err == nil {
		t.Fatalf("unregistered not existed callbackAfter")
	}

	_ = cbt.registerCallbackAfter(1, testCbAfter{})

	err = cbt.unregisterCallbackAfter(1)
	if err != nil {
		t.Fatalf("unexpected failure of unregistering callbackAfter")
	}

	cbCount := len(cbt.callbackAfter)
	if cbCount != 0 {
		t.Fatalf("Bad unregistering. CallbackAfter is still there.")
	}
}

func TestCallbackTable_getCallbackBefore(t *testing.T) {
	cbt := initCallbackTable()

	cb := cbt.getCallbackBefore(1)
	if cb != nil {
		t.Fatalf("get callback from empty table")
	}

	f := testCbBefore{}
	_ = cbt.registerCallbackBefore(1, f)

	cb = cbt.getCallbackBefore(1)
	if cb != f {
		t.Fatalf("registered and got callbacks differs")
	}
}

func TestCallbackTable_getCallbackAfter(t *testing.T) {
	cbt := initCallbackTable()

	cb := cbt.getCallbackAfter(1)
	if cb != nil {
		t.Fatalf("get callback from empty table")
	}

	f := testCbAfter{}
	_ = cbt.registerCallbackAfter(1, f)

	cb = cbt.getCallbackAfter(1)
	if cb != f {
		t.Fatalf("registered and got callbacks differs")
	}
}
