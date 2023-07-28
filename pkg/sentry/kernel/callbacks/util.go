package callbacks

import (
	"errors"
	"fmt"
	"github.com/dop251/goja"
	"github.com/dop251/goja/parser"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type Flag struct {
	mutex sync.Mutex
	flag  bool
}

// SetValueAndActionAtomically returns previous value of flag
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

// methods below extract go types from goja types

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

func ExtractStatementsFromScript(scriptSrs string) ([]string, error) {
	program, err := parser.ParseFile(nil, "", scriptSrs, 0)
	if err != nil {
		return nil, err
	}
	var statements []string

	for _, it := range program.Body {
		statement := scriptSrs[it.Idx0()-1 : it.Idx1()-1]
		statements = append(statements, statement)
	}

	return statements, nil
}

func (info *JsCallbackInfo) ToString() string {
	return fmt.Sprintf("[sysno: %d, entry-point: %s, body: %s, args: %v, type: %s]",
		info.Sysno, info.EntryPoint, info.CallbackBody, info.CallbackArgs, info.Type)
}

const CallbackRegex = `function\s+(sys_(\w+)_(\d+))\s*\(([^)]*)\)\s*{([^}]*)}`

func ExtractCallbacksFromScript(script string) ([]JsCallbackInfo, error) {
	statements, err := ExtractStatementsFromScript(script)
	if err != nil {
		return nil, err
	}
	var infos []JsCallbackInfo
	reCallback := regexp.MustCompile(CallbackRegex)

	convertToIntOrPanic := func(s string) int {
		num, err := strconv.Atoi(s)
		if err != nil {
			panic(err)
		}
		return num
	}

	extractArgs := func(s string) []string {
		reg := regexp.MustCompile(`\s+`)
		args := strings.Split(reg.ReplaceAllString(s, ""), ",")
		return args
	}

	for _, s := range statements {
		match := reCallback.FindStringSubmatch(s)
		if match != nil {
			info := JsCallbackInfo{
				EntryPoint:     match[1],
				Type:           match[2],
				Sysno:          convertToIntOrPanic(match[3]),
				CallbackArgs:   extractArgs(match[4]),
				CallbackBody:   s,
				CallbackSource: script,
			}
			infos = append(infos, info)
		}
	}

	return infos, nil
}
