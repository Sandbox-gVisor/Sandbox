package kernel

import "testing"

var pidInfoHookWithMoreArgs = `
	function cb() {
		hooks.getPidInfo(10)
	}
`

func TestPidInfoHook_withMoreArgs_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, pidInfoHookWithMoreArgs,
		"no error for hook, which does not require args, when given more then 0")
}
