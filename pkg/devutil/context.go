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

package devutil

import "gvisor.dev/gvisor/pkg/context"

// contextID is this package's type for context.Context.Value keys.
type contextID int

const (
	// CtxDevGoferClient is a Context.Value key for a /dev gofer client.
	CtxDevGoferClient contextID = iota
)

// GoferClientFromContext returns the device gofer client used by ctx.
func GoferClientFromContext(ctx context.Context) *GoferClient {
	if v := ctx.Value(CtxDevGoferClient); v != nil {
		return v.(*GoferClient)
	}
	return nil
}
