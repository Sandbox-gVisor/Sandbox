package kernel

import (
	json2 "encoding/json"
	"fmt"
	"github.com/dop251/goja"
	"gvisor.dev/gvisor/pkg/hostarch"
	util "gvisor.dev/gvisor/pkg/sentry/kernel/callbacks"
	"gvisor.dev/gvisor/pkg/sentry/vfs"
	"strconv"
	"strings"
	"sync"
)

// HookCallback is signature of hooks that are called from user`s js callback
type HookCallback func(...goja.Value) (interface{}, error)

// GoHook is an interface for hooks, that user can call from js callback
type GoHook interface {
	description() string

	jsName() string

	createCallBack(*Task) HookCallback
}

// HooksTable user`s js callback takes hooks from this table before execution
type HooksTable struct {
	hooks map[string]GoHook
	mutex sync.Mutex
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
func SigactionGetter(t *Task) string {
	actions := t.tg.signalHandlers.actions
	var actionsDesc []string
	for _, sigaction := range actions {
		actionsDesc = append(actionsDesc, sigaction.String())
	}
	return fmt.Sprintf("[%v]", strings.Join(actionsDesc, ",\n"))
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

func SessionGetter(t *Task) string {
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

// hooks impls

type PrintHook struct{}

func (ph *PrintHook) description() string {
	return "Prints passed arguments"
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

type WriteBytesHookImpl struct{}

func (hook *WriteBytesHookImpl) description() string {
	return "Write bytes from provided buffer by provided addr. Always tries to write all bytes from buffer"
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
		count, err = WriteBytes(t, addr, buff)
		if err != nil {
			return nil, err
		}

		return count, nil
	}
}

type ReadBytesHookImpl struct{}

func (hook *ReadBytesHookImpl) description() string {
	return "Read bytes to provided buffer by provided addr. Always tries to read len(buffer) bytes"
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
		countRead, err = ReadBytes(t, addr, buff)
		if err != nil {
			return nil, err
		}

		return buff[:countRead], nil
	}
}

type WriteStringHookImpl struct{}

func (hook *WriteStringHookImpl) description() string {
	return "Write provided string by provided addr"
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

		var count int
		count, err = WriteString(t, addr, str)
		if err != nil {
			return nil, err
		}

		return count, nil
	}
}

type ReadStringHookImpl struct{}

func (hook *ReadStringHookImpl) description() string {
	return "Read string by provided addr"
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

		var ret string
		ret, err = ReadString(t, addr, int(count))
		if err != nil {
			return nil, err
		}

		return ret, nil
	}
}

type EnvvGetterHookImpl struct{}

func (hook *EnvvGetterHookImpl) description() string {
	return "Provides environment variables of the Task"
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
	return "Provides mapping info like in procfs"
}

func (hook *MmapGetterHookImpl) jsName() string {
	return "getMmaps"
}

func (hook *MmapGetterHookImpl) createCallBack(t *Task) HookCallback {
	return func(args ...goja.Value) (interface{}, error) {

		if len(args) != 0 {
			return nil, util.ArgsCountMismatchError(0, len(args))
		}

		res := MmapsGetter(t)
		return res, nil
	}
}

type ArgvHookImpl struct{}

func (hook *ArgvHookImpl) description() string {
	return "Provides argv of the Task"
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
	return "Provides signal masks and sigactions of the Task"
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
			SignalMask:      int64(SignalMaskGetter(t)),
			SignalWaitMask:  int64(SigWaitMaskGetter(t)),
			SavedSignalMask: int64(SavedSignalMaskGetter(t)),
			SigActions:      SigactionGetter(t),
		}

		return dto, nil
	}
}

type PidHook struct{}

func (hook *PidHook) description() string {
	return "Provides PID, GID, UID and session info of Task"
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
			Session: SessionGetter(t),
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
		&FDHook{},
		&FDsHook{},
	}

	for _, hook := range hooks {
		err := cb.registerHook(hook)
		if err != nil {
			return err
		}
	}

	return nil
}

type FDsHook struct{}

func (hook *FDsHook) description() string {
	return "Provides information about all fds of Task"
}

func (hook *FDsHook) jsName() string {
	return "getFdsInfo"
}

func (hook *FDsHook) createCallBack(t *Task) HookCallback {
	return func(args ...goja.Value) (interface{}, error) {
		if len(args) != 0 {
			return nil, util.ArgsCountMismatchError(0, len(args))
		}

		if len(args) != 0 {
			return nil, util.ArgsCountMismatchError(0, len(args))
		}

		dto := FdsResolver(t)

		return dto, nil
	}
}

type FDHook struct{}

func (hook *FDHook) description() string {
	return "Provides information about one specific fd of Task"
}

func (hook *FDHook) jsName() string {
	return "getFdInfo"
}

func (hook *FDHook) createCallBack(t *Task) HookCallback {
	return func(args ...goja.Value) (interface{}, error) {
		if len(args) != 1 {
			return nil, util.ArgsCountMismatchError(0, len(args))
		}

		runtime := t.Kernel().GojaRuntime
		if len(args) != 2 {
			return nil, util.ArgsCountMismatchError(2, len(args))
		}

		val, err := util.ExtractInt64FromValue(runtime, args[0])
		if err != nil {
			return nil, err
		}

		fd := int32(val)

		dto := FdResolver(t, fd)

		return dto, nil
	}
}

type FDInfo struct {
	Path     string `json:"name"`
	FD       string `json:"fd"`
	Mode     string `json:"mode"`
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
		privMask := parseMask(stat.Mode)

		jsonPrivs = append(jsonPrivs, FDInfo{
			FD:       num,
			Path:     name,
			Mode:     privMask,
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
	privMask := parseMask(stat.Mode)

	jsonPrivs := FDInfo{
		Path:     name,
		FD:       num,
		Mode:     privMask,
		Writable: fdesc.IsWritable(),
		Readable: fdesc.IsReadable(),
	}

	jsonForm, _ := json2.Marshal(jsonPrivs)

	return jsonForm
}

func findPath(t *Task, fd int32) string {
	root := t.FSContext().RootDirectory()
	defer root.DecRef(t)

	vfsobj := root.Mount().Filesystem().VirtualFilesystem()
	file := t.GetFile(fd)
	defer file.DecRef(t)

	name, _ := vfsobj.PathnameInFilesystem(t, file.VirtualDentry())

	return name
}

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

	perm = reverseString(perm)

	return perm
}

func reverseString(str string) string {
	runes := []rune(str)
	reversed := make([]rune, len(runes))
	for i, j := 0, len(runes)-1; i <= j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = runes[j], runes[i]
	}
	return string(reversed)
}
