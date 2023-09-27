package kernel

import (
	json2 "encoding/json"
	"errors"
	"fmt"
	"github.com/dop251/goja"
	"gvisor.dev/gvisor/pkg/abi/linux"
	"gvisor.dev/gvisor/pkg/hostarch"
	util "gvisor.dev/gvisor/pkg/sentry/kernel/callbacks"
	"gvisor.dev/gvisor/pkg/sentry/vfs"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

// HookCallback is signature of DependentHooks that are called from user`s js callback
type HookCallback func(...goja.Value) (interface{}, error)

type HookInfoDto struct {
	// Name contains the jsName
	Name string `json:"name"`

	// Description of the hook
	Description string `json:"description"`

	// Args has such format:
	// argName type description
	Args string `json:"args"`

	// ReturnValue - description of the return value
	ReturnValue string `json:"return-value"`
}

// GoHook is an interface for DependentHooks, that user can call from js callback
type GoHook interface {
	// description should provide ingo about hook in the HookInfoDto
	description() HookInfoDto

	// jsName - with this name the hook will be called from js
	jsName() string
}

// TaskIndependentGoHook is an interface for DependentHooks, that user can call from js callback when cb run with/without task
type TaskIndependentGoHook interface {
	GoHook
	createCallBack() HookCallback
}

// TaskDependentGoHook is an interface for DependentHooks, that user can call from js callback when cb run with task
type TaskDependentGoHook interface {
	GoHook
	createCallBack(*Task) HookCallback
}

// disposableDecorator is used to prevent deadlocks when same callback is called twice
func disposableDecorator(callback HookCallback) HookCallback {
	callbackWasInvoked := false
	return func(args ...goja.Value) (interface{}, error) {
		if callbackWasInvoked {
			panic("this callback should use only one time")
		}

		callbackWasInvoked = true
		return callback(args...)
	}
}

// GoHookDecorator added for future restrictions of DependentHooks
type GoHookDecorator struct {
	wrapped TaskDependentGoHook
}

func (decorator *GoHookDecorator) description() HookInfoDto {
	return decorator.wrapped.description()
}

func (decorator *GoHookDecorator) jsName() string {
	return decorator.wrapped.jsName()
}

func (decorator *GoHookDecorator) createCallBack(t *Task) HookCallback {
	cb := decorator.wrapped.createCallBack(t)
	return disposableDecorator(cb)
}

func (ht *HooksTable) registerDependentHook(hook TaskDependentGoHook) error {
	if ht == nil {
		return errors.New("DependentHooks table is nil")
	}

	ht.mutex.Lock()
	defer ht.mutex.Unlock()

	ht.DependentHooks[hook.jsName()] = hook //&GoHookDecorator{wrapped: hook}
	return nil
}

func (ht *HooksTable) registerIndependentHook(hook TaskIndependentGoHook) error {
	if ht == nil {
		return errors.New("DependentHooks table is nil")
	}

	ht.mutex.Lock()
	defer ht.mutex.Unlock()

	ht.IndependentHooks[hook.jsName()] = hook //&GoHookDecorator{wrapped: hook}
	return nil
}

func (ht *HooksTable) getDependentHook(hookName string) TaskDependentGoHook {
	if ht == nil {
		panic("Hooks table is nil")
	}

	ht.mutex.Lock()
	defer ht.mutex.Unlock()

	f, ok := ht.DependentHooks[hookName]
	if ok {
		return f
	} else {
		return nil
	}
}

func (ht *HooksTable) getCurrentHooks() []GoHook {
	if ht == nil {
		panic("table is nil")
	}

	ht.mutex.Lock()
	defer ht.mutex.Unlock()

	var hooks []GoHook
	for _, hook := range ht.DependentHooks {
		hooks = append(hooks, hook)
	}
	for _, hook := range ht.IndependentHooks {
		hooks = append(hooks, hook)
	}

	return hooks
}

// HooksTable user`s js callback takes DependentHooks from this table before execution.
// Hooks from the table can be used by user in his js code to get / modify data
type HooksTable struct {
	DependentHooks   map[string]TaskDependentGoHook
	IndependentHooks map[string]TaskIndependentGoHook
	mutex            sync.Mutex
}

func ReadBytes(t *Task, addr uintptr, dst []byte) (int, error) {
	return t.CopyInBytes(hostarch.Addr(addr), dst)
}

func WriteBytes(t *Task, addr uintptr, src []byte) (int, error) {
	return t.CopyOutBytes(hostarch.Addr(addr), src)
}

func ReadString(t *Task, addr uintptr, len int) (string, error) {
	return t.CopyInString(hostarch.Addr(addr), len)
}

func WriteString(t *Task, addr uintptr, str string) (int, error) {
	bytes := []byte(str)
	return t.CopyOutBytes(hostarch.Addr(addr), bytes)
}

// SignalMaskGetter return Task.signalMask
// (signals which delivery is blocked)
func SignalMaskGetter(t *Task) uint64 {
	return t.signalMask.Load()
}

// SigWaitMaskGetter provides functions to return Task.realSignalMask
// (Task will be blocked until one of signals in Task.realSignalMask is pending)
func SigWaitMaskGetter(t *Task) uint64 {
	return uint64(t.realSignalMask)
}

// SavedSignalMaskGetter provides functions to return Task.savedSignalMask
func SavedSignalMaskGetter(t *Task) uint64 {
	return uint64(t.savedSignalMask)
}

// SigactionGetter provides functions to return sigactions in JSON format
func SigactionGetter(t *Task) []linux.SigActionDto {
	actions := t.tg.signalHandlers.actions
	var actionsDesc []linux.SigActionDto
	for _, sigaction := range actions {
		actionsDesc = append(actionsDesc, sigaction.ToDto())
	}
	return actionsDesc
}

func GIDGetter(t *Task) uint32 {
	return t.KGID()
}

func UIDGetter(t *Task) uint32 {
	return t.KUID()
}

func PIDGetter(t *Task) int32 {
	return int32(t.PIDNamespace().IDOfTask(t))
}

func EnvvGetter(t *Task) ([]byte, error) {
	mm := t.image.MemoryManager
	envvStart := mm.EnvvStart()
	envvEnd := mm.EnvvEnd()
	size := envvEnd - envvStart
	buf := make([]byte, size)
	_, err := ReadBytes(t, uintptr(envvStart), buf)
	return buf, err
}

// MmapsGetter returns a description of mappings like in procfs
func MmapsGetter(t *Task) string {
	return t.image.MemoryManager.String()
}

func ArgvGetter(t *Task) ([]byte, error) {
	mm := t.image.MemoryManager
	argvStart := mm.ArgvStart()
	argvEnd := mm.ArgvEnd()
	size := argvEnd - argvStart
	buf := make([]byte, size)
	_, err := ReadBytes(t, uintptr(argvStart), buf)
	return buf, err
}

type SessionDto struct {
	SessionId    int32
	PGID         int32
	ForegroundId int32
	OtherPGIDs   []int32
}

// SessionGetter provides info about session:
//
// - session id
//
// - PGID
//
// - foreground
//
// - other PGIDs of the session
func SessionGetter(t *Task) *SessionDto {
	if t.tg == nil {
		return nil
	}
	pg := t.tg.processGroup
	if pg == nil {
		return nil
	}
	var pgids []int32
	if pg.session != nil {
		sessionPGs := pg.session.processGroups
		if &sessionPGs != nil {
			for spg := sessionPGs.Front(); spg != nil; spg = spg.Next() {
				pgids = append(pgids, int32(spg.id))
			}
		}
	}
	if pg.session == nil {
		return nil
	}
	var foregroundGroupId ProcessGroupID
	if t.tg.TTY() == nil {
		t.Debugf("{\"error\": \"%v\"}", "t.tg.TTY() is nil")
		foregroundGroupId = 0
	} else {
		var err error
		foregroundGroupId, err = t.tg.ForegroundProcessGroupID(t.tg.TTY())
		if err != nil {
			t.Debugf("{\"error\": \"%v\"}", err.Error())
		}
	}
	return &SessionDto{
		SessionId:    int32(pg.session.id),
		PGID:         int32(pg.id),
		ForegroundId: int32(foregroundGroupId),
		OtherPGIDs:   pgids,
	}
}

// DependentHooks impls

type PrintHook struct{}

func (ph *PrintHook) description() HookInfoDto {
	return HookInfoDto{
		Name:        ph.jsName(),
		Description: "Prints all passed args",
		Args:        "msgs\t...any\t(values to be printed);\n",
		ReturnValue: "null",
	}
}

func (ph *PrintHook) jsName() string {
	return "print"
}

func (ph *PrintHook) createCallBack() HookCallback {
	return func(args ...goja.Value) (_ interface{}, err error) {
		strs := make([]string, len(args))

		runtime := GetJsRuntime()
		const functionNameInGlobalContext = "stringify"
		stringify, ok := goja.AssertFunction(runtime.JsVM.Get(functionNameInGlobalContext))
		if !ok {
			return nil, errors.New(fmt.Sprintf("failed to load %s", functionNameInGlobalContext))
		}

		for i, arg := range args {
			if arg.ExportType() == reflect.TypeOf("") {
				strs[i] = arg.String()
			} else {
				valueStr, err := stringify(goja.Undefined(), arg)
				if err != nil {
					return nil, err
				}
				strs[i] = valueStr.String()
			}
		}
		_, err = fmt.Print(strings.Join(strs, " "))
		return nil, err
	}
}

type WriteBytesHook struct{}

func (hook *WriteBytesHook) description() HookInfoDto {
	return HookInfoDto{
		Name:        hook.jsName(),
		Description: "Write bytes from provided buffer by provided addr. Always tries to write all bytes from buffer",
		Args: "addr\tnumber\t(data from buffer will be written starting from this addr);\n" +
			"buffer\tArrayBuffer\t(buffer which contains data to be written);\n",
		ReturnValue: "counter\tnumber\t(amount of really written bytes)",
	}
}

func (hook *WriteBytesHook) jsName() string {
	return "writeBytes"
}

func (hook *WriteBytesHook) createCallBack(t *Task) HookCallback {
	return func(args ...goja.Value) (interface{}, error) {

		runtime := GetJsRuntime()
		if len(args) != 2 {
			return nil, util.ArgsCountMismatchError(2, len(args))
		}

		addr, err := util.ExtractPtrFromValue(runtime.JsVM, args[0])
		if err != nil {
			return nil, err
		}

		var buff []byte
		buff, err = util.ExtractByteBufferFromValue(runtime.JsVM, args[1])
		if err != nil {
			return nil, err
		}

		var count int
		count, err = WriteBytes(t, addr, buff)
		if err != nil {
			return nil, err
		}

		return count, nil
	}
}

type ReadBytesHook struct{}

func (hook *ReadBytesHook) description() HookInfoDto {
	return HookInfoDto{
		Name:        hook.jsName(),
		Description: "Read bytes to provided buffer by provided addr. Always tries to read count bytes",
		Args: "addr\tnumber\t(data from address space will be read starting from this addr);\n" +
			"count\tnumber\t(amount of bytes to read from address space);\n",
		ReturnValue: "buffer\tArrayBuffer\t(contains read data)",
	}
}

func (hook *ReadBytesHook) jsName() string {
	return "readBytes"
}

func (hook *ReadBytesHook) createCallBack(t *Task) HookCallback {
	return func(args ...goja.Value) (interface{}, error) {

		runtime := GetJsRuntime()
		if len(args) != 2 {
			return nil, util.ArgsCountMismatchError(2, len(args))
		}

		addr, err := util.ExtractPtrFromValue(runtime.JsVM, args[0])
		if err != nil {
			return nil, err
		}

		var count int64
		count, err = util.ExtractInt64FromValue(runtime.JsVM, args[1])
		if err != nil {
			return nil, err
		}

		buff := make([]byte, count)
		var countRead int
		countRead, err = ReadBytes(t, addr, buff)
		if err != nil {
			return nil, err
		}

		return buff[:countRead], nil
	}
}

type WriteStringHook struct{}

func (hook *WriteStringHook) description() HookInfoDto {
	return HookInfoDto{
		Name:        hook.jsName(),
		Description: "Write provided string by provided addr",
		Args: "addr\tnumber\t(string will be written starting from this addr);\n" +
			"str\tstringt\t(string to be written);\n",
		ReturnValue: "count number (amount of bytes really written)",
	}
}

func (hook *WriteStringHook) jsName() string {
	return "writeString"
}

func (hook *WriteStringHook) createCallBack(t *Task) HookCallback {
	return func(args ...goja.Value) (interface{}, error) {

		runtime := GetJsRuntime()
		if len(args) != 2 {
			return nil, util.ArgsCountMismatchError(2, len(args))
		}

		addr, err := util.ExtractPtrFromValue(runtime.JsVM, args[0])
		if err != nil {
			return nil, err
		}

		var str string
		str, err = util.ExtractStringFromValue(runtime.JsVM, args[1])
		if err != nil {
			return nil, err
		}

		var count int
		count, err = WriteString(t, addr, str)
		if err != nil {
			return nil, err
		}

		return count, nil
	}
}

type ReadStringHook struct{}

func (hook *ReadStringHook) description() HookInfoDto {
	return HookInfoDto{
		Name:        hook.jsName(),
		Description: "Read string str by provided addr",
		Args: "addr\tnumber\t(string will be read starting from this addr);\n" +
			"count\tnumber\t(amount of bytes to read from address space);\n",
		ReturnValue: "str\tstring\t(read string)",
	}
}

func (hook *ReadStringHook) jsName() string {
	return "readString"
}

func (hook *ReadStringHook) createCallBack(t *Task) HookCallback {
	return func(args ...goja.Value) (interface{}, error) {

		runtime := GetJsRuntime()
		if len(args) != 2 {
			return nil, util.ArgsCountMismatchError(2, len(args))
		}

		addr, err := util.ExtractPtrFromValue(runtime.JsVM, args[0])
		if err != nil {
			return nil, err
		}

		var count int64
		count, err = util.ExtractInt64FromValue(runtime.JsVM, args[1])
		if err != nil {
			return nil, err
		}

		var ret string
		ret, err = ReadString(t, addr, int(count))
		if err != nil {
			return nil, err
		}

		return ret, nil
	}
}

type EnvvGetterHook struct{}

func (hook *EnvvGetterHook) description() HookInfoDto {
	return HookInfoDto{
		Name:        hook.jsName(),
		Description: "Provides environment variables of the Task",
		Args:        "no args;\n",
		ReturnValue: "envs\t[]string\t(array of strings, each string has the format ENV_NAME=env_val)",
	}
}

func (hook *EnvvGetterHook) jsName() string {
	return "getEnvs"
}

func (hook *EnvvGetterHook) createCallBack(t *Task) HookCallback {
	return func(args ...goja.Value) (interface{}, error) {

		if len(args) != 0 {
			return nil, util.ArgsCountMismatchError(0, len(args))
		}

		bytes, err := EnvvGetter(t)
		splitStrings := strings.Split(string(bytes), "\x00")
		if err != nil {
			return nil, err
		}

		return splitStrings, nil
	}
}

type MmapGetterHook struct{}

func (hook *MmapGetterHook) description() HookInfoDto {
	//return "Provides mapping info like in procfs"
	return HookInfoDto{
		Name:        hook.jsName(),
		Description: "Provides mapping info like in procfs",
		Args:        "no args;\n",
		ReturnValue: "str\tstring\t(mappings like in procfs)",
	}
}

func (hook *MmapGetterHook) jsName() string {
	return "getMmaps"
}

func (hook *MmapGetterHook) createCallBack(t *Task) HookCallback {
	return func(args ...goja.Value) (interface{}, error) {

		if len(args) != 0 {
			return nil, util.ArgsCountMismatchError(0, len(args))
		}

		res := MmapsGetter(t)
		return res, nil
	}
}

type ArgvHook struct{}

func (hook *ArgvHook) description() HookInfoDto {
	return HookInfoDto{
		Name:        hook.jsName(),
		Description: "Provides argv of the Task",
		Args:        "no args;\n",
		ReturnValue: "argv\t[]string\t(array of strings)",
	}
}

func (hook *ArgvHook) jsName() string {
	return "getArgv"
}

func (hook *ArgvHook) createCallBack(t *Task) HookCallback {
	return func(args ...goja.Value) (interface{}, error) {

		if len(args) != 0 {
			return nil, util.ArgsCountMismatchError(0, len(args))
		}

		bytes, err := ArgvGetter(t)
		splitStrings := strings.Split(string(bytes), "\x00")
		if err != nil {
			return nil, err
		}

		return splitStrings, nil
	}
}

type SignalInfoHook struct{}

func (hook *SignalInfoHook) description() HookInfoDto {
	return HookInfoDto{
		Name:        hook.jsName(),
		Description: "Provides signal masks and sigactions of the Task",
		Args:        "no args;\n",
		ReturnValue: "SignalMaskDto json \n" +
			"{\n" +
			"\tSignalMask number (Task.signalMask signal mask of the task),\n" +
			"\tSignalWaitMask number (Task.realSignalMask (Task will be blocked until one of signals in Task.realSignalMask is pending)),\n" +
			"\tSavedSignalMask number (Task.savedSignalMask (savedSignalMask is the signal mask that should be applied after the task has either delivered one signal to a user handler or is about to resume execution in the untrusted application)),\n" +
			"\tSigActions array of json\n" +
			"\t{\n" +
			"\t\tHandler string,\n" +
			"\t\tFlags string,\n" +
			"\t\tRestorer number,\n" +
			"\t\tMask []string (array of strings, each string is a signal name)\n" +
			"\t}\n" +
			"};\n",
	}
}

func (hook *SignalInfoHook) jsName() string {
	return "getSignalInfo"
}

type SignalMaskDto struct {
	SignalMask      int64
	SignalWaitMask  int64
	SavedSignalMask int64
	SigActions      []linux.SigActionDto
}

func (hook *SignalInfoHook) createCallBack(t *Task) HookCallback {
	return func(args ...goja.Value) (interface{}, error) {

		if len(args) != 0 {
			return nil, util.ArgsCountMismatchError(0, len(args))
		}

		dto := SignalMaskDto{
			SignalMask:      int64(SignalMaskGetter(t)),
			SignalWaitMask:  int64(SigWaitMaskGetter(t)),
			SavedSignalMask: int64(SavedSignalMaskGetter(t)),
			SigActions:      SigactionGetter(t),
		}

		return dto, nil
	}
}

type PidInfoHook struct{}

func (hook *PidInfoHook) description() HookInfoDto {
	return HookInfoDto{
		Name:        hook.jsName(),
		Description: "Provides PID, GID, UID and session info of Task",
		Args:        "no args;\n",
		ReturnValue: "PidDto json \n" +
			"{\n" +
			"\tPID number,\n" +
			"\tGID number,\n" +
			"\tUID number,\n" +
			"\tSession json\n" +
			"\t{\n" +
			"\t\tsessionId number,\n" +
			"\t\tPGID number,\n" +
			"\t\tforeground number,\n" +
			"\t\totherPGIDs []number (array of other PGIDS in session)\n" +
			"\t}\n" +
			"};\n",
	}
}

func (hook *PidInfoHook) jsName() string {
	return "getPidInfo"
}

type PidDto struct {
	PID     int32
	GID     int32
	UID     int32
	Session SessionDto
}

func (hook *PidInfoHook) createCallBack(t *Task) HookCallback {
	return func(args ...goja.Value) (interface{}, error) {

		if len(args) != 0 {
			return nil, util.ArgsCountMismatchError(0, len(args))
		}

		dto := PidDto{
			PID:     PIDGetter(t),
			GID:     int32(GIDGetter(t)),
			UID:     int32(UIDGetter(t)),
			Session: *SessionGetter(t),
		}

		return dto, nil
	}
}

type UserJSONLogHook struct{}

func (hook *UserJSONLogHook) jsName() string {
	return "log"
}

func (hook *UserJSONLogHook) description() HookInfoDto {
	return HookInfoDto{
		Name:        hook.jsName(),
		Description: "Logs the given message",
		Args:        "msg\tany\t(message to be printed)",
		ReturnValue: "null",
	}
}

func (hook *UserJSONLogHook) createCallBack(t *Task) HookCallback {
	return func(args ...goja.Value) (interface{}, error) {
		if len(args) != 1 {
			return nil, util.ArgsCountMismatchError(1, len(args))
		}
		arg := args[0]

		runtime := GetJsRuntime()
		const functionNameInGlobalContext = "stringify"
		stringify, ok := goja.AssertFunction(runtime.JsVM.Get(functionNameInGlobalContext))
		if !ok {
			return nil, errors.New(fmt.Sprintf("failed to load %s", functionNameInGlobalContext))
		}

		var str string
		if arg.ExportType() == reflect.TypeOf("") {
			str = arg.String()
		} else {
			valueStr, err := stringify(goja.Undefined(), arg)
			if err != nil {
				return nil, err
			}
			str = valueStr.String()
		}

		t.JSONInfof(str)
		return nil, nil
	}
}

// RegisterHooks register all hooks from this file in provided table
func RegisterHooks(cb *HooksTable) error {
	dependentGoHooks := []TaskDependentGoHook{
		&ReadBytesHook{},
		&WriteBytesHook{},
		&ReadStringHook{},
		&WriteStringHook{},
		&EnvvGetterHook{},
		&MmapGetterHook{},
		&ArgvHook{},
		&SignalInfoHook{},
		&PidInfoHook{},
		&FDHook{},
		&FDsHook{},
		&UserJSONLogHook{},
	}

	independentGoHooks := []TaskIndependentGoHook{
		&PrintHook{},
		&AddCbBeforeHook{},
		&AddCbAfterHook{},
	}

	for _, hook := range dependentGoHooks {
		err := cb.registerDependentHook(hook)
		if err != nil {
			return err
		}
	}

	for _, hook := range independentGoHooks {
		err := cb.registerIndependentHook(hook)
		if err != nil {
			return err
		}
	}

	return nil
}

type FDsHook struct{}

func (hook *FDsHook) description() HookInfoDto {
	return HookInfoDto{
		Name:        hook.jsName(),
		Description: "Provides information about all fds of Task",
		Args:        "no args;\n",
		ReturnValue: "dto ArrayBuffer (marshalled array of json (format below))\n" +
			"{\n" +
			"\tfd string,\n" +
			"\tname string,\n" +
			"\tmode string,\n" +
			"\tflags string, \n" +
			"\tnlinks string,\n" +
			"\treadable boolean,\n" +
			"\twritable boolean,\n" +
			"};\n",
	}
}

func (hook *FDsHook) jsName() string {
	return "getFdsInfo"
}

func (hook *FDsHook) createCallBack(t *Task) HookCallback {
	return func(args ...goja.Value) (interface{}, error) {
		if len(args) != 0 {
			return nil, util.ArgsCountMismatchError(0, len(args))
		}

		dto := FdsResolver(t)

		return dto, nil
	}
}

type FDHook struct{}

func (hook *FDHook) description() HookInfoDto {
	return HookInfoDto{
		Name:        hook.jsName(),
		Description: "Provides information about one specific fd of Task",
		Args:        "fd number (fd to get info about);\n",
		ReturnValue: "dto ArrayBuffer (marshalled json (format below))\n" +
			"{\n" +
			"\tfd string,\n" +
			"\tname string,\n" +
			"\tmode string,\n" +
			"\tnlinks string,\n" +
			"\tflags string,\n" +
			"\treadable boolean,\n" +
			"\twritable boolean,\n" +
			"};\n",
	}
}

func (hook *FDHook) jsName() string {
	return "getFdInfo"
}

func (hook *FDHook) createCallBack(t *Task) HookCallback {
	return func(args ...goja.Value) (interface{}, error) {
		if len(args) != 1 {
			return nil, util.ArgsCountMismatchError(1, len(args))
		}

		runtime := GetJsRuntime()
		val, err := util.ExtractInt64FromValue(runtime.JsVM, args[0])
		if err != nil {
			return nil, err
		}

		fd := int32(val)

		dto := FdResolver(t, fd)

		return string(dto), nil
	}
}

type FDInfo struct {
	Path     string `json:"path"`
	FD       string `json:"fd"`
	Mode     string `json:"mode"`
	Nlinks   string `json:"nlinks"`
	Flags    string `json:"flags"`
	Readable bool   `json:"readable"`
	Writable bool   `json:"writable"`
}

// FdsResolver resolves all file descriptors that belong to given task and returns
// path to fd, fd num and fd mask in JSON format
func FdsResolver(t *Task) []byte {
	jsonPrivs := make([]FDInfo, 5)

	fdt := t.fdTable

	fdt.forEach(t, func(fd int32, fdesc *vfs.FileDescription, _ FDFlags) {
		stat, err := fdesc.Stat(t, vfs.StatOptions{})
		if err != nil {
			return
		}

		name := findPath(t, fd)
		num := strconv.FormatInt(int64(fd), 10)
		nlinks := strconv.FormatInt(int64(stat.Nlink), 10)
		flags := parseAttributesMask(stat.AttributesMask)

		jsonPrivs = append(jsonPrivs, FDInfo{
			FD:       num,
			Path:     name,
			Mode:     parseMask(uint16(linux.FileMode(stat.Mode).Permissions())),
			Nlinks:   nlinks,
			Flags:    flags,
			Writable: fdesc.IsWritable(),
			Readable: fdesc.IsReadable(),
		})
	})

	jsonForm, _ := json2.Marshal(jsonPrivs)

	return jsonForm
}

// FdResolver resolves one specific fd for given task and returns
// path to fd, fd num and fd mask in JSON format
func FdResolver(t *Task, fd int32) []byte {
	fdesc, _ := t.fdTable.Get(fd)
	if fdesc == nil {
		return nil
	}
	defer fdesc.DecRef(t)
	stat, err := fdesc.Stat(t, vfs.StatOptions{})
	if err != nil {
		return nil
	}

	name := findPath(t, fd)
	num := strconv.FormatInt(int64(fd), 10)
	nlinks := strconv.FormatInt(int64(stat.Nlink), 10)

	jsonPrivs := FDInfo{
		Path:     name,
		FD:       num,
		Mode:     parseMask(uint16(linux.FileMode(stat.Mode).Permissions())),
		Nlinks:   nlinks,
		Writable: fdesc.IsWritable(),
		Readable: fdesc.IsReadable(),
	}

	jsonForm, _ := json2.Marshal(jsonPrivs)

	return jsonForm
}

// findPath resolves fd's path in virtual file system
func findPath(t *Task, fd int32) string {
	root := t.FSContext().RootDirectory()
	defer root.DecRef(t)

	vfsobj := root.Mount().Filesystem().VirtualFilesystem()
	file := t.GetFile(fd)
	defer file.DecRef(t)

	name, _ := vfsobj.PathnameInFilesystem(t, file.VirtualDentry())

	return name
}

// parseMask parses fd's mask into readable format
// Example: rwx---r--
func parseMask(mask uint16) string {
	perm := ""
	for i := 0; i < 9; i++ {
		if mask&(1<<uint16(i)) != 0 {
			if i%3 == 0 {
				perm += "x"
			} else if i%3 == 1 {
				perm += "w"
			} else {
				perm += "r"
			}
		} else {
			perm += "-"
		}
	}

	// before reverseString perm is reversed because of algorithm above
	perm = reverseString(perm)

	return perm
}

// parseAttributesMask is a helper function for
// FDs and FD resolvers that parses attribute mask of fd
func parseAttributesMask(mask uint64) string {
	s := linux.OpenMode.Parse(mask & linux.O_ACCMODE)
	if flags := linux.OpenFlagSet.Parse(mask &^ linux.O_ACCMODE); flags != "" {
		s += "|" + flags
	}

	return s
}

// reverseString is helping function for parseMask that
// reverses given string: bazel -> lezab
func reverseString(str string) string {
	runes := []rune(str)
	reversed := make([]rune, len(runes))
	for i, j := 0, len(runes)-1; i <= j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = runes[j], runes[i]
	}
	return string(reversed)
}

// addDependentHooksToContextObject from this context object user`s callback will take DependentHooks
func (ht *HooksTable) addDependentHooksToContextObject(object *goja.Object, task *Task) error {
	ht.mutex.Lock()
	defer ht.mutex.Unlock()

	for name, hook := range ht.DependentHooks {
		callback := hook.createCallBack(task)
		err := object.Set(name, callback)
		if err != nil {
			return err
		}
	}

	return nil
}

// DependentHookAddableAdapter implement ContextAddable
type DependentHookAddableAdapter struct {
	ht   *HooksTable
	task *Task
}

func (d *DependentHookAddableAdapter) addSelfToContextObject(object *goja.Object) error {
	return d.ht.addDependentHooksToContextObject(object, d.task)
}

// addIndependentHooksToContextObject from this context object user`s callback will take DependentHooks
func (ht *HooksTable) addIndependentHooksToContextObject(object *goja.Object) error {
	ht.mutex.Lock()
	defer ht.mutex.Unlock()

	for name, hook := range ht.IndependentHooks {
		callback := hook.createCallBack()
		err := object.Set(name, callback)
		if err != nil {
			return err
		}
	}

	return nil
}

// IndependentHookAddableAdapter implement ContextAddable
type IndependentHookAddableAdapter struct {
	ht *HooksTable
}

func (d *IndependentHookAddableAdapter) addSelfToContextObject(object *goja.Object) error {
	return d.ht.addIndependentHooksToContextObject(object)
}

// hooks for dynamic callback registration

type AddCbBeforeHook struct{}

func (a AddCbBeforeHook) description() HookInfoDto {
	return HookInfoDto{
		Name:        a.jsName(),
		Description: "Is used for dynamic callback registration (callback will be executed before syscall)",
		Args: "sysno\tnumber\t(syscall number, callback will be executed before syscall with this number)\n" +
			"callback\tfunction\t(js function to call before syscall execution)",
		ReturnValue: "null",
	}
}

func (a AddCbBeforeHook) jsName() string {
	return "AddCbBefore"
}

func (a AddCbBeforeHook) createCallBack() HookCallback {
	return func(args ...goja.Value) (interface{}, error) {
		if len(args) != 2 {
			return nil, util.ArgsCountMismatchError(2, len(args))
		}

		runtime := GetJsRuntime()
		sysno, err := util.ExtractPtrFromValue(runtime.JsVM, args[0])
		if err != nil {
			return nil, err
		}

		cbObj := args[1].ToObject(runtime.JsVM)
		table := runtime.callbackTable

		err = table.registerCallbackBefore(sysno, &DynamicJsCallbackBefore{Holder: cbObj})
		return nil, err
	}
}

type AddCbAfterHook struct{}

func (a AddCbAfterHook) description() HookInfoDto {
	return HookInfoDto{
		Name:        a.jsName(),
		Description: "Is used for dynamic callback registration (callback will be executed after syscall)",
		Args: "sysno\tnumber\t(syscall number, callback will be executed after syscall with this number)\n" +
			"callback\tfunction\t(js function to call after syscall execution)",
		ReturnValue: "null",
	}
}

func (a AddCbAfterHook) jsName() string {
	return "AddCbAfter"
}

func (a AddCbAfterHook) createCallBack() HookCallback {
	return func(args ...goja.Value) (interface{}, error) {
		if len(args) != 2 {
			return nil, util.ArgsCountMismatchError(2, len(args))
		}

		runtime := GetJsRuntime()
		sysno, err := util.ExtractPtrFromValue(runtime.JsVM, args[0])
		if err != nil {
			return nil, err
		}

		cbObj := args[1].ToObject(runtime.JsVM)
		table := runtime.callbackTable

		err = table.registerCallbackAfter(sysno, &DynamicJsCallbackAfter{Holder: cbObj})
		return nil, err
	}
}
