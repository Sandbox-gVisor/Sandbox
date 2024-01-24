package kernel

import "testing"

var threadResumingHookWithMoreArgs = `
	function cb() {
		hooks.resumeThreads(10)
	}
`

func TestThreadResumingHook_withMoreArgs_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, threadResumingHookWithMoreArgs,
		"no error for hook, which does not require args, when given more then 0")
}
