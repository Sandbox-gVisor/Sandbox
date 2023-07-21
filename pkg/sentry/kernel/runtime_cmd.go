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

const ResponseTypeError = "error"
const ResponseTypeOk = "ok"

func registerCommands(table *map[string]Command) error {
	if table == nil {
		return errors.New("table is null")
	}

	commands := []Command{
		&ChangeSyscallCallbackCommand{},
		&GetHooksInfoCommand{},
		&ChangeStateCommand{},
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
		response = messageResponse(ResponseTypeError, err.Error())
	} else {
		command, ok := kernel.runtimeCmdTable[request.Type]

		if !ok {
			response = messageResponse(ResponseTypeError, "no such command: "+request.Type)
		} else {
			response, err = command.execute(kernel, buffer[:n])
			if err != nil {
				response = messageResponse(ResponseTypeError, err.Error())
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

// change callbacks cmd

type ChangeSyscallCallbackCommand struct{}

func (c ChangeSyscallCallbackCommand) name() string {
	return "change-callbacks"
}

type ChangeSyscallDto struct {
	Type        string                     `json:"type"`
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

	var jsCallbacks []JsCallback
	for _, dto := range request.CallbackDto {
		jsCallback, err := JsCallbackByInfo(dto)
		if err != nil {
			return nil, err
		}
		jsCallbacks = append(jsCallbacks, jsCallback)
	}

	for _, cb := range jsCallbacks {
		cbCopy := cb // DON'T touch or golang will do trash
		err := cbCopy.registerAtCallbackTable(kernel.callbackTable)
		if err != nil {
			return nil, err
		}
	}

	return messageResponse(ResponseTypeOk, "Everything is OK"), nil
}

// hooks info command

type HookInfoDto struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Args        string `json:"args"`
	ReturnValue string `json:"return-value"`
}

type HooksInfoCommandResponse struct {
	Type      string        `json:"type"`
	HooksInfo []HookInfoDto `json:"hooks"`
}

type GetHooksInfoCommand struct{}

func (g GetHooksInfoCommand) name() string {
	return "change-info" // Bruh specification moment
}

func (g GetHooksInfoCommand) execute(kernel *Kernel, _ []byte) ([]byte, error) {
	var hookInfoDtos []HookInfoDto

	table := kernel.hooksTable
	table.mutex.Lock()
	defer table.mutex.Unlock()

	for _, hook := range table.hooks {
		hookInfoDtos = append(hookInfoDtos, HookInfoDto{Description: hook.description()})
	}

	response := HooksInfoCommandResponse{Type: ResponseTypeOk, HooksInfo: hookInfoDtos}
	bytes, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

// change state command

type ChangeStateRequest struct {
	EntryPoint string `json:"entry-point"`
	Source     string `json:"source"`
}

type ChangeStateCommand struct{}

func (c ChangeStateCommand) name() string {
	return "change-state"
}

func (c ChangeStateCommand) execute(kernel *Kernel, raw []byte) ([]byte, error) {
	var request ChangeStateRequest
	err := json.Unmarshal(raw, &request)
	if err != nil {
		return nil, err
	}

	if request.EntryPoint == "" || request.Source == "" {
		return nil, errors.New("script source or/and entry point is empty")
	}

	fmt.Println(request)

	// TODO implements (after adding persistence state)

	return messageResponse(ResponseTypeOk, "stub"), nil
}

//
