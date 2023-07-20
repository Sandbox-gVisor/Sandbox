package kernel

import (
	"errors"
	"gvisor.dev/gvisor/pkg/hostarch"
	"gvisor.dev/gvisor/pkg/sentry/vfs"
	"strconv"
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

func FdResolverProvider(t *Task) func() []string {
	return func() []string {
		fdt := t.fdTable
		privileges := make([]string, 10)

		fdt.forEach(t, func(fd int32, fdesc *vfs.FileDescription, _ FDFlags) {

			stat, err := fdesc.Stat(t, vfs.StatOptions{})
			if err != nil {
				return
			}
			name := findPath(t, fd)
			privileges = append(privileges, "("+name+")"+strconv.FormatInt(int64(fd), 10)+":"+parseMask(stat.Mode))
		})

		return privileges
	}
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
