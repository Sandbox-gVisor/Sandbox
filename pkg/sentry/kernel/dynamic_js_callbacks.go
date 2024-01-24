package kernel

import (
	"fmt"
	"github.com/dop251/goja"
	"gvisor.dev/gvisor/pkg/sentry/arch"
	"gvisor.dev/gvisor/pkg/sentry/kernel/callbacks"
	"strings"
	"unicode"
)

func dynamicJsCallbackEntryPoint() string {
	args := make([]string, len(arch.SyscallArguments{}))
	for i := range args {
		args[i] = fmt.Sprintf("%s.arg%d", ArgsJsName, i)
	}

	return fmt.Sprintf("__callback__.invoke(%s)", strings.Join(args, ", "))
}

// DynamicJsCallbackBefore implements CallbackBefore
type DynamicJsCallbackBefore struct {
	CallbackInfo callbacks.JsCallbackInfo
	Holder       *goja.Object
}

func (d *DynamicJsCallbackBefore) CallbackBeforeFunc(t *Task, _ uintptr,
	args *arch.SyscallArguments) (*arch.SyscallArguments, *SyscallReturnValue, error) {

	context := ScriptContextsBuilderOf().AddContext3("__callback__",
		&ObjectAddableAdapter{name: "invoke", object: d.Holder}).Build()

	return RunAbstractCallback(t, dynamicJsCallbackEntryPoint(), args, context)
}

func (d *DynamicJsCallbackBefore) Info() callbacks.JsCallbackInfo {
	return d.CallbackInfo
}

// DynamicJsCallbackAfter implements CallbackAfter
type DynamicJsCallbackAfter struct {
	CallbackInfo callbacks.JsCallbackInfo
	Holder       *goja.Object
}

func (d *DynamicJsCallbackAfter) CallbackAfterFunc(t *Task, _ uintptr,
	args *arch.SyscallArguments, ret uintptr, inputErr error) (*arch.SyscallArguments, *SyscallReturnValue, error) {

	builder := ScriptContextsBuilderOf()
	builder = builder.AddContext3(ArgsJsName, SyscallReturnValueWithError{returnValue: ret, errno: inputErr})
	builder = builder.AddContext3("__callback__", &ObjectAddableAdapter{name: "invoke", object: d.Holder})

	return RunAbstractCallback(t, dynamicJsCallbackEntryPoint(), args, builder.Build())
}

func (d *DynamicJsCallbackAfter) Info() callbacks.JsCallbackInfo {
	return d.CallbackInfo
}

func fillJsCallbackInfoForDynamicCallback(info callbacks.JsCallbackInfo, body string) callbacks.JsCallbackInfo {
	info.CallbackBody = body
	info.CallbackSource = body

	_, after, ok := strings.Cut(body, "function")
	if !ok {
		return *unknownCallback(uintptr(info.Sysno), info.Type)
	}
	trimmed := strings.TrimSpace(after)
	splited := strings.SplitN(trimmed, "(", 2)
	if len(splited) != 2 {
		return *unknownCallback(uintptr(info.Sysno), info.Type)
	}
	info.EntryPoint = strings.TrimSpace(splited[0])

	splited = strings.SplitN(splited[1], ")", 2)
	if len(splited) != 2 {
		return info
	}

	gluedArgsBuilder := strings.Builder{}
	for _, r := range splited[0] {
		if !unicode.IsSpace(r) && !unicode.IsControl(r) {
			_, _ = gluedArgsBuilder.WriteRune(r)
		}
	}
	gluedArgs := gluedArgsBuilder.String()
	if gluedArgs == "" {
		return info
	}
	args := strings.Split(gluedArgs, ",")
	info.CallbackArgs = make([]string, 0)
	for i := 0; i < len(args); i++ {
		cleanedArg := strings.Trim(args[i], "\t ,\n")
		info.CallbackArgs = append(info.CallbackArgs, cleanedArg)
	}
	return info
}
