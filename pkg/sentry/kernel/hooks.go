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
