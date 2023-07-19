package callbacks

import (
	"errors"
	"fmt"
	"sync"
)

// fuck atomic
type Flag struct {
	mutex sync.Mutex
	flag  bool
}

// returns previous
func (flag *Flag) SetValueAndActionAtomically(value bool, fn func()) bool {
	if flag == nil {
		panic("null pointer")
	}

	flag.mutex.Lock()
	defer flag.mutex.Unlock()

	ret := flag.flag
	flag.flag = value

	if fn != nil {
		fn()
	}

	return ret
}

func (flag *Flag) SetValue(value bool) bool {
	return flag.SetValueAndActionAtomically(value, nil)
}

func ArgsCountMismatchError(expected int, provided int) error {
	return errors.New(fmt.Sprintf("Incorrect count of args. Expected %d, but provided %d", expected, provided))
}
