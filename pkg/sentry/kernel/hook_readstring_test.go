package kernel

import "testing"

var readStringHookWithLessArgs = `
	function cb() {
		hooks.readString(1)
	}
`

func TestReadStringHook_withLessArgs_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, readStringHookWithLessArgs,
		"no error for hook, which requires 2 args, when given less then 2")
}

var readStringHookWithMoreArgs = `
	function cb() {
		hooks.readString(1, 2, 3)
	}
`

func TestReadStringHook_withMoreArgs_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, readStringHookWithMoreArgs,
		"no error for hook, which requires 2 args, when given more then 2")
}

var readStringHookWithNull1Arg = `
	function cb() {
		hooks.readString(null, 8)
	}
`

func TestReadStringHook_withNull1Arg_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, readStringHookWithNull1Arg,
		"no error for hook when 1 arg is null")
}

var readStringHookWithUndefined1Arg = `
	function cb() {
		hooks.readString(undefined, 8)
	}
`

func TestReadStringHook_withUndefined1Arg_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, readStringHookWithUndefined1Arg,
		"no error for hook when 1 arg is undefined")
}

var readStringHookWithNull2Arg = `
	function cb() {
		hooks.readString(1, null)
	}
`

func TestReadStringHook_withNull2Arg_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, readStringHookWithNull2Arg,
		"no error for hook when 2 arg is null")
}

var readStringHookWithUndefined2Arg = `
	function cb() {
		hooks.readString(1, undefined)
	}
`

func TestReadStringHook_withUndefined2Arg_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, readStringHookWithUndefined2Arg,
		"no error for hook when 2 arg is undefined")
}
