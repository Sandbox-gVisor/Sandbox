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

package inet

import (
	"fmt"
	"math/rand"

	"gvisor.dev/gvisor/pkg/context"
	"gvisor.dev/gvisor/pkg/refs"
	"gvisor.dev/gvisor/pkg/sentry/socket/unix/transport"
	"gvisor.dev/gvisor/pkg/syserr"
)

// +stateify savable
type abstractEndpoint struct {
	ep     transport.BoundEndpoint
	socket refs.TryRefCounter
	name   string
	ns     *AbstractSocketNamespace
}

// AbstractSocketNamespace is used to implement the Linux abstract socket functionality.
//
// +stateify savable
type AbstractSocketNamespace struct {
	mu abstractSocketNamespaceMutex `state:"nosave"`

	// Keeps a mapping from name to endpoint. AbstractSocketNamespace does not hold
	// any references on any sockets that it contains; when retrieving a socket,
	// TryIncRef() must be called in case the socket is concurrently being
	// destroyed. It is the responsibility of the socket to remove itself from the
	// abstract socket namespace when it is destroyed.
	endpoints map[string]abstractEndpoint
}

// A boundEndpoint wraps a transport.BoundEndpoint to maintain a reference on
// its backing socket.
type boundEndpoint struct {
	transport.BoundEndpoint
	socket refs.TryRefCounter
}

// Release implements transport.BoundEndpoint.Release.
func (e *boundEndpoint) Release(ctx context.Context) {
	e.socket.DecRef(ctx)
	e.BoundEndpoint.Release(ctx)
}

func (a *AbstractSocketNamespace) init() {
	a.endpoints = make(map[string]abstractEndpoint)
}

// BoundEndpoint retrieves the endpoint bound to the given name. The return
// value is nil if no endpoint was bound.
func (a *AbstractSocketNamespace) BoundEndpoint(name string) transport.BoundEndpoint {
	a.mu.Lock()
	defer a.mu.Unlock()

	ep, ok := a.endpoints[name]
	if !ok {
		return nil
	}

	if !ep.socket.TryIncRef() {
		// The socket has reached zero references and is being destroyed.
		return nil
	}

	return &boundEndpoint{ep.ep, ep.socket}
}

// Bind binds the given socket.
//
// When the last reference managed by socket is dropped, ep may be removed from the
// namespace.
func (a *AbstractSocketNamespace) Bind(ctx context.Context, path string, ep transport.BoundEndpoint, socket refs.TryRefCounter) (string, *syserr.Error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	name := ""
	if path == "" {
		// Autobind feature.
		mask := uint32(0xFFFFF)
		r := rand.Uint32()
		for i := uint32(0); i <= mask; i++ {
			p := fmt.Sprintf("X%05x", (r+i)&mask)
			if _, ok := a.endpoints[p[1:]]; ok {
				continue
			}
			b := ([]byte)(p)
			b[0] = 0
			path = string(b)
			break
		}
		if path == "" {
			return "", syserr.ErrNoSpace
		}
		name = path[1:]
	} else {
		name = path[1:]
		// Check if there is already a socket (which has not yet been destroyed) bound at name.
		if _, ok := a.endpoints[name]; ok {
			return "", syserr.ErrPortInUse
		}
	}

	ae := abstractEndpoint{ep: ep, name: name, ns: a}
	ae.socket = socket
	a.endpoints[name] = ae
	return path, nil
}

// Remove removes the specified socket at name from the abstract socket
// namespace, if it has not yet been replaced.
func (a *AbstractSocketNamespace) Remove(name string, socket refs.TryRefCounter) {
	a.mu.Lock()
	defer a.mu.Unlock()

	ep, ok := a.endpoints[name]
	if !ok {
		// We never delete a map entry apart from a socket's destructor (although the
		// map entry may be overwritten). Therefore, a socket should exist, even if it
		// may not be the one we expect.
		panic(fmt.Sprintf("expected socket to exist at '%s' in abstract socket namespace", name))
	}

	// A Bind() operation may race with callers of Remove(), e.g. in the
	// following case:
	//   socket1 reaches zero references and begins destruction
	//   a.Bind("foo", ep, socket2) replaces socket1 with socket2
	//   socket1's destructor calls a.Remove("foo", socket1)
	//
	// Therefore, we need to check that the socket at name is what we expect
	// before modifying the map.
	if ep.socket == socket {
		delete(a.endpoints, name)
	}
}
