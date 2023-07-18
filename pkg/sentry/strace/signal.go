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

package strace

import (
	"fmt"
	"strings"

	"gvisor.dev/gvisor/pkg/abi/linux"
	"gvisor.dev/gvisor/pkg/sentry/kernel"

	"gvisor.dev/gvisor/pkg/hostarch"
)

func sigSet(t *kernel.Task, addr hostarch.Addr) string {
	if addr == 0 {
		return "null"
	}

	var b [linux.SignalSetSize]byte
	if _, err := t.CopyInBytes(addr, b[:]); err != nil {
		return fmt.Sprintf("%#x (error copying sigset: %v)", addr, err)
	}

	set := linux.SignalSet(hostarch.ByteOrder.Uint64(b[:]))

	return fmt.Sprintf("%#x %s", addr, formatSigSet(set))
}

func formatSigSet(set linux.SignalSet) string {
	var signals []string
	linux.ForEachSignal(set, func(sig linux.Signal) {
		signals = append(signals, linux.SignalNames.ParseDecimal(uint64(sig)))
	})

	return fmt.Sprintf("[%v]", strings.Join(signals, " "))
}

func sigAction(t *kernel.Task, addr hostarch.Addr) string {
	if addr == 0 {
		return "null"
	}

	var sa linux.SigAction
	if _, err := sa.CopyIn(t, addr); err != nil {
		return fmt.Sprintf("%#x (error copying sigaction: %v)", addr, err)
	}

	var handler string
	switch sa.Handler {
	case linux.SIG_IGN:
		handler = "SIG_IGN"
	case linux.SIG_DFL:
		handler = "SIG_DFL"
	default:
		handler = fmt.Sprintf("%#x", sa.Handler)
	}

	return fmt.Sprintf("%#x {Handler: %s, Flags: %s, Restorer: %#x, Mask: %s}", addr, handler, linux.SigActionFlags.Parse(sa.Flags), sa.Restorer, formatSigSet(sa.Mask))
}
