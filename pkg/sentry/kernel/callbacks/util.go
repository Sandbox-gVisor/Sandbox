package callbacks

import (
	"errors"
	"fmt"
	"github.com/dop251/goja"
	"gvisor.dev/gvisor/pkg/sentry/arch"
	"strconv"
	"sync"
)

// Flag fuck atomic
type Flag struct {
	mutex sync.Mutex
	flag  bool
}

// SetValueAndActionAtomically returns previous
func (flag *Flag) SetValueAndActionAtomically(value bool, fn func()) bool {
	if flag == nil {
		panic("null pointer")
	}

	flag.mutex.Lock()
	defer flag.mutex.Unlock()

	ret := flag.flag
	flag.flag = value

	if fn != nil {
		fn()
	}

	return ret
}

func (flag *Flag) SetValue(value bool) bool {
	return flag.SetValueAndActionAtomically(value, nil)
}

func ArgsCountMismatchError(expected int, provided int) error {
	return errors.New(fmt.Sprintf("Incorrect count of args. Expected %d, but provided %d", expected, provided))
}

func ExtractPtrFromValue(vm *goja.Runtime, value goja.Value) (uintptr, error) {
	var ptr int64
	err := vm.ExportTo(value, &ptr)
	if err != nil {
		return 0, err
	}

	return uintptr(ptr), nil
}

func ExtractInt64FromValue(vm *goja.Runtime, value goja.Value) (int64, error) {
	var ret int64
	err := vm.ExportTo(value, &ret)
	if err != nil {
		return 0, err
	}

	return ret, nil
}

func ExtractByteBufferFromValue(vm *goja.Runtime, value goja.Value) ([]byte, error) {
	var arrBuf goja.ArrayBuffer
	err := vm.ExportTo(value, &arrBuf)
	if err != nil {
		return nil, err
	}

	return arrBuf.Bytes(), nil
}

func ExtractStringFromValue(vm *goja.Runtime, value goja.Value) (string, error) {
	var ret string
	err := vm.ExportTo(value, &ret)
	if err != nil {
		return "", err
	}

	return ret, nil
}

func ExtractArgsFromRetJsValue(
	inputArgs *arch.SyscallArguments, vm *goja.Runtime, value *goja.Value) (retArgs *arch.SyscallArguments, err error) {

	retArgs = &arch.SyscallArguments{}
	*retArgs = *inputArgs
	retObj := (*value).ToObject(vm)

	for _, key := range retObj.Keys() {
		var ind int
		ind, err = strconv.Atoi(key)
		if err != nil {
			return
		}

		if ind < 0 || len(inputArgs) < ind {
			err = errors.New("invalid index of ret args")
			return
		}

		ptrVal := retObj.Get(key)
		var ptr uintptr
		ptr, err = ExtractPtrFromValue(vm, ptrVal)
		if err != nil {
			return
		}
		retArgs[ind].Value = ptr
	}

	return retArgs, nil
}
