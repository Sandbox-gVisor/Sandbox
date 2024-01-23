package kernel

import "testing"

var anonMmapHookWithLessArgs = `
	function cb() {
		hooks.anonMmap()
	}
`

func TestAnonMmapHook_withLessArgs_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, anonMmapHookWithLessArgs,
		"no error for hook, which require 1 arg, when given 0")
}

var anonMmapHookWithMoreArgs = `
	function cb() {
		hooks.anonMmap(1, 2)
	}
`

func TestAnonMmapHook_withMoreArgs_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, anonMmapHookWithMoreArgs,
		"no error for hook, which require 1 arg, when given 3")
}

var anonMmapHookWithNullArg = `
	function cb() {
		hooks.anonMmap(null)
	}
`

func TestAnonMmapHook_withNullArg_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, anonMmapHookWithNullArg,
		"no error for hook then null arg given")
}

var anonMmapHookWithUndefinedArg = `
	function cb() {
		hooks.anonMmap(undefined)
	}
`

func TestAnonMmapHook_withUndefinedArg_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, anonMmapHookWithUndefinedArg,
		"no error for hook then undefined arg given")
}
