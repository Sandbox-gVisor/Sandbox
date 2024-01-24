package kernel

import "testing"

var getEnvsWithMoreArgs = `
	function cb() {
		hooks.getEnvs(10)
	}
`

func TestEnvvGetterHook_withMoreArgs_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, getEnvsWithMoreArgs,
		"no error for hook, which does not require args, when given more then 0")
}
