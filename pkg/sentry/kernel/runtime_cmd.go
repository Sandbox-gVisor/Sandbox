package kernel

import (
	"encoding/json"
	"errors"
	"fmt"
	"gvisor.dev/gvisor/pkg/sentry/kernel/callbacks"
	"net"
)

type Command interface {
	name() string

	execute(kernel *Kernel, raw []byte) ([]byte, error)
}

func messageResponse(type_ string, message string) []byte {
	return []byte(fmt.Sprintf("{\"type\": \"%s\", \"message\": \"%s\"}", type_, message))
}

type TypeDto struct {
	Type string `json:"type"`
}

type ChangeSyscallCallbackCommand struct{}

func (c ChangeSyscallCallbackCommand) name() string {
	return "change-callbacks"
}

type ChangeSyscallDto struct {
	Type string `json:"type"`

	CallbackDto []callbacks.JsCallbackInfo `json:"callbacks"`
}

func (c ChangeSyscallCallbackCommand) execute(kernel *Kernel, raw []byte) ([]byte, error) {

	var request ChangeSyscallDto
	err := json.Unmarshal(raw, &request)
	if err != nil {
		return nil, err
	}

	if len(request.CallbackDto) == 0 {
		return nil, errors.New("callbacks list is empty")
	}

	var jsCallbacksBefore []JsCallbackBefore
	for _, dto := range request.CallbackDto {
		if dto.Type == JsCallbackTypeBefore {
			cb := JsCallbackBefore{info: dto}
			err := checkJsCallback(&cb)
			if err != nil {
				return nil, err
			}
			jsCallbacksBefore = append(jsCallbacksBefore, cb)
		}
	}

	kernel.callbackTable.mutexBefore.Lock()
	defer kernel.callbackTable.mutexBefore.Unlock()

	for _, cb := range jsCallbacksBefore {
		cbCopy := cb // DON'T touch or golang will do trash
		err = kernel.callbackTable.registerCallbackBeforeNoLock(uintptr(cb.info.Sysno), &cbCopy)

		if err != nil {
			panic(err)
		}
	}

	return messageResponse("ok", "Everything is OK"), nil
}

func registerCommands(table *map[string]Command) error {
	if table == nil {
		return errors.New("table is null")
	}

	commands := []Command{
		&ChangeSyscallCallbackCommand{},
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

	//TODO так делать не круто
	buffer := make([]byte, 1<<15)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println(err)
		return
	}

	var request TypeDto
	var response []byte

	// очень плохо не понятно, исправить вложенность
	err = json.Unmarshal(buffer[:n], &request)
	if err != nil {
		response = messageResponse("error", err.Error())
	} else {
		command, ok := kernel.runtimeCmdTable[request.Type]

		if !ok {
			response = messageResponse("error", "no such command: "+request.Type)
		} else {
			response, err = command.execute(kernel, buffer[:n])
			if err != nil {
				response = messageResponse("error", err.Error())
			}
		}
	}

	for len(response) > 0 {
		n, err = conn.Write(response)
		if err != nil {
			fmt.Println(err)
			return
		}
		response = response[n:]
	}
}
