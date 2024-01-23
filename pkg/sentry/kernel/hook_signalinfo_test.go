package kernel

import "testing"

var signalInfoHookWithMoreArgs = `
	function cb() {
		hooks.getSignalInfo(10)
	}
`

func TestSignalInfoHook_withMoreArgs_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, signalInfoHookWithMoreArgs,
		"no error for hook, which does not require args, when given more then 0")
}
