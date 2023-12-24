// Copyright 2023 The gVisor Authors.
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

// Package nvproxy provides subcommands for the nvproxy command.
package nvproxy

import (
	"context"
	"fmt"

	"github.com/google/subcommands"
	"gvisor.dev/gvisor/pkg/sentry/devices/nvproxy"
	"gvisor.dev/gvisor/runsc/flag"
)

// listSupportedDrivers implements subcommands.Command for the "nvproxy list-supported-drivers" command.
type listSupportedDrivers struct{}

// Name implements subcommands.Command.
func (*listSupportedDrivers) Name() string {
	return "list-supported-drivers"
}

// Synopsis implements subcommands.Command.
func (*listSupportedDrivers) Synopsis() string {
	return "list all nvidia driver versions supported by nvproxy"
}

// Usage implements subcommands.Command.
func (*listSupportedDrivers) Usage() string {
	return `list-supported-drivers - list all nvidia driver versions supported by nvproxy
`
}

// SetFlags implements subcommands.Command.
func (*listSupportedDrivers) SetFlags(*flag.FlagSet) {}

// Execute implements subcommands.Command.
func (*listSupportedDrivers) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus {
	if f.NArg() != 0 {
		f.Usage()
		return subcommands.ExitUsageError
	}

	for version, _ := range nvproxy.GetSupportedDriversAndChecksums() {
		fmt.Println(version)
	}

	return subcommands.ExitSuccess
}
