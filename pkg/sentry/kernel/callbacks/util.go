package callbacks

import "sync"

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
