package kernel

import (
	json2 "encoding/json"
	"gvisor.dev/gvisor/pkg/abi/linux"
	"gvisor.dev/gvisor/pkg/hostarch"
	"gvisor.dev/gvisor/pkg/sentry/arch"
	"gvisor.dev/gvisor/pkg/sentry/vfs"
	"strconv"
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

func AnonMmap(t *Task, length uintptr) (uintptr, error) {
	var MmapSysno uintptr = 9
	mmapImpl := t.SyscallTable().Lookup(MmapSysno)
	args := arch.SyscallArguments{
		arch.SyscallArgument{Value: 0},
		arch.SyscallArgument{Value: length},
		arch.SyscallArgument{Value: linux.PROT_READ | linux.PROT_WRITE},
		arch.SyscallArgument{Value: linux.MAP_ANONYMOUS},
		arch.SyscallArgument{Value: uintptr(-1)},
		arch.SyscallArgument{Value: 0},
	}
	rval, _, err := mmapImpl(t, MmapSysno, args)
	return rval, err
}
