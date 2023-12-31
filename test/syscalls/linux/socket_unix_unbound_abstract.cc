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

#include <errno.h>
#include <fcntl.h>
#include <stddef.h>
#include <stdio.h>
#include <sys/un.h>

#include "gmock/gmock.h"
#include "gtest/gtest.h"
#include "test/syscalls/linux/unix_domain_socket_test_util.h"
#include "test/util/cleanup.h"
#include "test/util/file_descriptor.h"
#include "test/util/linux_capability_util.h"
#include "test/util/socket_util.h"
#include "test/util/test_util.h"

namespace gvisor {
namespace testing {

namespace {

// Test fixture for tests that apply to pairs of unbound abstract unix sockets.
using UnboundAbstractUnixSocketPairTest = SocketPairTest;

TEST_P(UnboundAbstractUnixSocketPairTest, AddressAfterNull) {
  auto sockets = ASSERT_NO_ERRNO_AND_VALUE(NewSocketPair());

  struct sockaddr_un addr =
      *reinterpret_cast<const struct sockaddr_un*>(sockets->first_addr());
  ASSERT_EQ(addr.sun_path[sizeof(addr.sun_path) - 1], 0);
  SKIP_IF(addr.sun_path[sizeof(addr.sun_path) - 2] != 0 ||
          addr.sun_path[sizeof(addr.sun_path) - 3] != 0);

  addr.sun_path[sizeof(addr.sun_path) - 2] = 'a';

  ASSERT_THAT(bind(sockets->first_fd(), sockets->first_addr(),
                   sockets->first_addr_size()),
              SyscallSucceeds());

  ASSERT_THAT(bind(sockets->second_fd(),
                   reinterpret_cast<struct sockaddr*>(&addr), sizeof(addr)),
              SyscallSucceeds());
}

TEST_P(UnboundAbstractUnixSocketPairTest, ShortAddressNotExtended) {
  auto sockets = ASSERT_NO_ERRNO_AND_VALUE(NewSocketPair());

  struct sockaddr_un addr =
      *reinterpret_cast<const struct sockaddr_un*>(sockets->first_addr());
  ASSERT_EQ(addr.sun_path[sizeof(addr.sun_path) - 1], 0);

  ASSERT_THAT(bind(sockets->first_fd(), sockets->first_addr(),
                   sockets->first_addr_size() - 1),
              SyscallSucceeds());

  ASSERT_THAT(bind(sockets->second_fd(), sockets->first_addr(),
                   sockets->first_addr_size()),
              SyscallSucceeds());
}

TEST_P(UnboundAbstractUnixSocketPairTest, BindNothing) {
  auto sockets = ASSERT_NO_ERRNO_AND_VALUE(NewSocketPair());
  struct sockaddr_un addr = {.sun_family = AF_UNIX};
  ASSERT_THAT(bind(sockets->first_fd(),
                   reinterpret_cast<struct sockaddr*>(&addr), sizeof(addr)),
              SyscallSucceeds());
}

TEST_P(UnboundAbstractUnixSocketPairTest, AutoBindSuccess) {
  auto sockets = ASSERT_NO_ERRNO_AND_VALUE(NewSocketPair());
  struct sockaddr_un addr = {.sun_family = AF_UNIX};
  ASSERT_THAT(
      bind(sockets->first_fd(), reinterpret_cast<struct sockaddr*>(&addr),
           sizeof(sa_family_t)),
      SyscallSucceeds());
  socklen_t addr_len = sizeof(addr);
  ASSERT_THAT(getsockname(sockets->first_fd(),
                          reinterpret_cast<struct sockaddr*>(&addr), &addr_len),
              SyscallSucceeds());
  // The address consists of a null byte followed by 5 bytes in the character
  // set [0-9a-f].
  EXPECT_EQ(offsetof(struct sockaddr_un, sun_path) + 6, addr_len);
  EXPECT_EQ(addr.sun_path[0], 0);
  for (int i = 1; i < 6; i++) {
    char c = addr.sun_path[i];
    EXPECT_TRUE((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f'));
  }
  if ((GetParam().type & SOCK_DGRAM) == 0) {
    ASSERT_THAT(listen(sockets->first_fd(), 0 /* backlog */),
                SyscallSucceeds());
  }
  ASSERT_THAT(connect(sockets->second_fd(),
                      reinterpret_cast<struct sockaddr*>(&addr), addr_len),
              SyscallSucceeds());
}

TEST_P(UnboundAbstractUnixSocketPairTest, AutoBindAddrInUse) {
  auto sockets = ASSERT_NO_ERRNO_AND_VALUE(NewSocketPair());
  struct sockaddr_un addr = {.sun_family = AF_UNIX};
  ASSERT_THAT(
      bind(sockets->first_fd(), reinterpret_cast<struct sockaddr*>(&addr),
           sizeof(sa_family_t)),
      SyscallSucceeds());
  socklen_t addr_len = sizeof(addr);
  ASSERT_THAT(getsockname(sockets->first_fd(),
                          reinterpret_cast<struct sockaddr*>(&addr), &addr_len),
              SyscallSucceeds());
  ASSERT_THAT(bind(sockets->second_fd(),
                   reinterpret_cast<struct sockaddr*>(&addr), addr_len),
              SyscallFailsWithErrno(EADDRINUSE));
}

TEST_P(UnboundAbstractUnixSocketPairTest, BindConnectInSubNamespace) {
  SKIP_IF(!ASSERT_NO_ERRNO_AND_VALUE(HaveCapability(CAP_NET_ADMIN)));

  const FileDescriptor ns =
      ASSERT_NO_ERRNO_AND_VALUE(Open("/proc/self/ns/net", O_RDONLY));
  auto cleanup =
      Cleanup([&ns] { ASSERT_THAT(setns(ns.get(), 0), SyscallSucceeds()); });
  ASSERT_THAT(unshare(CLONE_NEWNET), SyscallSucceeds());

  auto sockets = ASSERT_NO_ERRNO_AND_VALUE(NewSocketPair());
  ASSERT_THAT(unshare(CLONE_NEWNET), SyscallSucceeds());

  struct sockaddr_un addr = {.sun_family = AF_UNIX};
  ASSERT_THAT(
      bind(sockets->first_fd(), reinterpret_cast<struct sockaddr*>(&addr),
           sizeof(sa_family_t)),
      SyscallSucceeds());
  socklen_t addr_len = sizeof(addr);
  ASSERT_THAT(getsockname(sockets->first_fd(),
                          reinterpret_cast<struct sockaddr*>(&addr), &addr_len),
              SyscallSucceeds());
  if ((GetParam().type & SOCK_DGRAM) == 0) {
    ASSERT_THAT(listen(sockets->first_fd(), 1 /* backlog */),
                SyscallSucceeds());
  }
  EXPECT_THAT(connect(sockets->second_fd(),
                      reinterpret_cast<struct sockaddr*>(&addr), addr_len),
              SyscallSucceeds());

  auto socketsInSubNS = ASSERT_NO_ERRNO_AND_VALUE(NewSocketPair());
  EXPECT_THAT(connect(socketsInSubNS->second_fd(),
                      reinterpret_cast<struct sockaddr*>(&addr), addr_len),
              SyscallFailsWithErrno(ECONNREFUSED));
  EXPECT_THAT(bind(socketsInSubNS->first_fd(),
                   reinterpret_cast<struct sockaddr*>(&addr), addr_len),
              SyscallSucceeds());
}

TEST_P(UnboundAbstractUnixSocketPairTest, ListenZeroBacklog) {
  SKIP_IF((GetParam().type & SOCK_DGRAM) != 0);
  auto sockets = ASSERT_NO_ERRNO_AND_VALUE(NewSocketPair());
  struct sockaddr_un addr = {};
  addr.sun_family = AF_UNIX;
  constexpr char kPath[] = "\x00/foo_bar";
  memcpy(addr.sun_path, kPath, sizeof(kPath));
  ASSERT_THAT(bind(sockets->first_fd(),
                   reinterpret_cast<struct sockaddr*>(&addr), sizeof(addr)),
              SyscallSucceeds());
  ASSERT_THAT(listen(sockets->first_fd(), 0 /* backlog */), SyscallSucceeds());
  ASSERT_THAT(connect(sockets->second_fd(),
                      reinterpret_cast<struct sockaddr*>(&addr), sizeof(addr)),
              SyscallSucceeds());
  auto sockets2 = ASSERT_NO_ERRNO_AND_VALUE(NewSocketPair());
  {
    // Set the FD to O_NONBLOCK.
    int opts;
    ASSERT_THAT(opts = fcntl(sockets2->first_fd(), F_GETFL), SyscallSucceeds());
    opts |= O_NONBLOCK;
    ASSERT_THAT(fcntl(sockets2->first_fd(), F_SETFL, opts), SyscallSucceeds());

    ASSERT_THAT(
        connect(sockets2->first_fd(), reinterpret_cast<struct sockaddr*>(&addr),
                sizeof(addr)),
        SyscallFailsWithErrno(EAGAIN));
  }
  {
    // Set the FD to O_NONBLOCK.
    int opts;
    ASSERT_THAT(opts = fcntl(sockets2->second_fd(), F_GETFL),
                SyscallSucceeds());
    opts |= O_NONBLOCK;
    ASSERT_THAT(fcntl(sockets2->second_fd(), F_SETFL, opts), SyscallSucceeds());

    ASSERT_THAT(
        connect(sockets2->second_fd(),
                reinterpret_cast<struct sockaddr*>(&addr), sizeof(addr)),
        SyscallFailsWithErrno(EAGAIN));
  }
}

TEST_P(UnboundAbstractUnixSocketPairTest, GetSockNameFullLength) {
  auto sockets = ASSERT_NO_ERRNO_AND_VALUE(NewSocketPair());

  ASSERT_THAT(bind(sockets->first_fd(), sockets->first_addr(),
                   sockets->first_addr_size()),
              SyscallSucceeds());

  sockaddr_storage addr = {};
  socklen_t addr_len = sizeof(addr);
  ASSERT_THAT(getsockname(sockets->first_fd(),
                          reinterpret_cast<struct sockaddr*>(&addr), &addr_len),
              SyscallSucceeds());
  EXPECT_EQ(addr_len, sockets->first_addr_size());
}

TEST_P(UnboundAbstractUnixSocketPairTest, GetSockNamePartialLength) {
  auto sockets = ASSERT_NO_ERRNO_AND_VALUE(NewSocketPair());

  ASSERT_THAT(bind(sockets->first_fd(), sockets->first_addr(),
                   sockets->first_addr_size() - 1),
              SyscallSucceeds());

  sockaddr_storage addr = {};
  socklen_t addr_len = sizeof(addr);
  ASSERT_THAT(getsockname(sockets->first_fd(),
                          reinterpret_cast<struct sockaddr*>(&addr), &addr_len),
              SyscallSucceeds());
  EXPECT_EQ(addr_len, sockets->first_addr_size() - 1);
}

INSTANTIATE_TEST_SUITE_P(
    AllUnixDomainSockets, UnboundAbstractUnixSocketPairTest,
    ::testing::ValuesIn(ApplyVec<SocketPairKind>(
        AbstractUnboundUnixDomainSocketPair,
        AllBitwiseCombinations(List<int>{SOCK_STREAM, SOCK_SEQPACKET,
                                         SOCK_DGRAM},
                               List<int>{0, SOCK_NONBLOCK}))));

}  // namespace

}  // namespace testing
}  // namespace gvisor
