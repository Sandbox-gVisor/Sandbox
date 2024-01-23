package kernel

import "testing"

var writeStringHookWithLessArgs = `
	function cb() {
		hooks.writeString(1)
	}
`

func TestWriteStringHook_withLessArgs_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, writeStringHookWithLessArgs,
		"no error for hook, which requires 2 args, when given less then 2")
}

var writeStringHookWithMoreArgs = `
	function cb() {
		hooks.writeString(1, 2, 3)
	}
`

func TestWriteStringHook_withMoreArgs_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, writeStringHookWithMoreArgs,
		"no error for hook, which requires 2 args, when given more then 2")
}

var writeStringHookWithNull1Arg = `
	function cb() {
		hooks.writeString(null, "hello")
	}
`

func TestWriteStringHook_withNull1Arg_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, writeStringHookWithNull1Arg,
		"no error for hook when 1 arg is null")
}

var writeStringHookWithUndefined1Arg = `
	function cb() {
		hooks.writeString(undefined, "hello")
	}
`

func TestWriteStringHook_withUndefined1Arg_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, writeStringHookWithUndefined1Arg,
		"no error for hook when 1 arg is undefined")
}

var writeStringHookWithNull2Arg = `
	function cb() {
		hooks.writeString(1, null)
	}
`

func TestWriteStringHook_withNull2Arg_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, writeStringHookWithNull2Arg,
		"no error for hook when 2 arg is null")
}

var writeStringHookWithUndefined2Arg = `
	function cb() {
		hooks.writeString(1, undefined)
	}
`

func TestWriteStringHook_withUndefined2Arg_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, writeStringHookWithUndefined2Arg,
		"no error for hook when 2 arg is undefined")
}
