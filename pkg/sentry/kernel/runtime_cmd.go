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
		&CallbacksListCommand{},
		&UnregisterCallbacksCommand{},
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
		hookInfoDtos = append(hookInfoDtos, hook.description())
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

// get current callbacks

type CallbackListResponse struct {
	Type        string                     `json:"type"`
	JsCallbacks []callbacks.JsCallbackInfo `json:"callbacks"`
}

type CallbacksListCommand struct{}

func (c CallbacksListCommand) name() string {
	return "current-callbacks"
}

func unknownCallback(sysno uintptr, cbType string) *callbacks.JsCallbackInfo {
	return &callbacks.JsCallbackInfo{
		Sysno:          int(sysno),
		EntryPoint:     "unknown",
		CallbackSource: "unknown",
		Type:           cbType,
	}
}

func (c CallbacksListCommand) execute(kernel *Kernel, _ []byte) ([]byte, error) {
	table := kernel.callbackTable
	table.Lock()
	defer table.Unlock()
	var infos []callbacks.JsCallbackInfo

	for sysno, cbBefore := range table.callbackBefore {
		info, err := callbacks.JsCallbackInfoFromStr(cbBefore.Info())
		if err != nil {
			info = unknownCallback(sysno, JsCallbackTypeBefore)
		}
		infos = append(infos, *info)
	}

	for sysno, cbAfter := range table.callbackAfter {
		info, err := callbacks.JsCallbackInfoFromStr(cbAfter.Info())
		if err != nil {
			info = unknownCallback(sysno, JsCallbackTypeAfter)
		}
		infos = append(infos, *info)
	}

	response := CallbackListResponse{Type: ResponseTypeOk, JsCallbacks: infos}
	bytes, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

// unregister callbacks cmd

type UnregisterCallbackDto struct {
	Sysno int    `json:"sysno"`
	Type  string `json:"type"`
}

type UnregisterCallbacksRequest struct {
	List []UnregisterCallbackDto `json:"list"`
}

type UnregisterCallbacksCommand struct{}

func (u UnregisterCallbacksCommand) name() string {
	return "unregister-callbacks"
}

func (u UnregisterCallbacksCommand) execute(kernel *Kernel, raw []byte) ([]byte, error) {
	var request UnregisterCallbacksRequest
	err := json.Unmarshal(raw, &request)
	if err != nil {
		return nil, err
	}

	table := kernel.callbackTable
	for _, dto := range request.List {
		switch dto.Type {
		case JsCallbackTypeBefore:
			err := table.unregisterCallbackBefore(uintptr(dto.Sysno))
			if err != nil {
				return nil, err
			}

		case JsCallbackTypeAfter:
			err := table.unregisterCallbackAfter(uintptr(dto.Sysno))
			if err != nil {
				return nil, err
			}

		default:
			return nil, errors.New(fmt.Sprintf("unknown callback type %s", dto.Type))
		}
	}

	return messageResponse(ResponseTypeOk, "All callbacks in list disabled"), nil
}
