// Copyright 2018 The gVisor Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"
)

type jsonLog struct {
	Msg   string    `json:"msg"`
	Level Level     `json:"level"`
	Time  time.Time `json:"time"`
}

// MarshalJSON implements json.Marshaler.MarashalJSON.
func (l Level) MarshalJSON() ([]byte, error) {
	switch l {
	case Warning:
		return []byte(`"warning"`), nil
	case Info:
		return []byte(`"info"`), nil
	case Debug:
		return []byte(`"debug"`), nil
	default:
		return nil, fmt.Errorf("unknown level %v", l)
	}
}

// UnmarshalJSON implements json.Unmarshaler.UnmarshalJSON.  It can unmarshal
// from both string names and integers.
func (l *Level) UnmarshalJSON(b []byte) error {
	switch s := string(b); s {
	case "0", `"warning"`:
		*l = Warning
	case "1", `"info"`:
		*l = Info
	case "2", `"debug"`:
		*l = Debug
	default:
		return fmt.Errorf("unknown level %q", s)
	}
	return nil
}

// JSONEmitter logs messages in json format.
type JSONEmitter struct {
	*Writer
}

// Emit implements Emitter.Emit.
func (e JSONEmitter) Emit(_ int, level Level, timestamp time.Time, format string, v ...any) {
	j := jsonLog{
		Msg:   fmt.Sprintf(format, v...),
		Level: level,
		Time:  timestamp,
	}
	b, err := json.Marshal(j)
	if err != nil {
		panic(err)
	}
	e.Writer.Write(b)
}

type moreJSONLog struct {
	Msg   interface{} `json:"msg"`
	Level Level       `json:"level"`
	Time  time.Time   `json:"time"`
}

// MoreJSONEmitter logs in json format. Message should also be int json format
type MoreJSONEmitter struct {
	*Writer
}

// Emit implements Emitter.Emit.
//
// note that in this Emitter argument v is not supported
func (e MoreJSONEmitter) Emit(_ int, level Level, timestamp time.Time, format string, v ...any) {
	var jsObj interface{}
	err := json.Unmarshal([]byte(format), &jsObj)
	if err != nil {
		jsObj = fmt.Sprintf(format, v...)
	}
	j := moreJSONLog{
		Msg:   jsObj,
		Level: level,
		Time:  timestamp,
	}
	b, err := json.Marshal(j)
	if err != nil {
		panic(err)
	}
	e.Writer.Write(b)
}

// SetJSONTarget sets the log target.
//
// This is not thread safe and shouldn't be called concurrently with any
// logging calls.
//
// SetTarget should be called before any instances of log.Log() to avoid race conditions
func SetJSONTarget(target Emitter) {
	logMu.Lock()
	defer logMu.Unlock()
	oldLog := Log()
	jsonLogVal.Store(&JSONLogger{Level: oldLog.Level, Emitter: target})
}

type JSONLogger struct {
	Level
	Emitter
}

// Debugf implements logger.Debugf.
func (l *JSONLogger) Debugf(format string, v ...any) {
	if l == nil {
		return
	}
	l.DebugfAtDepth(1, format, v...)
}

// Infof implements logger.Infof.
func (l *JSONLogger) Infof(format string, v ...any) {
	if l == nil {
		return
	}
	l.InfofAtDepth(1, format, v...)
}

// Warningf implements logger.Warningf.
func (l *JSONLogger) Warningf(format string, v ...any) {
	if l == nil {
		return
	}
	l.WarningfAtDepth(1, format, v...)
}

// DebugfAtDepth logs at a specific depth.
func (l *JSONLogger) DebugfAtDepth(depth int, format string, v ...any) {
	if l == nil {
		return
	}
	if l.IsLogging(Debug) {
		l.Emit(1+depth, Debug, time.Now(), format, v...)
	}
}

// InfofAtDepth logs at a specific depth.
func (l *JSONLogger) InfofAtDepth(depth int, format string, v ...any) {
	if l == nil {
		return
	}
	if l.IsLogging(Info) {
		l.Emit(1+depth, Info, time.Now(), format, v...)
	}
}

// WarningfAtDepth logs at a specific depth.
func (l *JSONLogger) WarningfAtDepth(depth int, format string, v ...any) {
	if l == nil {
		return
	}
	if l.IsLogging(Warning) {
		l.Emit(1+depth, Warning, time.Now(), format, v...)
	}
}

// IsLogging implements logger.IsLogging.
func (l *JSONLogger) IsLogging(level Level) bool {
	return atomic.LoadUint32((*uint32)(&l.Level)) >= uint32(level)
}

// SetLevel sets the logging level.
func (l *JSONLogger) SetLevel(level Level) {
	atomic.StoreUint32((*uint32)(&l.Level), uint32(level))
}
