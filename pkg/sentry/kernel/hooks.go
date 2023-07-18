package kernel

import (
	"errors"
	"gvisor.dev/gvisor/pkg/hostarch"
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

func GIDGetterProvider(t *Task) (func() uint32, error) {
	if t == nil {
		return nil, errors.New("task is nil")
	}

	return func() uint32 {
		return t.KGID()
	}, nil
}

func UIDGetterProvider(t *Task) (func() uint32, error) {
	if t == nil {
		return nil, errors.New("task is nil")
	}

	return func() uint32 {
		return t.KUID()
	}, nil
}

func PIDGetterProvider(t *Task) (func() int32, error) {
	if t == nil {
		return nil, errors.New("task is nil")
	}

	return func() int32 {
		return int32(t.PIDNamespace().IDOfTask(t))
	}, nil
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
