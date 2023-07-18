package kernel

import (
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

func GIDGetterProvider(t *Task) func() uint32 {
	return func() uint32 {
		return t.KGID()
	}
}

func UIDGetterProvider(t *Task) func() uint32 {
	return func() uint32 {
		return t.KUID()
	}
}

func PIDGetterProvider(t *Task) func() int32 {
	return func() int32 {
		return int32(t.childPIDNamespace.IDOfTask(t))
	}
}
