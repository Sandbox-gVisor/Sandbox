package kernel

import (
	json2 "encoding/json"
	"errors"
	"github.com/dop251/goja"
	"gvisor.dev/gvisor/pkg/abi/linux"
	"gvisor.dev/gvisor/pkg/hostarch"
	"gvisor.dev/gvisor/pkg/sentry/vfs"
	"strconv"
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

type SessionDTO struct {
	SessionID    int32
	PGID         int32
	ForegroundID int32
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
func SessionGetter(t *Task) *SessionDTO {
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
		for spg := sessionPGs.Front(); spg != nil; spg = spg.Next() {
			pgids = append(pgids, int32(spg.id))
		}
	}
	if pg.session == nil {
		return nil
	}
	var foregroundGroupID ProcessGroupID
	if t.tg.TTY() == nil {
		t.Debugf("{\"error\": \"%v\"}", "t.tg.TTY() is nil")
		foregroundGroupID = 0
	} else {
		var err error
		foregroundGroupID, err = t.tg.ForegroundProcessGroupID(t.tg.TTY())
		if err != nil {
			t.Debugf("{\"error\": \"%v\"}", err.Error())
		}
	}
	return &SessionDTO{
		SessionID:    int32(pg.session.id),
		PGID:         int32(pg.id),
		ForegroundID: int32(foregroundGroupID),
		OtherPGIDs:   pgids,
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
