package kernel

import "testing"

var getMmapsWithMoreArgs = `
	function cb() {
		hooks.getMmaps(10)
	}
`

func TestMmapGetterHook_withMoreArgs_Fails(t *testing.T) {
	testThatCbFailsWithErr(
		t, getMmapsWithMoreArgs,
		"no error for hook, which does not require args, when given more then 0")
}
