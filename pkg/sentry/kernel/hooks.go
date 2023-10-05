package kernel

import (
	"errors"
	"github.com/dop251/goja"
	"sync"
)

// HookCallback is signature of DependentHooks that are called from user`s js callback
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

// GoHook is an interface for DependentHooks, that user can call from js callback
type GoHook interface {
	// description should provide ingo about hook in the HookInfoDto
	description() HookInfoDto

	// jsName - with this name the hook will be called from js
	jsName() string
}

// TaskIndependentGoHook is an interface for DependentHooks, that user can call from js callback when cb run with/without task
type TaskIndependentGoHook interface {
	GoHook
	createCallBack() HookCallback
}

// TaskDependentGoHook is an interface for DependentHooks, that user can call from js callback when cb run with task
type TaskDependentGoHook interface {
	GoHook
	createCallBack(*Task) HookCallback
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

// GoHookDecorator added for future restrictions of DependentHooks
type GoHookDecorator struct {
	wrapped TaskDependentGoHook
}

func (decorator *GoHookDecorator) description() HookInfoDto {
	return decorator.wrapped.description()
}

func (decorator *GoHookDecorator) jsName() string {
	return decorator.wrapped.jsName()
}

func (decorator *GoHookDecorator) createCallBack(t *Task) HookCallback {
	cb := decorator.wrapped.createCallBack(t)
	return disposableDecorator(cb)
}

// HooksTable user`s js callback takes DependentHooks from this table before execution.
// Hooks from the table can be used by user in his js code to get / modify data
type HooksTable struct {
	DependentHooks   map[string]TaskDependentGoHook
	IndependentHooks map[string]TaskIndependentGoHook
	mutex            sync.Mutex
}

func (ht *HooksTable) registerDependentHook(hook TaskDependentGoHook) error {
	if ht == nil {
		return errors.New("DependentHooks table is nil")
	}

	ht.mutex.Lock()
	defer ht.mutex.Unlock()

	ht.DependentHooks[hook.jsName()] = hook //&GoHookDecorator{wrapped: hook}
	return nil
}

func (ht *HooksTable) registerIndependentHook(hook TaskIndependentGoHook) error {
	if ht == nil {
		return errors.New("DependentHooks table is nil")
	}

	ht.mutex.Lock()
	defer ht.mutex.Unlock()

	ht.IndependentHooks[hook.jsName()] = hook //&GoHookDecorator{wrapped: hook}
	return nil
}

func (ht *HooksTable) getDependentHook(hookName string) TaskDependentGoHook {
	if ht == nil {
		panic("Hooks table is nil")
	}

	ht.mutex.Lock()
	defer ht.mutex.Unlock()

	f, ok := ht.DependentHooks[hookName]
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
	defer ht.mutex.Unlock()

	var hooks []GoHook
	for _, hook := range ht.DependentHooks {
		hooks = append(hooks, hook)
	}
	for _, hook := range ht.IndependentHooks {
		hooks = append(hooks, hook)
	}

	return hooks
}

// addDependentHooksToContextObject from this context object user`s callback will take DependentHooks
func (ht *HooksTable) addDependentHooksToContextObject(object *goja.Object, task *Task) error {
	ht.mutex.Lock()
	defer ht.mutex.Unlock()

	for name, hook := range ht.DependentHooks {
		callback := hook.createCallBack(task)
		err := object.Set(name, callback)
		if err != nil {
			return err
		}
	}

	return nil
}

// addIndependentHooksToContextObject from this context object user`s callback will take DependentHooks
func (ht *HooksTable) addIndependentHooksToContextObject(object *goja.Object) error {
	ht.mutex.Lock()
	defer ht.mutex.Unlock()

	for name, hook := range ht.IndependentHooks {
		callback := hook.createCallBack()
		err := object.Set(name, callback)
		if err != nil {
			return err
		}
	}

	return nil
}

// RegisterHooks register all hooks from this file in provided table
func RegisterHooks(cb *HooksTable) error {
	dependentGoHooks := []TaskDependentGoHook{
		&ReadBytesHook{},
		&WriteBytesHook{},
		&ReadStringHook{},
		&WriteStringHook{},
		&EnvvGetterHook{},
		&MmapGetterHook{},
		&ArgvHook{},
		&SignalInfoHook{},
		&PidInfoHook{},
		&FDHook{},
		&FDsHook{},
		&UserJSONLogHook{},
		&AnonMmapHook{},
	}

	independentGoHooks := []TaskIndependentGoHook{
		&PrintHook{},
		&AddCbBeforeHook{},
		&AddCbAfterHook{},
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
