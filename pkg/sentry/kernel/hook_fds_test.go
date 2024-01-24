package kernel

import "testing"

var fdsHookWithMoreArgs = `
	function cb() {
		hooks.getFdsInfo(10)
	}
`

func TestFDsHook_withMoreArgs_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, getEnvsWithMoreArgs,
		"no error for hook, which does not require args, when given more then 0")
}
