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

var testCbList = []string{
	`
	function cb() {
		a = 0
		a += 1
	}

	hooks.AddCbBefore(1, cb)
	`,
	`
	function cb() {
		hooks.print("hello world")
	}

	hooks.AddCbBefore(2, cb)
	`,
	`
	function after_22(_, arg_2) {
		hooks.print(arg_2)
	}
	
	function cb() {
		a = 0
		a += 1
	}

	hooks.AddCbAfter(22, after_22)
	hooks.AddCbAfter(1, cb)
	`,
}

func fillCmds(t *testing.T) {
	for i, cbSrc := range testCbList {
		reqDto := ChangeStateRequestDto{Source: cbSrc}
		reqBytes, err := json.Marshal(reqDto)
		if err != nil {
			t.Fatalf("failed to marshal request dto for callback[%v] with err: %s", i, err)
		}
		var cmd Command = ChangeStateCommand{}
		_, err = cmd.execute(nil, reqBytes)
		if err != nil {
			t.Fatalf("unexpected error while executing change state command for callback[%v]: %s", i, err)
		}
	}
}

func TestUnregisterCallbacksCommand_execute_withAllOption(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()
	fillCmds(t)

	cmd := UnregisterCallbacksCommand{}
	req := UnregisterCallbacksRequest{
		Options: UnregisterAllOption,
		List:    nil,
	}
	reqDtoBytes, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal unregister callback request dto with err: %s", err)
	}
	res, err := cmd.execute(nil, reqDtoBytes)
	if err != nil {
		t.Fatalf("unexpected error while executing unregister command: %s", err)
	}
	if res != nil {
		t.Fatalf("expected nil result")
	}
	if len(jsRuntime.callbackTable.callbackBefore) != 0 {
		t.Fatalf("not all callbacks before were unregistered")
	}
	if len(jsRuntime.callbackTable.callbackAfter) != 0 {
		t.Fatalf("not all callbacks after were unregistered")
	}
}

func TestUnregisterCallbacksCommand_execute_withListOption(t *testing.T) {
	testInitJsRuntime()
	defer testDestroyJsRuntime()
	fillCmds(t)

	cmd := UnregisterCallbacksCommand{}
	req := UnregisterCallbacksRequest{
		Options: UnregisterListOption,
		List: []UnregisterCallbackDto{
			{Sysno: 22, Type: JsCallbackTypeBefore},
			{Sysno: 1, Type: JsCallbackTypeBefore},
		},
	}
	reqDtoBytes, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal unregister callback request dto with err: %s", err)
	}
	res, err := cmd.execute(nil, reqDtoBytes)
	if err == nil {
		t.Fatalf("expected error for tring to delete not existing callback")
	}

	req.List[0].Type = JsCallbackTypeAfter
	reqDtoBytes, err = json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal unregister callback request dto with err: %s", err)
	}
	res, err = cmd.execute(nil, reqDtoBytes)
	if err != nil {
		t.Fatalf("unexpected error while executing unregister callbacks command: %s", err)
	}
	if res != nil {
		t.Fatalf("expected nil result")
	}
	cbBeforeCount := len(jsRuntime.callbackTable.callbackBefore)
	if cbBeforeCount != 1 {
		t.Fatalf("wrong number of callbacks before: got %v, expected 1", cbBeforeCount)
	}
	cbAfterCount := len(jsRuntime.callbackTable.callbackAfter)
	if cbAfterCount != 1 {
		t.Fatalf("wrong number of callbacks after: got %v, expected 1", cbAfterCount)
	}
}

func TestUnregisterCallbacksCommand_name(t *testing.T) {
	cmd := UnregisterCallbacksCommand{}
	wantName := "unregister-callbacks"

	if cmd.name() != wantName {
		t.Fatalf("wrong command name: got '%s', expected '%s'", cmd.name(), wantName)
	}
}
