package callbacks

import (
	"gvisor.dev/gvisor/pkg/hostarch"
	"gvisor.dev/gvisor/pkg/sentry/kernel"
)

func ReadBytesHook(t *kernel.Task, addr uintptr, dst []byte) (int, error) {
	return t.CopyInBytes(hostarch.Addr(addr), dst)
}

func WriteBytesHook(t *kernel.Task, addr uintptr, src []byte) (int, error) {
	return t.CopyOutBytes(hostarch.Addr(addr), src)
}

func ReadBytesProvider(t *kernel.Task) func(addr uintptr, dst []byte) (int, error) {
	return func(addr uintptr, dst []byte) (int, error) {
		return t.CopyInBytes(hostarch.Addr(addr), dst)
	}
}

func WriteBytesProvider(t *kernel.Task) func(addr uintptr, src []byte) (int, error) {
	return func(addr uintptr, src []byte) (int, error) {
		return t.CopyOutBytes(hostarch.Addr(addr), src)
	}
}
