package strace

import (
	"encoding/json"
	"fmt"
	"strings"
)

type RvalJSONLog struct {
	Retval  []string
	Err     string
	Errno   string
	Elapsed string
}

type straceJSONLog struct {
	LogType     string
	Taskname    string
	Syscallname string
	Output      []string
	Rval        RvalJSONLog
}

func (s *straceJSONLog) ToString() string {
	b, err := json.Marshal(&s)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func (s *straceJSONLog) GVisorString() string {
	if len(s.Rval.Retval) != 0 {
		return fmt.Sprintf("%v %v %v(%v) = %v (%v)", s.Taskname, s.LogType, s.Syscallname, strings.Join(s.Output, " "), strings.Join(s.Rval.Retval, " "), s.Rval.Elapsed)
	} else if len(s.Rval.Err) != 0 {
		return fmt.Sprintf("%v %v %v(%v) = %v errno=%v (%v) (%v)", s.Taskname, s.LogType, s.Syscallname, strings.Join(s.Output, " "), strings.Join(s.Rval.Retval, " "), s.Rval.Errno, s.Rval.Err, s.Rval.Elapsed)
	}
	return fmt.Sprintf("%v %v %v(%v)", s.Taskname, s.LogType, s.Syscallname, strings.Join(s.Output, " "))
}

// add "" to all list elements
func toJSONEnum(args []string) []string {
	var result []string
	for _, s := range args {
		result = append(result, fmt.Sprintf(`"%s"`, s))
	}
	return result
}
