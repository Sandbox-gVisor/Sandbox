package kernel

import "testing"

var writeBytesHookWithLessArgs = `
	function cb() {
		hooks.writeBytes(1)
	}
`

func TestWriteBytesHook_withLessArgs_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, writeBytesHookWithLessArgs,
		"no error for hook, which requires 2 args, when given less then 2")
}

var writeBytesHookWithMoreArgs = `
	function cb() {
		hooks.writeBytes(1, 2, 3)
	}
`

func TestWriteBytesHook_withMoreArgs_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, writeBytesHookWithMoreArgs,
		"no error for hook, which requires 2 args, when given more then 2")
}

var writeBytesHookWithNull1Arg = `
	function cb() {
		buf = new ArrayBuffer(8)
		hooks.writeBytes(null, buf)
	}
`

func TestWriteBytesHook_withNull1Arg_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, writeBytesHookWithNull1Arg,
		"no error for hook when 1 arg is null")
}

var writeBytesHookWithUndefined1Arg = `
	function cb() {
		buf = new ArrayBuffer(8)
		hooks.writeBytes(undefined, buf)
	}
`

func TestWriteBytesHook_withUndefined1Arg_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, writeBytesHookWithUndefined1Arg,
		"no error for hook when 1 arg is undefined")
}

var writeBytesHookWithNull2Arg = `
	function cb() {
		hooks.writeBytes(1, null)
	}
`

func TestWriteBytesHook_withNull2Arg_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, writeBytesHookWithNull2Arg,
		"no error for hook when 2 arg is null")
}

var writeBytesHookWithUndefined2Arg = `
	function cb() {
		hooks.writeBytes(1, undefined)
	}
`

func TestWriteBytesHook_withUndefined2Arg_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, writeBytesHookWithUndefined2Arg,
		"no error for hook when 2 arg is undefined")
}
