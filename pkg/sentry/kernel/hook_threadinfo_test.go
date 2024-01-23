package kernel

import "testing"

var threadInfoWith2Args = `
	function cb() {
		hooks.getThreadInfo(1, 2)
	}
`

func TestThreadInfoHook_with3Args_Fails(t *testing.T) {
	testThatCbFailsWithErr(t, threadInfoWith2Args,
		"no error when calling hook, which needs 1 or 0 args, with 2 args")
}

var threadInfoWithNullArg = `
	function cb() {
		hooks.getThreadInfo(null)
	}
`

func TestThreadInfoHook_withNullArg_Fails(t *testing.T) {
	testThatCbFailsWithErr(t, threadInfoWithNullArg,
		"no error when calling hook with null arg")
}

var threadInfoWithUndefinedArg = `
	function cb() {
		hooks.getThreadInfo(undefined)
	}
`

func TestThreadInfoHook_withUndefinedArg_Fails(t *testing.T) {
	testThatCbFailsWithErr(t, threadInfoWithUndefinedArg,
		"no error when calling hook with undefined arg")
}
