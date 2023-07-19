package kernel

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
)

type Command interface {
	name() string

	execute(kernel *Kernel, raw []byte) ([]byte, error)
}

type NoSuchCommand struct{}

func (n NoSuchCommand) name() string {
	return ""
}

func (n NoSuchCommand) execute(_ *Kernel, _ []byte) ([]byte, error) {
	ret := "{\"type\": \"not_ok\", \"cause\": \"no such command\"}"
	return []byte(ret), nil
}

type TypeDto struct {
	Type string `json:"type"`
}

func registerCommands(table *map[string]Command) error {
	if table == nil {
		return errors.New("table is null")
	}

	commands := []Command{
		NoSuchCommand{},
	}

	for _, command := range commands {
		(*table)[command.name()] = command
	}

	return nil
}

func handleRequest(kernel *Kernel, conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(conn)

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println(err)
		return
	}

	var request TypeDto
	err = json.Unmarshal(buffer[:n], &request)
	if err != nil {
		fmt.Println(err)
		return
	}

	command, ok := kernel.runtimeCmdTable[request.Type]
	if !ok {
		command = NoSuchCommand{}
	}
	response, err := command.execute(kernel, buffer[:n])

	_, err = conn.Write(response)
	if err != nil {
		fmt.Println(err)
		return
	}
}
