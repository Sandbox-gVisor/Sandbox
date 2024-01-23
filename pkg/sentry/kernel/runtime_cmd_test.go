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
}

func TestChangeStateCommand_name(t *testing.T) {
	cmd := ChangeStateCommand{}
	if cmd.name() != "change-state" {
		t.Fatalf("wrong cmd name: got '%s', expected 'change-state'", cmd.name())
	}
}
