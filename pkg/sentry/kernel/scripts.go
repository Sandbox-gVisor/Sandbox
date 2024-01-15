package kernel

import (
	"errors"
	"fmt"
	"github.com/dop251/goja"
	"gvisor.dev/gvisor/pkg/sentry/arch"
	"gvisor.dev/gvisor/pkg/sentry/kernel/callbacks"
	"slices"
	"strconv"
	"strings"
)

const (
	JsCallbackTypeAfter          = "after"
	JsCallbackTypeBefore         = "before"
	HooksJsName                  = "hooks"
	ArgsJsName                   = "args"
	JsSyscallReturnValue         = "ret"
	JsSyscallErrno               = "errno"
	JsPersistenceContextName     = "persistence"
	JsGlobalPersistenceObject    = "glb"
	JsTaskLocalPersistenceObject = "local"
)

type ContextAddable interface {
	addSelfToContextObject(object *goja.Object) error
}

type ObjectAddableAdapter struct {
	object *goja.Object
	name   string
}

func (adapter *ObjectAddableAdapter) addSelfToContextObject(object *goja.Object) error {
	err := object.Set(adapter.name, adapter.object)
	if err != nil {
		return err
	}
	return nil
}

// jsCallbackInvocationTemplate generate string that represent user callback script + invocation of it with injected args
func jsCallbackInvocationTemplate(jsCallback JsCallback) string {
	info := jsCallback.callbackInfo()
	args := make([]string, len(arch.SyscallArguments{}))
	for i := range args {
		args[i] = fmt.Sprintf("%s.arg%d", ArgsJsName, i)
	}

	return fmt.Sprintf("%s; %s(%s)", info.CallbackSource, info.EntryPoint, strings.Join(args, ", "))
}

func extractArgsFromRetJsValue(
	inputArgs *arch.SyscallArguments, vm *goja.Runtime, value goja.Value) (retArgs *arch.SyscallArguments, err error) {

	if value == nil {
		return inputArgs, nil
	}

	retArgs = &arch.SyscallArguments{}
	*retArgs = *inputArgs
	retObj := value.ToObject(vm)

	for _, key := range retObj.Keys() {
		var ind int
		ind, err = strconv.Atoi(key)
		if err != nil {
			continue
		}

		if ind < 0 || len(inputArgs) < ind {
			err = errors.New("invalid index of ret args")
			return
		}

		ptrVal := retObj.Get(key)
		var ptr uintptr
		ptr, err = callbacks.ExtractPtrFromValue(vm, ptrVal)
		if err != nil {
			return
		}
		retArgs[ind].Value = ptr
	}

	return retArgs, nil
}

func extractSubstitutionFromRetJsValue(vm *goja.Runtime, value goja.Value) (*SyscallReturnValue, error) {
	if value == nil {
		return nil, nil
	}
	obj := value.ToObject(vm)

	if slices.Contains(obj.Keys(), JsSyscallReturnValue) && slices.Contains(obj.Keys(), JsSyscallErrno) {
		retVal := obj.Get(JsSyscallReturnValue)
		errnoVal := obj.Get(JsSyscallErrno)

		ret, err := callbacks.ExtractPtrFromValue(vm, retVal)
		if err != nil {
			return nil, err
		}

		errno, err := callbacks.ExtractPtrFromValue(vm, errnoVal)
		if err != nil {
			return nil, err
		}

		return &SyscallReturnValue{returnValue: ret, errno: errno}, nil
	}

	return nil, nil
}

type ScriptContext struct {
	Name  string
	Items []ContextAddable
}

type ScriptContexts []ScriptContext

type ScriptContextsBuilder struct {
	contexts map[string][]ContextAddable
}

func ScriptContextsBuilderOf() *ScriptContextsBuilder {
	return &ScriptContextsBuilder{contexts: map[string][]ContextAddable{}}
}

func (builder *ScriptContextsBuilder) AddContext(context ScriptContext) *ScriptContextsBuilder {
	items, ok := builder.contexts[context.Name]
	if !ok {
		items = make([]ContextAddable, 0)
	}

	items = append(items, context.Items...)
	builder.contexts[context.Name] = items

	return builder
}

func (builder *ScriptContextsBuilder) AddAll(contexts ScriptContexts) *ScriptContextsBuilder {
	for _, it := range contexts {
		builder.AddContext(it)
	}

	return builder
}

func (builder *ScriptContextsBuilder) AddContext2(contextName string, items []ContextAddable) *ScriptContextsBuilder {
	return builder.AddContext(ScriptContext{Name: contextName, Items: items})
}

func (builder *ScriptContextsBuilder) AddContext3(contextName string, item ContextAddable) *ScriptContextsBuilder {
	return builder.AddContext2(contextName, []ContextAddable{item})
}

func (builder *ScriptContextsBuilder) Build() ScriptContexts {
	var result ScriptContexts

	for name, items := range builder.contexts {
		result = append(result, ScriptContext{Name: name, Items: items})
	}

	return result
}

// RunJsScript NB!!!! invoke this method only when you own vm
func RunJsScript(vm *goja.Runtime, jsSource string, contexts []ScriptContext) (goja.Value, error) {

	for _, context := range contexts {
		contextObject := vm.NewObject()

		for _, item := range context.Items {
			err := item.addSelfToContextObject(contextObject)
			if err != nil {
				return nil, err
			}
		}

		err := vm.Set(context.Name, contextObject)
		if err != nil {
			return nil, err
		}
	}

	val, err := vm.RunString(jsSource)
	if err != nil {
		return nil, err
	}

	return val, nil
}

func RunAbstractCallback(t *Task, jsSource string,
	args *arch.SyscallArguments, additionalContexts ScriptContexts) (*arch.SyscallArguments, *SyscallReturnValue, error) {

	runtime := GetJsRuntime()
	runtime.Mutex.Lock()
	defer runtime.Mutex.Unlock()

	builder := ScriptContextsBuilderOf().AddAll(additionalContexts)
	builder = builder.AddContext3(ArgsJsName, &SyscallArgsAddableAdapter{args})
	builder = builder.AddContext3(HooksJsName, &IndependentHookAddableAdapter{ht: runtime.hooksTable})
	builder = builder.AddContext3(HooksJsName, &DependentHookAddableAdapter{ht: runtime.hooksTable, task: t})
	builder = builder.AddContext3(JsPersistenceContextName,
		&ObjectAddableAdapter{name: JsGlobalPersistenceObject, object: runtime.Global})

	if t.taskLocalStorage == nil {
		t.taskLocalStorage = runtime.JsVM.NewObject()
	}
	builder = builder.AddContext3(JsPersistenceContextName,
		&ObjectAddableAdapter{name: JsTaskLocalPersistenceObject, object: t.taskLocalStorage})

	contexts := builder.Build()
	val, err := RunJsScript(runtime.JsVM, jsSource, contexts)
	if err != nil {
		return nil, nil, err
	}
	// TODO
	if val.String() == "undefined" {
		return args, nil, nil
	}

	retArgs, err := extractArgsFromRetJsValue(args, runtime.JsVM, val)
	if err != nil {
		return nil, nil, err
	}

	retSub, err := extractSubstitutionFromRetJsValue(runtime.JsVM, val)
	if err != nil {
		return nil, nil, err
	}

	return retArgs, retSub, nil
}
