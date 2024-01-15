package kernel

import "testing"

func testCreateMockTask() *Task {
	return &Task{}
}

func testInitJsRuntime() {
	jsRuntime = initJsRuntime()
}

func testDestroyJsRuntime() {
	jsRuntime = nil
}

func TestRunJsScript_RunsAndReturnsValue(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()

	src := `
	function testF() {
		a = 0
		for (i = 0; i < 10; i++) {
			a = a + 1
		}
		return a
	}

	testF()
	`
	jsVal, err := RunJsScript(jsRuntime.JsVM, src, []ScriptContext{})
	if err != nil {
		t.Fatalf("failed to execute script with err %s", err)
	}
	var val int64
	err = jsRuntime.JsVM.ExportTo(jsVal, &val)
	if err != nil {
		t.Fatalf("failed to convert return value %s", err)
	}
	if val != 10 {
		t.Fatalf("Wrong value. Got %v, expected 10", val)
	}
}
