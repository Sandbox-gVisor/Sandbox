package kernel

import (
	"encoding/json"
	"testing"
)

func TestGetHooksInfoCommand_execute(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()

	hooks := jsRuntime.hooksTable.getCurrentHooks()
	cmd := GetHooksInfoCommand{}
	res, err := cmd.execute(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error while executing command %s", err)
	}
	rsp, ok := res.(HooksInfoCommandResponse)
	if !ok {
		t.Fatalf("bad return value, got %T expected %T", res, HooksInfoCommandResponse{})
	}
	dtos := rsp.HooksInfo
	if len(dtos) != len(hooks) {
		t.Fatalf("Mismatch of dto count: got %v, expected %v", len(dtos), len(hooks))
	}
	for i := 0; i < len(hooks); i++ {
		info := hooks[i].description()
		if info != dtos[i] {
			t.Fatalf("info dtos do not match for hooks %v (got), %v (expected)", dtos[i].Name, info.Name)
		}
	}
}

func TestGetHooksInfoCommand_name(t *testing.T) {
	cmd := GetHooksInfoCommand{}
	if cmd.name() != "hooks-info" {
		t.Fatalf("wrong cmd name: got '%s', expected 'hooks-info'", cmd.name())
	}
}

var simpleCallbackWithRegistration = `
	function cb() {}

	hooks.AddCbBefore(1, cb)
`

func TestChangeStateCommand_execute(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()

	reqDto := ChangeStateRequestDto{Source: simpleCallbackWithRegistration}
	reqBytes, err := json.Marshal(reqDto)
	if err != nil {
		t.Fatalf("failed to marshal request dto with err: %s", err)
	}
	cmd := ChangeStateCommand{}
	res, err := cmd.execute(nil, reqBytes)
	if err != nil {
		t.Fatalf("unexpected error while executing command %s", err)
	}
	val, ok := res.(string)
	if !ok {
		t.Fatalf("result should have type string, but got %T", res)
	}
	if val != "{}" {
		t.Fatalf("'{}' expected, got %s", val)
	}
	_, ok = jsRuntime.callbackTable.callbackBefore[1]
	if !ok {
		t.Fatalf("script did not register the callback")
	}
}

func TestChangeStateCommand_name(t *testing.T) {
	cmd := ChangeStateCommand{}
	if cmd.name() != "change-state" {
		t.Fatalf("wrong cmd name: got '%s', expected 'change-state'", cmd.name())
	}
}

var simpleCallbackWithEverything = `
	function cb(_, count) {
		hooks.print(count)
	}

	hooks.AddCbBefore(1, cb)
`

func TestCallbackListCommand_execute(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()
	reqDto := ChangeStateRequestDto{Source: simpleCallbackWithEverything}
	reqBytes, err := json.Marshal(reqDto)
	if err != nil {
		t.Fatalf("failed to marshal request dto with err: %s", err)
	}
	var cmd Command = ChangeStateCommand{}
	_, err = cmd.execute(nil, reqBytes)
	if err != nil {
		t.Fatalf("unexpected error while executing change state command %s", err)
	}

	cmd = CallbacksListCommand{}
	res, err := cmd.execute(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error while executing callback list command %s", err)
	}
	val, ok := res.(CallbackListResponse)
	if !ok {
		t.Fatalf("type mismatch: got %T, expected %T", res, CallbackListResponse{})
	}
	infos := val.JsCallbacks
	if len(infos) != 1 {
		t.Fatalf("wrong callback count: got %v, expected 1", len(infos))
	}
	info := infos[0]
	if info.Sysno != 1 {
		t.Log(info.ToString())
		t.Fatalf("wrong callback sysno: got %v, expected 1", info.Sysno)
	}
	if info.EntryPoint != "cb" {
		t.Log(info.ToString())
		t.Fatalf("wrong callback entry point: got \"%v\", expected \"cb\"", info.EntryPoint)
	}
	wantArgs := []string{"_", "count"}
	if len(info.CallbackArgs) != len(wantArgs) {
		t.Log(info.ToString())
		t.Fatalf("wrong count of callback args: got %v, expected %v", len(info.CallbackArgs), len(wantArgs))
	}
	for i := 0; i < len(wantArgs); i++ {
		if info.CallbackArgs[i] != wantArgs[i] {
			t.Log(info.ToString())
			t.Fatalf("wrong [%v] arg: got \"%s\", expected \"%s\"", i, info.CallbackArgs[i], wantArgs[i])
		}
	}
	wantSrc := `function cb(_, count) {
		hooks.print(count)
	}`
	if info.CallbackSource != wantSrc {
		t.Log(info.ToString())
		t.Fatalf("wrong callback source: \n--- got:\n%s\n\n--- expected:\n%s\n", info.CallbackSource, wantSrc)
	}
	if info.CallbackBody != wantSrc {
		t.Log(info.ToString())
		t.Fatalf("wrong callback body: \n--- got:\n%s\n\n--- expected:\n%s\n", info.CallbackBody, wantSrc)
	}
}

func TestCallbacksListCommand_name(t *testing.T) {
	cmd := CallbacksListCommand{}
	wantName := "current-callbacks"

	if cmd.name() != wantName {
		t.Fatalf("wrong command name: got '%s', expected '%s'", cmd.name(), wantName)
	}
}
