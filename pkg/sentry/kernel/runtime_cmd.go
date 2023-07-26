package kernel

import (
	"encoding/json"
	"errors"
	"fmt"
	"gvisor.dev/gvisor/pkg/sentry/kernel/callbacks"
	"log"
	"net"
	"sync"
)

// Command is the interface used to configure hooks
type Command interface {
	name() string

	// execute method get bytes of request and return bytes of response / error
	execute(kernel *Kernel, raw []byte) (any, error)
}

type CommandTable struct {
	commands map[string]Command
	mutex    sync.Mutex
}

func (table *CommandTable) Register(command Command) error {
	if table == nil {
		return errors.New("table is null")
	}
	if table.commands == nil {
		return errors.New("commands map is uninitialized")
	}
	if command == nil {
		return errors.New("command in nil")
	}

	table.mutex.Lock()
	defer table.mutex.Unlock()

	table.commands[command.name()] = command
	return nil
}

func (table *CommandTable) GetCommand(name string) (Command, error) {
	if table == nil {
		return nil, errors.New("table is null")
	}
	table.mutex.Lock()
	defer table.mutex.Unlock()

	command, ok := table.commands[name]
	if !ok {
		return nil, errors.New(fmt.Sprintf("command %s not exists", name))
	}
	return command, nil
}

type Response struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Payload any    `json:"payload"`
}

func messageResponse(type_ string, message string) []byte {
	return []byte(fmt.Sprintf("{\"type\": \"%s\", \"message\": \"%s\"}", type_, message))
}

const ResponseTypeError = "error"
const ResponseTypeOk = "ok"

func registerCommands(table *CommandTable) error {
	commands := []Command{
		&ChangeSyscallCallbackCommand{},
		&GetHooksInfoCommand{},
		&ChangeStateCommand{},
		&CallbacksListCommand{},
		&UnregisterCallbacksCommand{},
	}

	for _, command := range commands {
		err := table.Register(command)
		if err != nil {
			return err
		}
	}

	return nil
}

type jsonRequest map[string]interface{}

const typeKey = "type"
const payloadKey = "payload"

func extractTypeAndPayload(request *jsonRequest) (string, []byte, error) {
	typeAny, ok := (*request)[typeKey]
	if !ok {
		return "", nil, errors.New(fmt.Sprintf("request not contains %s field", typeKey))
	}
	typeString, ok := typeAny.(string)
	if !ok {
		return "", nil, errors.New("type field in request should be string")
	}

	payloadAny, ok := (*request)[payloadKey]
	if !ok {
		return "", nil, errors.New(fmt.Sprintf("request not contains %s", payloadKey))
	}
	payloadBytes, err := json.Marshal(payloadAny)
	if err != nil {
		return "", nil, err
	}

	return typeString, payloadBytes, nil
}

func handleRequest(kernel *Kernel, jsonDecoder *json.Decoder) ([]byte, error) {
	var request jsonRequest
	err := jsonDecoder.Decode(&request)
	fmt.Println("request err", err, request)
	if err != nil {
		return nil, err
	}
	requestType, payloadBytes, err := extractTypeAndPayload(&request)
	if err != nil {
		return nil, err
	}

	table := kernel.runtimeCmdTable
	command, err := table.GetCommand(requestType)
	if err != nil {
		return nil, err
	}

	responsePayload, err := command.execute(kernel, payloadBytes)
	if err != nil {
		return nil, err
	}
	response := Response{
		Type:    ResponseTypeOk,
		Message: "Everything ok",
		Payload: responsePayload,
	}
	responseBytes, err := json.Marshal(&response)

	return responseBytes, err
}

func writeToConn(conn net.Conn, content []byte) error {
	for len(content) > 0 {
		n, err := conn.Write(content)
		if err != nil {
			return err
		}
		content = content[n:]
	}

	return nil
}

func handleConnection(kernel *Kernel, conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println(err)
		}
	}(conn)

	jsonDecoder := json.NewDecoder(conn)
	response, err := handleRequest(kernel, jsonDecoder)
	if err != nil {
		response = messageResponse(ResponseTypeError, fmt.Sprintf("error: %s", err.Error()))
	}

	err = writeToConn(conn, response)
	if err != nil {
		log.Println(err)
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

func (c ChangeSyscallCallbackCommand) execute(kernel *Kernel, raw []byte) (any, error) {

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

	return nil, nil
}

// hooks info command

type HooksInfoCommandResponse struct {
	HooksInfo []HookInfoDto `json:"hooks"`
}

type GetHooksInfoCommand struct{}

func (g GetHooksInfoCommand) name() string {
	return "change-info" // Bruh specification moment
}

func (g GetHooksInfoCommand) execute(kernel *Kernel, _ []byte) (any, error) {
	var hookInfoDtos []HookInfoDto

	table := kernel.hooksTable
	table.mutex.Lock()
	defer table.mutex.Unlock()

	for _, hook := range table.hooks {
		hookInfoDtos = append(hookInfoDtos, hook.description())
	}

	response := HooksInfoCommandResponse{HooksInfo: hookInfoDtos}
	return response, nil
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

func (c ChangeStateCommand) execute(_ *Kernel, raw []byte) (any, error) {

	return nil, errors.New("change state command not implemented yet")
	//var request ChangeStateRequest
	//err := json.Unmarshal(raw, &request)
	//if err != nil {
	//	return nil, err
	//}
	//
	//if request.EntryPoint == "" || request.Source == "" {
	//	return nil, errors.New("script source or/and entry point is empty")
	//}
	//
	//fmt.Println(request)
	//
	//// TODO implements (after adding persistence state)
	//
	//return nil, nil
}

// get current callbacks

type CallbackListResponse struct {
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

func (c CallbacksListCommand) execute(kernel *Kernel, _ []byte) (any, error) {
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

	response := CallbackListResponse{JsCallbacks: infos}
	return response, nil
}

// unregister callbacks cmd

const UnregisterAllOption = "all"
const UnregisterListOption = "list"

type UnregisterCallbackDto struct {
	Sysno int    `json:"sysno"`
	Type  string `json:"type"`
}

type UnregisterCallbacksRequest struct {
	Options string                  `json:"options"`
	List    []UnregisterCallbackDto `json:"list"`
}

type UnregisterCallbacksCommand struct{}

func (u UnregisterCallbacksCommand) name() string {
	return "unregister-callbacks"
}

func executeListOption(table *CallbackTable, request *UnregisterCallbacksRequest) error {
	for _, dto := range request.List {
		switch dto.Type {
		case JsCallbackTypeBefore:
			err := table.unregisterCallbackBefore(uintptr(dto.Sysno))
			if err != nil {
				return err
			}

		case JsCallbackTypeAfter:
			err := table.unregisterCallbackAfter(uintptr(dto.Sysno))
			if err != nil {
				return err
			}

		default:
			return errors.New(fmt.Sprintf("unknown callback type [%s]", dto.Type))
		}
	}

	return nil
}

func executeAllOption(table *CallbackTable, _ *UnregisterCallbacksRequest) error {
	table.UnregisterAll()
	return nil
}

func (u UnregisterCallbacksCommand) execute(kernel *Kernel, raw []byte) (any, error) {
	var request UnregisterCallbacksRequest
	err := json.Unmarshal(raw, &request)
	if err != nil {
		return nil, err
	}
	table := kernel.callbackTable

	switch request.Options {
	case UnregisterAllOption:
		err := executeAllOption(table, &request)
		if err != nil {
			return nil, err
		}

	case UnregisterListOption:
		err := executeListOption(table, &request)
		if err != nil {
			return nil, err
		}

	default:
		return nil, errors.New(fmt.Sprintf("unknown options [%s]", request.Options))
	}

	return nil, nil
}
