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

func stringifyArg(arg string) string {
	return fmt.Sprintf("%s", arg)
}

func toJsonEnum(args []string) []string {
	var result []string
	for _, s := range args {
		result = append(result, fmt.Sprintf(`"%s"`, s))
	}
	return result
}

func createJsonLog(jsonType string, taskname string, syscallname string, out string, value string) string {
	return fmt.Sprintf(`{"type": "%s", "taskname": "%s", "syscallname": "%s", "output": %s, "rval": %s}`, jsonType, taskname, syscallname, out, value)
}
