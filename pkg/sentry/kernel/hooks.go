package kernel

import (
	"fmt"
	"github.com/dop251/goja"
	"gvisor.dev/gvisor/pkg/hostarch"
	util "gvisor.dev/gvisor/pkg/sentry/kernel/callbacks"
	"strings"
)

func ReadBytesHook(t *Task, addr uintptr, dst []byte) (int, error) {
	return t.CopyInBytes(hostarch.Addr(addr), dst)
}

func WriteBytesHook(t *Task, addr uintptr, src []byte) (int, error) {
	return t.CopyOutBytes(hostarch.Addr(addr), src)
}

func ReadBytesProvider(t *Task) func(addr uintptr, dst []byte) (int, error) {
	return func(addr uintptr, dst []byte) (int, error) {
		return t.CopyInBytes(hostarch.Addr(addr), dst)
	}
}

func WriteBytesProvider(t *Task) func(addr uintptr, src []byte) (int, error) {
	return func(addr uintptr, src []byte) (int, error) {
		return t.CopyOutBytes(hostarch.Addr(addr), src)
	}
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

func EnvvGetterProvider(t *Task) func() ([]byte, error) {
	return func() ([]byte, error) {
		mm := t.image.MemoryManager
		envvStart := mm.EnvvStart()
		envvEnd := mm.EnvvEnd()
		size := envvEnd - envvStart
		buf := make([]byte, size)
		_, err := ReadBytesHook(t, uintptr(envvStart), buf)
		return buf, err
	}
}

func MmapsGetterProvider(t *Task) func() string {
	return func() string {
		return t.image.MemoryManager.String()
	}
}

func ArgvGetterProvider(t *Task) func() ([]byte, error) {
	return func() ([]byte, error) {
		mm := t.image.MemoryManager
		argvStart := mm.ArgvStart()
		argvEnd := mm.ArgvEnd()
		size := argvEnd - argvStart
		buf := make([]byte, size)
		_, err := ReadBytesHook(t, uintptr(argvStart), buf)
		return buf, err
	}
}

// hooks impls

type PrintHook struct {
}

func (ph *PrintHook) description() string {
	return "default"
}

func (ph *PrintHook) jsName() string {
	return "print"
}

func (ph *PrintHook) createCallBack(t *Task) HookCallback {
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

		runtime := t.Kernel().V8Go
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

		runtime := t.Kernel().V8Go
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

		runtime := t.Kernel().V8Go
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

		runtime := t.Kernel().V8Go
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

func RegisterHooks(cb *HooksTable) error {
	hooks := []GoHook{
		&PrintHook{},
		&ReadBytesHookImpl{},
		&WriteBytesHookImpl{},
		&ReadStringHookImpl{},
		&WriteStringHookImpl{},
	}

	for _, hook := range hooks {
		err := cb.registerHook(hook)
		if err != nil {
			return err
		}
	}

	return nil
}
