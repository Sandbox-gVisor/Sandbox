package kernel

import (
	"cmp"
	"errors"
	"github.com/dop251/goja"
	"slices"
	"strings"
	"sync"
)

// HookCallback is signature of dependentHooks that are called from user`s js callback
type HookCallback func(...goja.Value) (interface{}, error)

// HookInfoDto is used to describe the hook: it's name, arguments, return value and description (what hook do)
type HookInfoDto struct {
	// Name contains the jsName
	Name string `json:"name"`

	// Description of the hook
	Description string `json:"description"`

	// Args has such format:
	// argName type description
	Args string `json:"args"`

	// ReturnValue - description of the return value
	ReturnValue string `json:"return-value"`
}

// GoHook is an interface for dependentHooks, that user can call from js callback
type GoHook interface {
	// description should provide ingo about hook in the HookInfoDto
	description() HookInfoDto

	// jsName - with this name the hook will be called from js
	jsName() string
}

// TaskIndependentGoHook is an interface for hooks, that user can call from js callback when cb run with/without task
type TaskIndependentGoHook interface {
	GoHook
	createCallback() HookCallback
}

// TaskDependentGoHook is an interface for hooks, that user can call from js callback when cb run with task
type TaskDependentGoHook interface {
	GoHook
	createCallback(*Task) HookCallback
}

// disposableDecorator is used to prevent deadlocks when same callback is called twice
func disposableDecorator(callback HookCallback) HookCallback {
	callbackWasInvoked := false
	return func(args ...goja.Value) (interface{}, error) {
		if callbackWasInvoked {
			panic("this callback should use only one time")
		}

		callbackWasInvoked = true
		return callback(args...)
	}
}

// GoHookDecorator added for future restrictions of dependentHooks
type GoHookDecorator struct {
	wrapped TaskDependentGoHook
}

func (decorator *GoHookDecorator) description() HookInfoDto {
	return decorator.wrapped.description()
}

func (decorator *GoHookDecorator) jsName() string {
	return decorator.wrapped.jsName()
}

func (decorator *GoHookDecorator) createCallback(t *Task) HookCallback {
	cb := decorator.wrapped.createCallback(t)
	return disposableDecorator(cb)
}

// HooksTable user`s js callback takes Dependent (and/or Independent) Hooks from this table before execution.
// Hooks from the table can be used by user in his js code to get / modify data
type HooksTable struct {
	dependentHooks   map[string]TaskDependentGoHook
	independentHooks map[string]TaskIndependentGoHook
	mutex            sync.Mutex
}

func (ht *HooksTable) registerDependentHook(hook TaskDependentGoHook) error {
	if ht == nil {
		return errors.New("dependentHooks table is nil")
	}

	ht.mutex.Lock()
	defer ht.mutex.Unlock()

	ht.dependentHooks[hook.jsName()] = hook //&GoHookDecorator{wrapped: hook}
	return nil
}

func (ht *HooksTable) registerIndependentHook(hook TaskIndependentGoHook) error {
	if ht == nil {
		return errors.New("dependentHooks table is nil")
	}

	ht.mutex.Lock()
	defer ht.mutex.Unlock()

	ht.independentHooks[hook.jsName()] = hook //&GoHookDecorator{wrapped: hook}
	return nil
}

func (ht *HooksTable) getDependentHook(hookName string) TaskDependentGoHook {
	if ht == nil {
		panic("Hooks table is nil")
	}

	ht.mutex.Lock()
	defer ht.mutex.Unlock()

	f, ok := ht.dependentHooks[hookName]
	if ok {
		return f
	} else {
		return nil
	}
}

func (ht *HooksTable) getCurrentHooks() []GoHook {
	if ht == nil {
		panic("table is nil")
	}

	ht.mutex.Lock()

	var hooks []GoHook
	for _, hook := range ht.dependentHooks {
		hooks = append(hooks, hook)
	}
	for _, hook := range ht.independentHooks {
		hooks = append(hooks, hook)
	}

	ht.mutex.Unlock()

	slices.SortFunc[[]GoHook, GoHook](hooks, func(a GoHook, b GoHook) int {
		return cmp.Compare[string](
			strings.ToLower(a.jsName()),
			strings.ToLower(b.jsName()))
	})

	return hooks
}

// addDependentHooksToContextObject from this context object user`s callback will take dependentHooks
func (ht *HooksTable) addDependentHooksToContextObject(object *goja.Object, task *Task) error {
	ht.mutex.Lock()
	defer ht.mutex.Unlock()

	for name, hook := range ht.dependentHooks {
		callback := hook.createCallback(task)
		err := object.Set(name, callback)
		if err != nil {
			return err
		}
	}

	return nil
}

// addIndependentHooksToContextObject from this context object user`s callback will take independentHooks
func (ht *HooksTable) addIndependentHooksToContextObject(object *goja.Object) error {
	ht.mutex.Lock()
	defer ht.mutex.Unlock()

	for name, hook := range ht.independentHooks {
		callback := hook.createCallback()
		err := object.Set(name, callback)
		if err != nil {
			return err
		}
	}

	return nil
}

// RegisterHooks register all hooks from ./hooks_impl.go in provided table.
// New hooks should be registered here and in ./hooks_test.go in prepareSetOf...Hooks.
func RegisterHooks(cb *HooksTable) error {
	dependentGoHooks := []TaskDependentGoHook{
		&ReadBytesHook{},
		&ReadStringHook{},
		&EnvvGetterHook{},
		&MmapGetterHook{},
		&ArgvHook{},
		&SignalInfoHook{},
		&PidInfoHook{},
		&FDHook{},
		&FDsHook{},
		&UserJSONLogHook{},
		&AnonMmapHook{},
		&MunmapHook{},
		&SignalSendingHook{},
		&ThreadsStoppingHook{},
		&ThreadsResumingHook{},
		&ThreadInfoHook{},
		&WriteBytesHook{},
		&WriteStringHook{},
	}

	independentGoHooks := []TaskIndependentGoHook{
		&AddCbAfterHook{},
		&AddCbBeforeHook{},
		&PrintHook{},
		&SignalMaskToSignalNamesHook{},
		&SignalByNameHook{},
	}

	for _, hook := range dependentGoHooks {
		err := cb.registerDependentHook(hook)
		if err != nil {
			return err
		}
	}

	for _, hook := range independentGoHooks {
		err := cb.registerIndependentHook(hook)
		if err != nil {
			return err
		}
	}

	return nil
}

// DependentHookAddableAdapter implement ContextAddable
type DependentHookAddableAdapter struct {
	ht   *HooksTable
	task *Task
}

func (d *DependentHookAddableAdapter) addSelfToContextObject(object *goja.Object) error {
	return d.ht.addDependentHooksToContextObject(object, d.task)
}

// IndependentHookAddableAdapter implement ContextAddable
type IndependentHookAddableAdapter struct {
	ht *HooksTable
}

func (d *IndependentHookAddableAdapter) addSelfToContextObject(object *goja.Object) error {
	return d.ht.addIndependentHooksToContextObject(object)
}
