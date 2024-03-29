package kernel

import (
	"fmt"
	"gvisor.dev/gvisor/pkg/abi/linux"
	"gvisor.dev/gvisor/pkg/errors/linuxerr"
	"gvisor.dev/gvisor/pkg/hostarch"
	"gvisor.dev/gvisor/pkg/sentry/arch"
	"gvisor.dev/gvisor/pkg/sentry/vfs"
)

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
	SessionID    int32 `json:"sessionID"`
	PGID         int32
	ForegroundID int32   `json:"foregroundID"`
	OtherPGIDs   []int32 `json:"otherPGIDs"`
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

type FDInfo struct {
	Path     string `json:"path"`
	FD       int32  `json:"fd"`
	Mode     string `json:"mode"`
	Nlinks   uint32 `json:"nlinks"`
	Flags    string `json:"flags"`
	Readable bool   `json:"readable"`
	Writable bool   `json:"writable"`
}

// FdsResolver resolves all file descriptors that belong to given task and returns
// path to fd, fd num and fd mask in JSON format
func FdsResolver(t *Task) ([]FDInfo, error) {
	jsonPrivs := make([]FDInfo, 0)

	fdt := t.fdTable

	fdt.forEach(t, func(fd int32, fdesc *vfs.FileDescription, _ FDFlags) {
		stat, err := fdesc.Stat(t, vfs.StatOptions{})
		if err != nil {
			return
		}

		name := findPath(t, fd)
		flags := parseAttributesMask(stat.AttributesMask)

		jsonPrivs = append(jsonPrivs, FDInfo{
			FD:       fd,
			Path:     name,
			Mode:     parseMask(uint16(linux.FileMode(stat.Mode).Permissions())),
			Nlinks:   stat.Nlink,
			Flags:    flags,
			Writable: fdesc.IsWritable(),
			Readable: fdesc.IsReadable(),
		})
	})

	return jsonPrivs, nil
}

// FdResolver resolves one specific fd for given task and returns
// path to fd, fd num and fd mask in JSON format
func FdResolver(t *Task, fd int32) (FDInfo, error) {
	fdesc, _ := t.fdTable.Get(fd)
	if fdesc == nil {
		return FDInfo{}, fmt.Errorf("description for fd %v not found", fd)
	}
	defer fdesc.DecRef(t)
	stat, err := fdesc.Stat(t, vfs.StatOptions{})
	if err != nil {
		return FDInfo{}, fmt.Errorf("metadata for fd %v not found", fd)
	}

	name := findPath(t, fd)

	jsonPrivs := FDInfo{
		Path:     name,
		FD:       fd,
		Mode:     parseMask(uint16(linux.FileMode(stat.Mode).Permissions())),
		Nlinks:   stat.Nlink,
		Flags:    parseAttributesMask(stat.AttributesMask),
		Writable: fdesc.IsWritable(),
		Readable: fdesc.IsReadable(),
	}

	return jsonPrivs, nil
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

func AnonMmap(t *Task, length uintptr) (uintptr, error) {
	var MmapSysno uintptr = 9
	var missingFd uintptr = 0
	mmapImpl := t.SyscallTable().Lookup(MmapSysno)
	args := arch.SyscallArguments{
		arch.SyscallArgument{Value: 0},
		arch.SyscallArgument{Value: length},
		arch.SyscallArgument{Value: linux.PROT_READ | linux.PROT_WRITE},
		arch.SyscallArgument{Value: linux.MAP_ANONYMOUS | linux.MAP_PRIVATE},
		arch.SyscallArgument{Value: missingFd},
		arch.SyscallArgument{Value: 0},
	}
	rval, _, err := mmapImpl(t, MmapSysno, args)
	return rval, err
}

func Unmap(t *Task, addr uintptr, length uintptr) error {
	var MunmapSysno uintptr = 11
	munmapImpl := t.SyscallTable().Lookup(MunmapSysno)
	args := arch.SyscallArguments{
		arch.SyscallArgument{Value: addr},
		arch.SyscallArgument{Value: length},
		arch.SyscallArgument{},
		arch.SyscallArgument{},
		arch.SyscallArgument{},
		arch.SyscallArgument{},
	}
	_, _, err := munmapImpl(t, MunmapSysno, args)
	return err
}

// SendSignalToTaskWithID has similar logic to kill implementation (may be found in pkg/sentry/syscalls/linux/sys_signal.go)
func SendSignalToTaskWithID(t *Task, tid ThreadID, sig linux.Signal) error {
	if !sig.IsValid() {
		return fmt.Errorf("bad signal number")
	}

	target := t.PIDNamespace().TaskWithID(tid)

	if target == nil {
		return linuxerr.ESRCH
	}

	info := &linux.SignalInfo{
		Signo: int32(sig),
		Code:  linux.SI_USER,
	}

	info.SetPID(int32(t.ThreadID()))
	info.SetUID(int32(t.Credentials().RealKUID.In(t.UserNamespace()).OrOverflow()))

	return target.SendSignal(info)
}

func fillThreadInfoDto(t *Task) ThreadInfoDto {
	tids := make([]int32, 0)
	for thread := t.tg.tasks.Front(); thread != nil; thread = thread.Next() {
		tids = append(tids, int32(t.tg.pidns.tids[thread]))
	}

	dto := ThreadInfoDto{
		TID:      int32(t.tg.pidns.tids[t]),
		TGID:     int32(t.tg.pidns.tgids[t.tg]),
		TIDsInTg: tids,
	}

	return dto
}
