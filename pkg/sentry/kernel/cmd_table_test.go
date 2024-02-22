package kernel

import (
	"sync"
	"testing"
)

type stubCommand struct {
	nameCalled    int
	executeCalled int
}

func (c *stubCommand) name() string {
	c.nameCalled += 1
	return "stub"
}

func (c *stubCommand) execute(k *Kernel, raw []byte) (any, error) {
	c.executeCalled += 1
	return nil, nil
}

func TestCommandTable_Register(t *testing.T) {
	var ct *CommandTable
	err := ct.Register(nil)
	if err == nil {
		t.Fatalf("no error then calling method on nil table")
	}
	ct = &CommandTable{}
	err = ct.Register(nil)
	if err == nil {
		t.Fatalf("no error then calling method on table with nil map inside")
	}

	ct.commands = make(map[string]Command)
	ct.mutex = sync.Mutex{}

	cmd := stubCommand{}
	err = ct.Register(&cmd)
	if err != nil {
		t.Fatalf("unexpected error while registering new cmd %s", err)
	}
	if cmd.nameCalled != 1 {
		t.Fatalf("expected count of calls to name() is 1, got %v", cmd.nameCalled)
	}

	_, ok := ct.commands[cmd.name()]
	if !ok {
		t.Fatalf("command was not registered")
	}
}

func TestCommandTable_GetCommand(t *testing.T) {
	var ct *CommandTable
	_, err := ct.GetCommand("stub")
	if err == nil {
		t.Fatalf("no error then calling method on nil table")
	}

	ct = &CommandTable{
		commands: make(map[string]Command),
		mutex:    sync.Mutex{},
	}
	_, err = ct.GetCommand("stub")
	if err == nil {
		t.Fatalf("error should be returned if command does not exist")
	}

	stored := stubCommand{}
	ct.commands[stored.name()] = &stored

	cmd, err := ct.GetCommand(stored.name())
	if err != nil {
		t.Fatalf("error while getting command %s", err)
	}
	stub, ok := cmd.(*stubCommand)
	if !ok {
		t.Fatalf("got cmd and expected cmd has different types: got %T, expected %T", cmd, stored)
	}
	if stub.executeCalled != stored.executeCalled || stub.nameCalled != stored.nameCalled {
		t.Fatalf("objects does not match")
	}
}
