// Copyright 2021 The gVisor Authors.
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

//go:build !false
// +build !false

package kernel

// pidNamespaceData may contain extra per-PID-namespace data.
// +stateify savable
type pidNamespaceData struct {
}

// newPIDNamespaceData returns a new `pidNamespaceData` struct.
func newPIDNamespaceData() pidNamespaceData {
	return pidNamespaceData{}
}

// addTask adds a Task into this PIDNamespace.
// It is always performed under TaskSet lock.
func (ns *PIDNamespace) addTask(t *Task, tid ThreadID) error {
	ns.tasks[tid] = t
	ns.tids[t] = tid
	if t.tg.leader == nil {
		// New thread group.
		ns.tgids[t.tg] = tid
	}
	return nil
}

// deleteTask deletes a Task from this PIDNamespace.
// It is always performed under TaskSet lock.
func (ns *PIDNamespace) deleteTask(t *Task) {
	delete(ns.tasks, ns.tids[t])
	delete(ns.tids, t)
	if t == t.tg.leader || t.tg.leader == nil {
		delete(ns.tgids, t.tg)
	}
}
