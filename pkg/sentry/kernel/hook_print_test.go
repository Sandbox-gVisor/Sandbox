package kernel

import (
	"gvisor.dev/gvisor/pkg/sentry/arch"
	util "gvisor.dev/gvisor/pkg/sentry/kernel/callbacks"
	"os"
	"testing"
)

var printHookWorks = `
	function cb() {
		hooks.print("hello world", "from test")
	}
`

func TestPrintHook_Works(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()
	task := testCreateEmptyTask()
	args := arch.SyscallArguments{}
	cb := JsCallbackBefore{info: util.JsCallbackInfo{
		Sysno:          1,
		CallbackSource: printHookWorks,
		CallbackBody:   printHookWorks,
		CallbackArgs:   []string{},
		Type:           JsCallbackTypeBefore,
		EntryPoint:     "cb",
	}}
	f, err := os.CreateTemp("", "testPrintHook*")
	if err != nil {
		t.Fatalf("failed to create tmp file")
	}
	prevStdout := os.Stdout
	os.Stdout = f

	defer func() {
		os.Stdout = prevStdout
		fileName := f.Name()
		_ = f.Close()
		_ = os.Remove(fileName)
	}()

	_, _, err = RunAbstractCallback(&task, jsCallbackInvocationTemplate(&cb), &args, ScriptContextsBuilderOf().Build())
	if err != nil {
		t.Fatalf("unexpected error while executing callback")
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		t.Fatalf("failed to use seek")
	}
	expectedStr := "hello world from test"
	buf := make([]byte, len(expectedStr))
	wasRead, err := f.Read(buf)
	if wasRead != len(expectedStr) {
		t.Fatalf("amount of bytes written and read do not match: got %v, expected %v", wasRead, len(expectedStr))
	}
	if string(buf) != expectedStr {
		t.Fatalf("the contents of the files do not match: got '%s', expected '%s'", string(buf), expectedStr)
	}
}
