package kernel

import "testing"

var getArgvWithMoreArgs = `
	function cb() {
		hooks.getArgv(10)
	}
`

func TestArgvHook_withMoreArgs_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, getArgvWithMoreArgs,
		"no error for hook, which does not require args, when given more then 0")
}
