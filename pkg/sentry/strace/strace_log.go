package strace

import (
	"encoding/json"
	"fmt"
	"strings"
)

type RvalJsonLog struct {
	Retval  []string
	Err     string
	Errno   string
	Elapsed string
}

type straceJsonLog struct {
	LogType     string
	Taskname    string
	Syscallname string
	Output      []string
	Rval        RvalJsonLog
}

func (s *straceJsonLog) ToString() string {
	b, err := json.Marshal(&s)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func (s *straceJsonLog) GVisorString() string {
	return fmt.Sprintf("%v %v %v(%v) = %v errno=%v (%v) (%v)", s.Taskname, s.LogType, s.Syscallname, strings.Join(s.Output, " "), strings.Join(s.Rval.Retval, " "), s.Rval.Errno, s.Rval.Err, s.Rval.Elapsed)
}

// add "" to all list elements
func toJsonEnum(args []string) []string {
	var result []string
	for _, s := range args {
		result = append(result, fmt.Sprintf(`"%s"`, s))
	}
	return result
}
