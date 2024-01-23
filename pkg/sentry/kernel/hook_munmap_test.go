package kernel

import "testing"

var munmapHookWithLessArgs = `
	function cb() {
		hooks.munmap(1)
	}
`

func TestMunmapHook_withLessArgs_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, munmapHookWithLessArgs,
		"no error for hook, which requires 2 args, when given less then 2")
}

var munmapHookWithMoreArgs = `
	function cb() {
		hooks.munmap(1, 2, 3)
	}
`

func TestMunmapHook_withMoreArgs_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, munmapHookWithMoreArgs,
		"no error for hook, which requires 2 args, when given more then 2")
}

var munmapHookWithNull1Arg = `
	function cb() {
		hooks.munmap(null, 10)
	}
`

func TestMunmapHook_withNull1Arg_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, munmapHookWithNull1Arg,
		"no error for hook when 1 arg is null")
}

var munmapHookWithUndefined1Arg = `
	function cb() {
		hooks.munmap(undefined, 10)
	}
`

func TestMunmapHook_withUndefined1Arg_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, munmapHookWithUndefined1Arg,
		"no error for hook when 1 arg is undefined")
}

var munmapHookWithNull2Arg = `
	function cb() {
		hooks.munmap(1, null)
	}
`

func TestMunmapHook_withNull2Arg_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, munmapHookWithNull2Arg,
		"no error for hook when 2 arg is null")
}

var munmapHookWithUndefined2Arg = `
	function cb() {
		hooks.munmap(1, undefined)
	}
`

func TestMunmapHook_withUndefined2Arg_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, munmapHookWithUndefined2Arg,
		"no error for hook when 2 arg is undefined")
}
