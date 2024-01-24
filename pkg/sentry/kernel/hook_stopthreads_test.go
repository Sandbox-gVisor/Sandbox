package kernel

import "testing"

var threadStoppingHookWithMoreArgs = `
	function cb() {
		hooks.stopThreads(10)
	}
`

func TestThreadStoppingHook_withMoreArgs_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, threadStoppingHookWithMoreArgs,
		"no error for hook, which does not require args, when given more then 0")
}
