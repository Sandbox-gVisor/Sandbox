package kernel

import (
	"errors"
	"fmt"
	"github.com/dop251/goja"
	"gvisor.dev/gvisor/pkg/abi/linux"
	util "gvisor.dev/gvisor/pkg/sentry/kernel/callbacks"
	"reflect"
	"strings"
)

// DependentHooks impls

type PrintHook struct{}

func (ph *PrintHook) description() HookInfoDto {
	return HookInfoDto{
		Name:        ph.jsName(),
		Description: "Prints all passed args",
		Args:        "\nmsgs\t...any\t(values to be printed);\n",
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
		Args: "\naddr\tnumber\t(data from buffer will be written starting from this addr);\n" +
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
		Args: "\naddr\tnumber\t(data from address space will be read starting from this addr);\n" +
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
		Args: "\naddr\tnumber\t(string will be written starting from this addr);\n" +
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
		Args: "\naddr\tnumber\t(string will be read starting from this addr);\n" +
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
		Args:        "\nno args;\n",
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
		Args:        "\nno args;\n",
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
		Args:        "\nno args;\n",
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
		Args:        "\nno args;\n",
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
		Args:        "\nno args;\n",
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
	Session SessionDTO
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
		Args:        "\nmsg\tany\t(message to be printed);\n",
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

type FDsHook struct{}

func (hook *FDsHook) description() HookInfoDto {
	return HookInfoDto{
		Name:        hook.jsName(),
		Description: "Provides information about all fds of Task",
		Args:        "\nno args;\n",
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
		Args:        "\nfd\tnumber\t(fd to get info about);\n",
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

type AnonMmapHook struct{}

func (m AnonMmapHook) description() HookInfoDto {
	return HookInfoDto{
		Name:        m.jsName(),
		Description: "Creates new anonymous mapping in the virtual address space of the calling process",
		Args:        "\nlength\tnumber\t(amount of bytes to allocate);\n",
		ReturnValue: "addr\tnumber\t()",
	}
}

func (m AnonMmapHook) jsName() string {
	return "anonMmap"
}

func (m AnonMmapHook) createCallBack(t *Task) HookCallback {
	return func(args ...goja.Value) (interface{}, error) {
		runtime := GetJsRuntime()
		if len(args) != 1 {
			return nil, util.ArgsCountMismatchError(1, len(args))
		}

		length, err := util.ExtractInt64FromValue(runtime.JsVM, args[0])
		if err != nil {
			return nil, err
		}

		rval, err := AnonMmap(t, uintptr(length))
		if err != nil {
			return nil, err
		}

		return int64(rval), nil
	}
}

type MunmapHook struct{}

func (m MunmapHook) description() HookInfoDto {
	return HookInfoDto{
		Name:        m.jsName(),
		Description: "Delete the mappings from the specified address range",
		Args: "\naddr\tnumber\t(start address, must be a multiple of the page size);\n" +
			"length\tnumber\t(amount of bytes to set range);\n",
		ReturnValue: "null",
	}
}

func (m MunmapHook) jsName() string {
	return "munmap"
}

func (m MunmapHook) createCallBack(t *Task) HookCallback {
	return func(args ...goja.Value) (interface{}, error) {
		runtime := GetJsRuntime()
		if len(args) != 2 {
			return nil, util.ArgsCountMismatchError(2, len(args))
		}

		addr, err := util.ExtractInt64FromValue(runtime.JsVM, args[0])
		if err != nil {
			return nil, err
		}

		var length int64
		length, err = util.ExtractInt64FromValue(runtime.JsVM, args[1])
		if err != nil {
			return nil, err
		}

		err = Unmap(t, uintptr(addr), uintptr(length))
		return nil, err
	}
}

// hooks for dynamic callback registration

type AddCbBeforeHook struct{}

func (a AddCbBeforeHook) description() HookInfoDto {
	return HookInfoDto{
		Name:        a.jsName(),
		Description: "Is used for dynamic callback registration (callback will be executed before syscall)",
		Args: "\nsysno\tnumber\t(syscall number, callback will be executed before syscall with this number);\n" +
			"callback\tfunction\t(js function to call before syscall execution);\n",
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
		Args: "\nsysno\tnumber\t(syscall number, callback will be executed after syscall with this number);\n" +
			"callback\tfunction\t(js function to call after syscall execution);\n",
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
