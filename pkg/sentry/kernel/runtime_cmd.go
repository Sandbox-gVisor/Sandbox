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

// Command is the interface used to configure dependentHooks
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
	if err != nil {
		return nil, err
	}
	requestType, payloadBytes, err := extractTypeAndPayload(&request)
	if err != nil {
		return nil, err
	}

	table := GetJsRuntime().runtimeCmdTable
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
		response = messageResponse(ResponseTypeError, err.Error())
	}

	err = writeToConn(conn, response)
	if err != nil {
		log.Println(err)
	}
}

// dependentHooks info command

type HooksInfoCommandResponse struct {
	HooksInfo []HookInfoDto `json:"hooks"`
}

type GetHooksInfoCommand struct{}

func (g GetHooksInfoCommand) name() string {
	return "hooks-info" // Bruh specification moment
}

func (g GetHooksInfoCommand) execute(_ *Kernel, _ []byte) (any, error) {
	var hookInfoDtos []HookInfoDto

	runtime := GetJsRuntime()
	table := runtime.hooksTable

	hooks := table.getCurrentHooks()
	for _, hook := range hooks {
		hookInfoDtos = append(hookInfoDtos, hook.description())
	}

	response := HooksInfoCommandResponse{HooksInfo: hookInfoDtos}
	return response, nil
}

// change state command

type ChangeStateRequestDto struct {
	Source string `json:"source"`
}

type ChangeStateCommand struct{}

func (c ChangeStateCommand) name() string {
	return "change-state"
}

func (c ChangeStateCommand) execute(_ *Kernel, raw []byte) (any, error) {
	var request ChangeStateRequestDto
	err := json.Unmarshal(raw, &request)
	if err != nil {
		return nil, err
	}
	err = callbacks.CheckSyntaxError(request.Source)
	if err != nil {
		return nil, err
	}

	runtime := GetJsRuntime()
	runtime.Mutex.Lock()
	defer runtime.Mutex.Unlock()

	builder := ScriptContextsBuilderOf()
	builder = builder.AddContext3(HooksJsName, &IndependentHookAddableAdapter{ht: runtime.hooksTable})
	builder = builder.AddContext3(JsPersistenceContextName,
		&ObjectAddableAdapter{name: JsGlobalPersistenceObject, object: runtime.Global})

	contexts := builder.Build()
	val, err := RunJsScript(runtime.JsVM, request.Source, contexts)
	if err != nil {
		return nil, err
	}
	valBytes, err := json.Marshal(val)
	return string(valBytes), err
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
		CallbackBody:   "unknown",
		CallbackArgs:   make([]string, 0),
		Type:           cbType,
	}
}

func (c CallbacksListCommand) execute(_ *Kernel, _ []byte) (any, error) {
	table := GetJsRuntime().callbackTable
	table.rwLockBefore.Lock()
	table.rwLockAfter.Lock()

	defer table.rwLockAfter.Unlock()
	defer table.rwLockBefore.Unlock()

	var infos []callbacks.JsCallbackInfo

	for _, cbBefore := range table.callbackBefore {
		info := cbBefore.Info()
		infos = append(infos, info)
	}

	for _, cbAfter := range table.callbackAfter {
		info := cbAfter.Info()
		infos = append(infos, info)
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

func (u UnregisterCallbacksCommand) execute(_ *Kernel, raw []byte) (any, error) {
	var request UnregisterCallbacksRequest
	err := json.Unmarshal(raw, &request)
	if err != nil {
		return nil, err
	}
	table := GetJsRuntime().callbackTable

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
