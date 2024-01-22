package kernel

import (
	"gvisor.dev/gvisor/pkg/sentry/arch"
	util "gvisor.dev/gvisor/pkg/sentry/kernel/callbacks"
	"testing"
)

func testBuildContexts() ScriptContexts {
	builder := ScriptContextsBuilderOf()
	builder = builder.AddContext3(HooksJsName, &IndependentHookAddableAdapter{ht: jsRuntime.hooksTable})
	builder = builder.AddContext3(JsPersistenceContextName,
		&ObjectAddableAdapter{name: JsGlobalPersistenceObject, object: jsRuntime.Global})
	return builder.Build()
}

// testThatCbFailsWithErr does the following
//   - init js runtime
//   - create empty task
//   - create js callback before, note that entry point (function name) should be "cb"
//   - call RunAbstractCallback
//   - fails if err == nil with given message
func testThatCbFailsWithErr(t *testing.T, cbSource string, failMessage string) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()
	task := testCreateEmptyTask()
	args := arch.SyscallArguments{}
	cb := JsCallbackBefore{info: util.JsCallbackInfo{
		Sysno:          1,
		CallbackSource: cbSource,
		CallbackBody:   cbSource,
		CallbackArgs:   []string{},
		Type:           JsCallbackTypeBefore,
		EntryPoint:     "cb",
	}}

	_, _, err := RunAbstractCallback(&task, jsCallbackInvocationTemplate(&cb), &args, ScriptContextsBuilderOf().Build())
	if err == nil {
		t.Fatalf(failMessage)
	}
}
