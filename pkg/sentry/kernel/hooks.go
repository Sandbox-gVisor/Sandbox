package kernel

import (
	"errors"
	"fmt"
	"gvisor.dev/gvisor/pkg/hostarch"
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
		return fmt.Sprintf("[\n%v]", strings.Join(actionsDesc, ",\n"))
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

func SessionGetterProvider(t *Task) func() string {
	return func() string {
		if t.tg == nil {
			return fmt.Sprintf("{error: %v}", "thread group is nil")
		}
		pg := t.tg.processGroup
		if pg == nil {
			return fmt.Sprintf("{error: %v}", "process group is nil")
		}
		var pgids []string
		if pg.session != nil {
			sessionPGs := pg.session.processGroups
			if sessionPGs != nil {
				for spg := sessionPGs.Front(); spg != nil; spg = spg.Next() {
					pgids = append(pgids, string(int32(spg.id)))
				}
			}
		}
		return fmt.Sprintf("{sessionId: %v, PGID: %v, foreground: %v, otherPGIDs: [%v]}", pg.session.id, pg.id, pg.session.foreground.id, strings.Join(pgids, ",\n"))
	}
}
