package kernel

import "testing"

var readBytesHookWithLessArgs = `
	function cb() {
		hooks.readBytes(1)
	}
`

func TestReadBytesHook_withLessArgs_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, readBytesHookWithLessArgs,
		"no error for hook, which requires 2 args, when given less then 2")
}

var readBytesHookWithMoreArgs = `
	function cb() {
		hooks.readBytes(1, 2, 3)
	}
`

func TestReadBytesHook_withMoreArgs_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, readBytesHookWithMoreArgs,
		"no error for hook, which requires 2 args, when given more then 2")
}

var readBytesHookWithNull1Arg = `
	function cb() {
		hooks.readBytes(null, 8)
	}
`

func TestReadBytesHook_withNull1Arg_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, readBytesHookWithNull1Arg,
		"no error for hook when 1 arg is null")
}

var readBytesHookWithUndefined1Arg = `
	function cb() {
		hooks.readBytes(undefined, 8)
	}
`

func TestReadBytesHook_withUndefined1Arg_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, readBytesHookWithUndefined1Arg,
		"no error for hook when 1 arg is undefined")
}

var readBytesHookWithNull2Arg = `
	function cb() {
		hooks.readBytes(1, null)
	}
`

func TestReadBytesHook_withNull2Arg_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, readBytesHookWithNull2Arg,
		"no error for hook when 2 arg is null")
}

var readBytesHookWithUndefined2Arg = `
	function cb() {
		hooks.readBytes(1, undefined)
	}
`

func TestReadBytesHook_withUndefined2Arg_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, writeBytesHookWithUndefined2Arg,
		"no error for hook when 2 arg is undefined")
}
