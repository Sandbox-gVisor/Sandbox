package kernel

import "testing"

var fdHookWithLessArgs = `
	function cb() {
		hooks.getFdInfo()
	}
`

func TestFdHook_withLessArgs_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, fdHookWithLessArgs,
		"no error for hook, which require 1 arg, when given 0")
}

var fdHookWithMoreArgs = `
	function cb() {
		hooks.getFdInfo(1, 2)
	}
`

func TestFdHook_withMoreArgs_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, fdHookWithMoreArgs,
		"no error for hook, which require 1 arg, when given 3")
}

var fdHookWithNullArg = `
	function cb() {
		hooks.getFdInfo(null)
	}
`

func TestFdHook_withNullArg_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, fdHookWithNullArg,
		"no error for hook then null arg given")
}

var fdHookWithUndefinedArg = `
	function cb() {
		hooks.getFdInfo(undefined)
	}
`

func TestFdHook_withUndefinedArg_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, fdHookWithUndefinedArg,
		"no error for hook then undefined arg given")
}
