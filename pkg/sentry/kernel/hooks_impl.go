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

// dependentHooks impls

type PrintHook struct{}

func (ph *PrintHook) description() HookInfoDto {
	return HookInfoDto{
		Name:        ph.jsName(),
		Description: "Prints all passed args",
		Args:        "\nmsgs\t...any\t(values to be printed);\n",
		ReturnValue: "null\n",
	}
}

func (ph *PrintHook) jsName() string {
	return "print"
}

func (ph *PrintHook) createCallback() HookCallback {
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
		ReturnValue: "counter\tnumber\t(amount of really written bytes)\n",
	}
}

func (hook *WriteBytesHook) jsName() string {
	return "writeBytes"
}

func (hook *WriteBytesHook) createCallback(t *Task) HookCallback {
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
		ReturnValue: "buffer\tArrayBuffer\t(contains read data)\n",
	}
}

func (hook *ReadBytesHook) jsName() string {
	return "readBytes"
}

func (hook *ReadBytesHook) createCallback(t *Task) HookCallback {
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
			"str\tstring\t(string to be written);\n",
		ReturnValue: "count number (amount of bytes really written)\n",
	}
}

func (hook *WriteStringHook) jsName() string {
	return "writeString"
}

func (hook *WriteStringHook) createCallback(t *Task) HookCallback {
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
		ReturnValue: "str\tstring\t(read string)\n",
	}
}

func (hook *ReadStringHook) jsName() string {
	return "readString"
}

func (hook *ReadStringHook) createCallback(t *Task) HookCallback {
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
		ReturnValue: "envs\t[]string\t(array of strings, each string has the format ENV_NAME=env_val)\n",
	}
}

func (hook *EnvvGetterHook) jsName() string {
	return "getEnvs"
}

func (hook *EnvvGetterHook) createCallback(t *Task) HookCallback {
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
		ReturnValue: "str\tstring\t(mappings like in procfs)\n",
	}
}

func (hook *MmapGetterHook) jsName() string {
	return "getMmaps"
}

func (hook *MmapGetterHook) createCallback(t *Task) HookCallback {
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
		ReturnValue: "argv\t[]string\t(array of strings)\n",
	}
}

func (hook *ArgvHook) jsName() string {
	return "getArgv"
}

func (hook *ArgvHook) createCallback(t *Task) HookCallback {
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
			"\tsignalMask number (Task.signalMask signal mask of the task),\n" +
			"\tsignalWaitMask number (Task.realSignalMask (Task will be blocked until one of signals in Task.realSignalMask is pending)),\n" +
			"\tsavedSignalMask number (Task.savedSignalMask (savedSignalMask is the signal mask that should be applied after the task has either delivered one signal to a user handler or is about to resume execution in the untrusted application)),\n" +
			"\tsigActions array of json\n" +
			"\t{\n" +
			"\t\thandler string,\n" +
			"\t\tflags string,\n" +
			"\t\trestorer number,\n" +
			"\t\tsignalsInSet []string (array of strings, each string is a signal name)\n" +
			"\t}\n" +
			"};\n",
	}
}

func (hook *SignalInfoHook) jsName() string {
	return "getSignalInfo"
}

type SignalMaskDto struct {
	SignalMask      int64                `json:"signalMask"`
	SignalWaitMask  int64                `json:"signalWaitMask"`
	SavedSignalMask int64                `json:"savedSignalMask"`
	SigActions      []linux.SigActionDto `json:"sigActions"`
}

func (hook *SignalInfoHook) createCallback(t *Task) HookCallback {
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
			"\tsession json\n" +
			"\t{\n" +
			"\t\tsessionID number,\n" +
			"\t\tPGID number,\n" +
			"\t\tforegroundID number,\n" +
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
	Session SessionDTO `json:"session"`
}

func (hook *PidInfoHook) createCallback(t *Task) HookCallback {
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
	return "logJson"
}

func (hook *UserJSONLogHook) description() HookInfoDto {
	return HookInfoDto{
		Name:        hook.jsName(),
		Description: "Logs the given message",
		Args:        "\nmsg\tany\t(message to be logged);\n",
		ReturnValue: "null\n",
	}
}

func (hook *UserJSONLogHook) createCallback(t *Task) HookCallback {
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
		ReturnValue: "dtos []object (array of file description dtos)\n" +
			"{\n" +
			"\tfd number,\n" +
			"\tname string,\n" +
			"\tmode string,\n" +
			"\tflags string, \n" +
			"\tnlinks number,\n" +
			"\treadable boolean,\n" +
			"\twritable boolean,\n" +
			"};\n",
	}
}

func (hook *FDsHook) jsName() string {
	return "getFdsInfo"
}

func (hook *FDsHook) createCallback(t *Task) HookCallback {
	return func(args ...goja.Value) (interface{}, error) {
		if len(args) != 0 {
			return nil, util.ArgsCountMismatchError(0, len(args))
		}

		dto, err := FdsResolver(t)
		if err != nil {
			return nil, err
		}

		return dto, nil
	}
}

type FDHook struct{}

func (hook *FDHook) description() HookInfoDto {
	return HookInfoDto{
		Name:        hook.jsName(),
		Description: "Provides information about one specific fd of Task",
		Args:        "\nfd\tnumber\t(fd to get info about);\n",
		ReturnValue: "dto object (file description dto (format see below))\n" +
			"{\n" +
			"\tfd number,\n" +
			"\tname string,\n" +
			"\tmode string,\n" +
			"\tnlinks number,\n" +
			"\tflags string,\n" +
			"\treadable boolean,\n" +
			"\twritable boolean,\n" +
			"};\n",
	}
}

func (hook *FDHook) jsName() string {
	return "getFdInfo"
}

func (hook *FDHook) createCallback(t *Task) HookCallback {
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

		dto, err := FdResolver(t, fd)
		if err != nil {
			return nil, err
		}

		return dto, nil
	}
}

type AnonMmapHook struct{}

func (m AnonMmapHook) description() HookInfoDto {
	return HookInfoDto{
		Name:        m.jsName(),
		Description: "Creates new anonymous mapping in the virtual address space of the calling process",
		Args:        "\nlength\tnumber\t(amount of bytes to allocate);\n",
		ReturnValue: "addr\tnumber\t(memory region start address)\n",
	}
}

func (m AnonMmapHook) jsName() string {
	return "anonMmap"
}

func (m AnonMmapHook) createCallback(t *Task) HookCallback {
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
		ReturnValue: "null\n",
	}
}

func (m MunmapHook) jsName() string {
	return "munmap"
}

func (m MunmapHook) createCallback(t *Task) HookCallback {
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

type SignalByNameHook struct{}

func (s SignalByNameHook) description() HookInfoDto {
	return HookInfoDto{
		Name:        s.jsName(),
		Description: "Returns number of signal by given signal name",
		Args:        "\nname\tstring\t(name of signal to get value);\n",
		ReturnValue: "sig\tnumber\t(the number of the signal or -1 if signal with such name doesn't exist)\n",
	}
}

func (s SignalByNameHook) jsName() string {
	return "nameToSignal"
}

func (s SignalByNameHook) createCallback() HookCallback {
	return func(args ...goja.Value) (interface{}, error) {
		runtime := GetJsRuntime()
		if len(args) != 1 {
			return nil, util.ArgsCountMismatchError(1, len(args))
		}

		name, err := util.ExtractStringFromValue(runtime.JsVM, args[0])
		if err != nil {
			return nil, err
		}

		sig, err := linux.GetSignalByName(name)
		if err != nil {
			return nil, err
		}

		return int64(sig), nil
	}
}

type SignalMaskToSignalNamesHook struct{}

func (s SignalMaskToSignalNamesHook) description() HookInfoDto {
	return HookInfoDto{
		Name:        s.jsName(),
		Description: "Returns array of names of signals in mask",
		Args:        "\nmask\tnumber\t(signal mask);\n",
		ReturnValue: "names\t[]string\t(names of signals in mask)\n",
	}
}

func (s SignalMaskToSignalNamesHook) jsName() string {
	return "signalMaskToNames"
}

func (s SignalMaskToSignalNamesHook) createCallback() HookCallback {
	return func(args ...goja.Value) (interface{}, error) {
		runtime := GetJsRuntime()
		if len(args) != 1 {
			return nil, util.ArgsCountMismatchError(2, len(args))
		}

		mask, err := util.ExtractInt64FromValue(runtime.JsVM, args[0])
		if err != nil {
			return nil, err
		}

		sigMask := linux.SignalSet(uint64(mask))
		return sigMask.Signals(), nil
	}
}

type SignalSendingHook struct{}

func (s SignalSendingHook) description() HookInfoDto {
	return HookInfoDto{
		Name:        s.jsName(),
		Description: "Sends the given signal to task with given pid",
		Args: "\npid\tnumber\t(pid of the task to send signal);\n" +
			"signo\tnumber\t(the number of the signal to send);\n",
		ReturnValue: "null\n",
	}
}

func (s SignalSendingHook) jsName() string {
	return "sendSignal"
}

func (s SignalSendingHook) createCallback(t *Task) HookCallback {
	return func(args ...goja.Value) (interface{}, error) {

		runtime := GetJsRuntime()
		if len(args) != 2 {
			return nil, util.ArgsCountMismatchError(2, len(args))
		}

		pid, err := util.ExtractInt64FromValue(runtime.JsVM, args[0])
		if err != nil {
			return nil, err
		}

		var signo int64
		signo, err = util.ExtractInt64FromValue(runtime.JsVM, args[1])
		if err != nil {
			return nil, err
		}

		sig := linux.Signal(signo)

		err = SendSignalToTaskWithID(t, ThreadID(int32(pid)), sig)
		if err != nil {
			return nil, err
		}

		return nil, nil
	}
}

// ThreadsStoppingHook is used for stopping all threads except the caller
type ThreadsStoppingHook struct{}

func (hook *ThreadsStoppingHook) description() HookInfoDto {
	return HookInfoDto{
		Name:        hook.jsName(),
		Description: "Stops all threads except the caller. May be useful for preventing TOCTOU attack.",
		Args:        "\nno args;\n",
		ReturnValue: "null\n",
	}
}

func (hook *ThreadsStoppingHook) jsName() string {
	return "stopThreads"
}

func (hook *ThreadsStoppingHook) createCallback(t *Task) HookCallback {
	return func(args ...goja.Value) (interface{}, error) {
		t.stopOtherThreadsInTg()
		return nil, nil
	}
}

// ThreadsResumingHook should be used after ThreadsStoppingHook to resume stopped threads
type ThreadsResumingHook struct{}

func (hook *ThreadsResumingHook) description() HookInfoDto {
	return HookInfoDto{
		Name:        hook.jsName(),
		Description: "Resume threads stopped by `stopThreads`.",
		Args:        "\nno args;\n",
		ReturnValue: "null\n",
	}
}

func (hook *ThreadsResumingHook) jsName() string {
	return "resumeThreads"
}

func (hook *ThreadsResumingHook) createCallback(t *Task) HookCallback {
	return func(args ...goja.Value) (interface{}, error) {
		t.resumeOtherThreadsInTg()
		return nil, nil
	}
}

type ThreadInfoHook struct{}

func (hook *ThreadInfoHook) description() HookInfoDto {
	return HookInfoDto{
		Name:        hook.jsName(),
		Description: "Provides TID, TGID (PID) and list of other TIDs in thread group.",
		Args: "\nno args;\n" +
			"or\n" +
			"tid\tnumber\t(thread id to get info);\n",
		ReturnValue: "ThreadInfoDto json \n" +
			"{\n" +
			"\tTID number,\n" +
			"\tTGID number (same as PID of process to which thread is related),\n" +
			"\tTIDsInTg []number (array of other PGIDS in session)\n" +
			"};\n",
	}
}

func (hook *ThreadInfoHook) jsName() string {
	return "getThreadInfo"
}

type ThreadInfoDto struct {
	TID      int32
	TGID     int32
	TIDsInTg []int32 `json:"TIDsInTg"`
}

func (hook *ThreadInfoHook) createCallback(t *Task) HookCallback {
	return func(args ...goja.Value) (interface{}, error) {

		if len(args) > 1 {
			return nil, util.ArgsCountMismatchError(1, len(args))
		} else if len(args) == 0 {
			return fillThreadInfoDto(t), nil
		} else {
			runtime := GetJsRuntime()
			val, err := util.ExtractInt64FromValue(runtime.JsVM, args[0])
			if err != nil {
				return nil, err
			}
			tid := ThreadID(val)
			task := t.tg.pidns.tasks[tid]
			if task == nil {
				return nil, fmt.Errorf("no such task")
			}
			return fillThreadInfoDto(task), nil
		}
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
		ReturnValue: "null\n",
	}
}

func (a AddCbBeforeHook) jsName() string {
	return "AddCbBefore"
}

func (a AddCbBeforeHook) createCallback() HookCallback {
	return func(args ...goja.Value) (interface{}, error) {
		if len(args) != 2 {
			return nil, util.ArgsCountMismatchError(2, len(args))
		}

		runtime := GetJsRuntime()
		sysno, err := util.ExtractPtrFromValue(runtime.JsVM, args[0])
		if err != nil {
			return nil, err
		}

		if goja.IsNull(args[1]) || goja.IsUndefined(args[1]) {
			return nil, util.ErrNullOrUndefined
		}

		cbObj := args[1].ToObject(runtime.JsVM)
		table := runtime.callbackTable

		info := *unknownCallback(sysno, JsCallbackTypeBefore)
		info = fillJsCallbackInfoForDynamicCallback(info, cbObj.String())

		err = table.registerCallbackBefore(sysno, &DynamicJsCallbackBefore{CallbackInfo: info, Holder: cbObj})
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
		ReturnValue: "null\n",
	}
}

func (a AddCbAfterHook) jsName() string {
	return "AddCbAfter"
}

func (a AddCbAfterHook) createCallback() HookCallback {
	return func(args ...goja.Value) (interface{}, error) {
		if len(args) != 2 {
			return nil, util.ArgsCountMismatchError(2, len(args))
		}

		runtime := GetJsRuntime()
		sysno, err := util.ExtractPtrFromValue(runtime.JsVM, args[0])
		if err != nil {
			return nil, err
		}

		if goja.IsNull(args[1]) || goja.IsUndefined(args[1]) {
			return nil, util.ErrNullOrUndefined
		}

		cbObj := args[1].ToObject(runtime.JsVM)
		table := runtime.callbackTable

		info := *unknownCallback(sysno, JsCallbackTypeAfter)
		info = fillJsCallbackInfoForDynamicCallback(info, cbObj.String())

		err = table.registerCallbackAfter(sysno, &DynamicJsCallbackAfter{CallbackInfo: info, Holder: cbObj})
		return nil, err
	}
}
