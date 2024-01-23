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
		t.Fatalf("unexpected error while registering new cmd")
	}

	_, ok := ct.commands[cmd.name()]
	if !ok {
		t.Fatalf("command was not registered")
	}
}
