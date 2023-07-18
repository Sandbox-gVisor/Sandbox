package strace

import (
	"fmt"
	"encoding/json"
)

type RvalJsonLog struct {
	Retval []string
	Err string
	Errno string
	Elapsed string
}

type straceJsonLog struct {
	LogType string
	Taskname string
	Syscallname string
	Output []string
	Rval RvalJsonLog
}

func (s *straceJsonLog) ToString() string {
	b, err := json.Marshal(&s)
	if err != nil {
		return  "{}"
	}
	return string(b)
}

// add "" to all list elements
func toJsonEnum(args []string) []string {
	var result []string
	for _, s := range args {
		result = append(result, fmt.Sprintf(`"%s"`, s))
	}
	return result
}
