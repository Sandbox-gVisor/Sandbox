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

// Package precompiled does not exist. This file is used in a go:embed
// directive inside `precompile_gen.go`.
package precompiled

import (
	"gvisor.dev/gvisor/pkg/seccomp/precompiledseccomp"
	"gvisor.dev/gvisor/pkg/sync"
)

var (
	// precompiledPrograms holds registered programs.
	// It is populated in `registerPrograms`.
	precompiledPrograms map[string]precompiledseccomp.Program = nil

	// registerPrecompiledProgramsOnce ensures that program registration
	// happens only once.
	registerPrecompiledProgramsOnce sync.Once
)

// GetPrecompiled returns the precompiled program for the given name,
// and whether that program name exists.
func GetPrecompiled(programName string) (precompiledseccomp.Program, bool) {
	registerPrecompiledProgramsOnce.Do(registerPrograms)
	program, ok := precompiledPrograms[programName]
	return program, ok
}

// registerPrograms registers available programs inside `precompiledPrograms`.
func registerPrograms() {
	programs := make(map[string]precompiledseccomp.Program)
	// PROGRAM_REGISTRATION_GOES_HERE_THIS_IS_A_LOAD_BEARING_COMMENT
	precompiledPrograms = programs
}
