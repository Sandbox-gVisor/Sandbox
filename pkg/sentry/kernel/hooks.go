package kernel

import (
	"fmt"
	"github.com/dop251/goja"
	"gvisor.dev/gvisor/pkg/hostarch"
	util "gvisor.dev/gvisor/pkg/sentry/kernel/callbacks"
	"strconv"
	"strings"
)

func ReadBytesHook(t *Task, addr uintptr, dst []byte) (int, error) {
	return t.CopyInBytes(hostarch.Addr(addr), dst)
}

func WriteBytesHook(t *Task, addr uintptr, src []byte) (int, error) {
	return t.CopyOutBytes(hostarch.Addr(addr), src)
}

func ReadStringProvider(t *Task) func(addr uintptr, len int) (string, error) {
	return func(addr uintptr, length int) (string, error) {
		return t.CopyInString(hostarch.Addr(addr), length)
	}
}

func WriteStringProvider(t *Task) func(addr uintptr, str string) (int, error) {
	return func(addr uintptr, str string) (int, error) {
		bytes := []byte(str)
		return t.CopyOutBytes(hostarch.Addr(addr), bytes)
	}
}

// SignalMaskProvider provides functions to return Task.signalMask
// (signals which delivery is blocked)
func SignalMaskProvider(t *Task) func() uint64 {
	return func() uint64 {
		return t.signalMask.Load()
	}
}

// SigWaitMaskProvider provides functions to return Task.realSignalMask
// (Task will be blocked until one of signals in Task.realSignalMask is pending)
func SigWaitMaskProvider(t *Task) func() uint64 {
	return func() uint64 {
		return uint64(t.realSignalMask)
	}
}

// SavedSignalMaskProvider provides functions to return Task.savedSignalMask
func SavedSignalMaskProvider(t *Task) func() uint64 {
	return func() uint64 {
		return uint64(t.savedSignalMask)
	}
}

// SigactionGetterProvider provides functions to return sigactions in JSON format
func SigactionGetterProvider(t *Task) func() string {
	return func() string {
		actions := t.tg.signalHandlers.actions
		var actionsDesc []string
		for _, sigaction := range actions {
			actionsDesc = append(actionsDesc, sigaction.String())
		}
		return fmt.Sprintf("[%v]", strings.Join(actionsDesc, ",\n"))
	}
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
	_, err := ReadBytesHook(t, uintptr(envvStart), buf)
	return buf, err
}

func MmapsGetterProvider(t *Task) func() string {
	return func() string {
		return t.image.MemoryManager.String()
	}
}

func ArgvGetter(t *Task) ([]byte, error) {
	mm := t.image.MemoryManager
	argvStart := mm.ArgvStart()
	argvEnd := mm.ArgvEnd()
	size := argvEnd - argvStart
	buf := make([]byte, size)
	_, err := ReadBytesHook(t, uintptr(argvStart), buf)
	return buf, err
}

func SessionGetterProvider(t *Task) func() string {
	return func() string {
		if t.tg == nil {
			return fmt.Sprintf("{\"error\": \"%v\"}", "thread group is nil")
		}
		pg := t.tg.processGroup
		if pg == nil {
			return fmt.Sprintf("{\"error\": \"%v\"}", "process group is nil")
		}
		var pgids []string
		if pg.session != nil {
			sessionPGs := pg.session.processGroups
			if &sessionPGs != nil {
				for spg := sessionPGs.Front(); spg != nil; spg = spg.Next() {
					pgids = append(pgids, strconv.Itoa(int(spg.id)))
				}
			}
		}
		if pg.session == nil {
			return fmt.Sprintf("{\"error\": \"%v\"}", "session is nil")
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
		return fmt.Sprintf("{\"sessionId\": %v, \"PGID\": %v, \"foreground\": %v, \"otherPGIDs\": [%v]}", pg.session.id, pg.id, foregroundGroupId, strings.Join(pgids, ", "))
	}
}

// hooks impls

type PrintHook struct {
}

func (ph *PrintHook) description() string {
	return "Prints passed args"
}

func (ph *PrintHook) jsName() string {
	return "print"
}

func (ph *PrintHook) createCallBack(_ *Task) HookCallback {
	return func(args ...goja.Value) (_ interface{}, err error) {
		//map в go не завезли?
		strs := make([]string, len(args))
		for i, arg := range args {
			strs[i] = arg.String()
		}
		_, err = fmt.Println(strings.Join(strs, " "))
		return nil, err
	}
}

type WriteBytesHookImpl struct {
}

func (hook *WriteBytesHookImpl) description() string {
	return "default"
}

func (hook *WriteBytesHookImpl) jsName() string {
	return "writeBytes"
}

func (hook *WriteBytesHookImpl) createCallBack(t *Task) HookCallback {
	return func(args ...goja.Value) (interface{}, error) {

		runtime := t.Kernel().GojaRuntime
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
		count, err = WriteBytesHook(t, addr, buff)
		if err != nil {
			return nil, err
		}

		return count, nil
	}
}

type ReadBytesHookImpl struct {
}

func (hook *ReadBytesHookImpl) description() string {
	return "default"
}

func (hook *ReadBytesHookImpl) jsName() string {
	return "readBytes"
}

func (hook *ReadBytesHookImpl) createCallBack(t *Task) HookCallback {
	return func(args ...goja.Value) (interface{}, error) {

		runtime := t.Kernel().GojaRuntime
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
		countRead, err = ReadBytesHook(t, addr, buff)
		if err != nil {
			return nil, err
		}

		return buff[:countRead], nil
	}
}

type WriteStringHookImpl struct {
}

func (hook *WriteStringHookImpl) description() string {
	return "default"
}

func (hook *WriteStringHookImpl) jsName() string {
	return "writeString"
}

func (hook *WriteStringHookImpl) createCallBack(t *Task) HookCallback {
	return func(args ...goja.Value) (interface{}, error) {

		runtime := t.Kernel().GojaRuntime
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

		cb := WriteStringProvider(t)
		var count int
		count, err = cb(addr, str)
		if err != nil {
			return nil, err
		}

		return count, nil
	}
}

type ReadStringHookImpl struct {
}

func (hook *ReadStringHookImpl) description() string {
	return "default"
}

func (hook *ReadStringHookImpl) jsName() string {
	return "readString"
}

func (hook *ReadStringHookImpl) createCallBack(t *Task) HookCallback {
	return func(args ...goja.Value) (interface{}, error) {

		runtime := t.Kernel().GojaRuntime
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

		cb := ReadStringProvider(t)
		var ret string
		ret, err = cb(addr, int(count))
		if err != nil {
			return nil, err
		}

		return ret, nil
	}
}

type EnvvGetterHookImpl struct {
}

func (hook *EnvvGetterHookImpl) description() string {
	return "default"
}

func (hook *EnvvGetterHookImpl) jsName() string {
	return "getEnvs"
}

func (hook *EnvvGetterHookImpl) createCallBack(t *Task) HookCallback {
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

type MmapGetterHookImpl struct{}

func (hook *MmapGetterHookImpl) description() string {
	return "default"
}

func (hook *MmapGetterHookImpl) jsName() string {
	return "getMmaps"
}

func (hook *MmapGetterHookImpl) createCallBack(t *Task) HookCallback {
	return func(args ...goja.Value) (interface{}, error) {

		if len(args) != 0 {
			return nil, util.ArgsCountMismatchError(0, len(args))
		}

		res := MmapsGetterProvider(t)()
		return res, nil
	}
}

type ArgvHookImpl struct{}

func (hook *ArgvHookImpl) description() string {
	return "default"
}

func (hook *ArgvHookImpl) jsName() string {
	return "getArgv"
}

func (hook *ArgvHookImpl) createCallBack(t *Task) HookCallback {
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

type SignalMaskHook struct{}

func (hook *SignalMaskHook) description() string {
	return "default"
}

func (hook *SignalMaskHook) jsName() string {
	return "getSignalInfo"
}

type SignalMaskDto struct {
	SignalMask      int64
	SignalWaitMask  int64
	SavedSignalMask int64
	SigActions      string
}

func (hook *SignalMaskHook) createCallBack(t *Task) HookCallback {
	return func(args ...goja.Value) (interface{}, error) {

		if len(args) != 0 {
			return nil, util.ArgsCountMismatchError(0, len(args))
		}

		dto := SignalMaskDto{
			SignalMask:      int64(SignalMaskProvider(t)()),
			SignalWaitMask:  int64(SigWaitMaskProvider(t)()),
			SavedSignalMask: int64(SavedSignalMaskProvider(t)()),
			SigActions:      SigactionGetterProvider(t)(),
		}

		return dto, nil
	}
}

type PidHook struct{}

func (hook *PidHook) description() string {
	return "default"
}

func (hook *PidHook) jsName() string {
	return "getPidInfo"
}

type PidDto struct {
	Pid     int32
	Gid     int32
	Uid     int32
	Session string
}

func (hook *PidHook) createCallBack(t *Task) HookCallback {
	return func(args ...goja.Value) (interface{}, error) {

		if len(args) != 0 {
			return nil, util.ArgsCountMismatchError(0, len(args))
		}

		dto := PidDto{
			Pid:     PIDGetter(t),
			Gid:     int32(GIDGetter(t)),
			Uid:     int32(UIDGetter(t)),
			Session: SessionGetterProvider(t)(),
		}

		return dto, nil
	}
}

// RegisterHooks register all hooks from this file in provided table
func RegisterHooks(cb *HooksTable) error {
	hooks := []GoHook{
		&PrintHook{},
		&ReadBytesHookImpl{},
		&WriteBytesHookImpl{},
		&ReadStringHookImpl{},
		&WriteStringHookImpl{},
		&EnvvGetterHookImpl{},
		&MmapGetterHookImpl{},
		&ArgvHookImpl{},
		&SignalMaskHook{},
		&PidHook{},
	}

	for _, hook := range hooks {
		err := cb.registerHook(hook)
		if err != nil {
			return err
		}
	}

	return nil
}
