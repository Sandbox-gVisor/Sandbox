package kernel

import "testing"

var signalSendingHookWithLessArgs = `
	function cb() {
		hooks.sendSignal(1)
	}
`

func TestSignalSendingHook_withLessArgs_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, signalSendingHookWithLessArgs,
		"no error for hook, which requires 2 args, when given less then 2")
}

var signalSendingHookWithMoreArgs = `
	function cb() {
		hooks.sendSignal(1, 2, 3)
	}
`

func TestSignalSendingHook_withMoreArgs_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, signalSendingHookWithMoreArgs,
		"no error for hook, which requires 2 args, when given more then 2")
}

var signalSendingHookWithNull1Arg = `
	function cb() {
		hooks.sendSignal(null, 2)
	}
`

func TestSignalSendingHook_withNull1Arg_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, signalSendingHookWithNull1Arg,
		"no error for hook when 1 arg is null")
}

var signalSendingHookWithUndefined1Arg = `
	function cb() {
		hooks.sendSignal(undefined, 2)
	}
`

func TestSignalSendHook_withUndefined1Arg_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, signalSendingHookWithUndefined1Arg,
		"no error for hook when 1 arg is undefined")
}

var signalSendingHookWithNull2Arg = `
	function cb() {
		hooks.sendSignal(1, null)
	}
`

func TestSignalSendingHook_withNull2Arg_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, signalSendingHookWithNull2Arg,
		"no error for hook when 2 arg is null")
}

var signalSendingHookWithUndefined2Arg = `
	function cb() {
		hooks.sendSignal(1, undefined)
	}
`

func TestSignalSendingHook_withUndefined2Arg_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, signalSendingHookWithUndefined2Arg,
		"no error for hook when 2 arg is undefined")
}
