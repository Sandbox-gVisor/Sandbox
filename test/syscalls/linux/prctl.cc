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

#include <sys/prctl.h>
#include <sys/ptrace.h>
#include <sys/types.h>
#include <sys/wait.h>
#include <unistd.h>

#include <string>

#include "gtest/gtest.h"
#include "absl/flags/flag.h"
#include "test/util/capability_util.h"
#include "test/util/cleanup.h"
#include "test/util/multiprocess_util.h"
#include "test/util/posix_error.h"
#include "test/util/signal_util.h"
#include "test/util/test_util.h"
#include "test/util/thread_util.h"

ABSL_FLAG(bool, prctl_no_new_privs_test_child, false,
          "If true, exit with the return value of prctl(PR_GET_NO_NEW_PRIVS) "
          "plus an offset (see test source).");

namespace gvisor {
namespace testing {

namespace {

#ifndef SUID_DUMP_DISABLE
#define SUID_DUMP_DISABLE 0
#endif /* SUID_DUMP_DISABLE */
#ifndef SUID_DUMP_USER
#define SUID_DUMP_USER 1
#endif /* SUID_DUMP_USER */
#ifndef SUID_DUMP_ROOT
#define SUID_DUMP_ROOT 2
#endif /* SUID_DUMP_ROOT */

TEST(PrctlTest, NameInitialized) {
  const size_t name_length = 20;
  char name[name_length] = {};
  ASSERT_THAT(prctl(PR_GET_NAME, name), SyscallSucceeds());
  ASSERT_NE(std::string(name), "");
}

TEST(PrctlTest, SetNameLongName) {
  const size_t name_length = 20;
  const std::string long_name(name_length, 'A');
  ASSERT_THAT(prctl(PR_SET_NAME, long_name.c_str()), SyscallSucceeds());
  char truncated_name[name_length] = {};
  ASSERT_THAT(prctl(PR_GET_NAME, truncated_name), SyscallSucceeds());
  const size_t truncated_length = 15;
  ASSERT_EQ(long_name.substr(0, truncated_length), std::string(truncated_name));
}

TEST(PrctlTest, ChildProcessName) {
  constexpr size_t kMaxNameLength = 15;

  char parent_name[kMaxNameLength + 1] = {};
  memset(parent_name, 'a', kMaxNameLength);

  ASSERT_THAT(prctl(PR_SET_NAME, parent_name), SyscallSucceeds());

  pid_t child_pid = fork();
  TEST_PCHECK(child_pid >= 0);
  if (child_pid == 0) {
    char child_name[kMaxNameLength + 1] = {};
    TEST_PCHECK(prctl(PR_GET_NAME, child_name) >= 0);
    TEST_CHECK(memcmp(parent_name, child_name, sizeof(parent_name)) == 0);
    _exit(0);
  }

  int status;
  ASSERT_THAT(waitpid(child_pid, &status, 0),
              SyscallSucceedsWithValue(child_pid));
  EXPECT_TRUE(WIFEXITED(status) && WEXITSTATUS(status) == 0)
      << "status =" << status;
}

// Offset added to exit code from test child to distinguish from other abnormal
// exits.
constexpr int kPrctlNoNewPrivsTestChildExitBase = 100;

TEST(PrctlTest, NoNewPrivsPreservedAcrossCloneForkAndExecve) {
  // Check if no_new_privs is already set. If it is, we can still test that it's
  // preserved across clone/fork/execve, but we also expect it to still be set
  // at the end of the test. Otherwise, call prctl(PR_SET_NO_NEW_PRIVS) so as
  // not to contaminate the original thread.
  int no_new_privs;
  ASSERT_THAT(no_new_privs = prctl(PR_GET_NO_NEW_PRIVS, 0, 0, 0, 0),
              SyscallSucceeds());
  ScopedThread thread = ScopedThread([] {
    ASSERT_THAT(prctl(PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0), SyscallSucceeds());
    EXPECT_THAT(prctl(PR_GET_NO_NEW_PRIVS, 0, 0, 0, 0),
                SyscallSucceedsWithValue(1));
    ScopedThread threadInner = ScopedThread([] {
      EXPECT_THAT(prctl(PR_GET_NO_NEW_PRIVS, 0, 0, 0, 0),
                  SyscallSucceedsWithValue(1));
      // Note that these ASSERT_*s failing will only return from this thread,
      // but this is the intended behavior.
      pid_t child_pid = -1;
      int execve_errno = 0;
      auto cleanup = ASSERT_NO_ERRNO_AND_VALUE(
          ForkAndExec("/proc/self/exe",
                      {"/proc/self/exe", "--prctl_no_new_privs_test_child"}, {},
                      nullptr, &child_pid, &execve_errno));

      ASSERT_GT(child_pid, 0);
      ASSERT_EQ(execve_errno, 0);

      int status = 0;
      ASSERT_THAT(RetryEINTR(waitpid)(child_pid, &status, 0),
                  SyscallSucceeds());
      ASSERT_TRUE(WIFEXITED(status));
      ASSERT_EQ(WEXITSTATUS(status), kPrctlNoNewPrivsTestChildExitBase + 1);

      EXPECT_THAT(prctl(PR_GET_NO_NEW_PRIVS, 0, 0, 0, 0),
                  SyscallSucceedsWithValue(1));
    });
    threadInner.Join();
    EXPECT_THAT(prctl(PR_GET_NO_NEW_PRIVS, 0, 0, 0, 0),
                SyscallSucceedsWithValue(1));
  });
  thread.Join();
  EXPECT_THAT(prctl(PR_GET_NO_NEW_PRIVS, 0, 0, 0, 0),
              SyscallSucceedsWithValue(no_new_privs));
}

TEST(PrctlTest, PDeathSig) {
  pid_t child_pid;

  // Make the new process' parent a separate thread since the parent death
  // signal fires when the parent *thread* exits.
  ScopedThread thread = ScopedThread([&] {
    child_pid = fork();
    TEST_CHECK(child_pid >= 0);
    if (child_pid == 0) {
      // In child process.
      TEST_CHECK(prctl(PR_SET_PDEATHSIG, SIGKILL) >= 0);
      int signo;
      TEST_CHECK(prctl(PR_GET_PDEATHSIG, &signo) >= 0);
      TEST_CHECK(signo == SIGKILL);
      // Enable tracing, then raise SIGSTOP and expect our parent to suppress
      // it.
      TEST_CHECK(ptrace(PTRACE_TRACEME, 0, 0, 0) >= 0);
      TEST_CHECK(raise(SIGSTOP) == 0);
      // Sleep until killed by our parent death signal. sleep(3) is
      // async-signal-safe, absl::SleepFor isn't.
      while (true) {
        sleep(10);
      }
    }
    // In parent process.

    // Wait for the child to send itself SIGSTOP and enter signal-delivery-stop.
    int status;
    ASSERT_THAT(waitpid(child_pid, &status, 0),
                SyscallSucceedsWithValue(child_pid));
    EXPECT_TRUE(WIFSTOPPED(status) && WSTOPSIG(status) == SIGSTOP)
        << "status = " << status;

    // Suppress the SIGSTOP and detach from the child.
    ASSERT_THAT(ptrace(PTRACE_DETACH, child_pid, 0, 0), SyscallSucceeds());
  });
  thread.Join();

  // The child should have been killed by its parent death SIGKILL.
  int status;
  ASSERT_THAT(waitpid(child_pid, &status, 0),
              SyscallSucceedsWithValue(child_pid));
  EXPECT_TRUE(WIFSIGNALED(status) && WTERMSIG(status) == SIGKILL)
      << "status = " << status;
}

// This test is to validate that calling prctl with PR_SET_MM without the
// CAP_SYS_RESOURCE returns EPERM.
TEST(PrctlTest, InvalidPrSetMM) {
  // Drop capability to test below.
  AutoCapability cap(CAP_SYS_RESOURCE, false);
  ASSERT_THAT(prctl(PR_SET_MM, 0, 0, 0, 0), SyscallFailsWithErrno(EPERM));
}

// Sanity check that dumpability is remembered.
TEST(PrctlTest, SetGetDumpability) {
  int before;
  ASSERT_THAT(before = prctl(PR_GET_DUMPABLE), SyscallSucceeds());
  auto cleanup = Cleanup([before] {
    ASSERT_THAT(prctl(PR_SET_DUMPABLE, before), SyscallSucceeds());
  });

  EXPECT_THAT(prctl(PR_SET_DUMPABLE, SUID_DUMP_DISABLE), SyscallSucceeds());
  EXPECT_THAT(prctl(PR_GET_DUMPABLE),
              SyscallSucceedsWithValue(SUID_DUMP_DISABLE));

  EXPECT_THAT(prctl(PR_SET_DUMPABLE, SUID_DUMP_USER), SyscallSucceeds());
  EXPECT_THAT(prctl(PR_GET_DUMPABLE), SyscallSucceedsWithValue(SUID_DUMP_USER));
}

// SUID_DUMP_ROOT cannot be set via PR_SET_DUMPABLE.
TEST(PrctlTest, RootDumpability) {
  EXPECT_THAT(prctl(PR_SET_DUMPABLE, SUID_DUMP_ROOT),
              SyscallFailsWithErrno(EINVAL));
}

TEST(PrctlTest, SimpleSetGetChildSubreaper) {
  // Tasks start off not subreaper.
  int is_subreaper = 0;
  EXPECT_THAT(prctl(PR_GET_CHILD_SUBREAPER, &is_subreaper), SyscallSucceeds());
  EXPECT_EQ(is_subreaper, 0);

  // Set to 1.
  EXPECT_THAT(prctl(PR_SET_CHILD_SUBREAPER, 1), SyscallSucceeds());
  EXPECT_THAT(prctl(PR_GET_CHILD_SUBREAPER, &is_subreaper), SyscallSucceeds());
  EXPECT_EQ(is_subreaper, 1);

  // Set to something positive but not 1.
  EXPECT_THAT(prctl(PR_SET_CHILD_SUBREAPER, 42), SyscallSucceeds());
  // Get still returns 1.
  EXPECT_THAT(prctl(PR_GET_CHILD_SUBREAPER, &is_subreaper), SyscallSucceeds());
  EXPECT_EQ(is_subreaper, 1);

  // Set to something negative.
  EXPECT_THAT(prctl(PR_SET_CHILD_SUBREAPER, -42), SyscallSucceeds());
  // Get still returns 1.
  EXPECT_THAT(prctl(PR_GET_CHILD_SUBREAPER, &is_subreaper), SyscallSucceeds());
  EXPECT_EQ(is_subreaper, 1);
}

TEST(PrctlTest, ThreadsInheritChildSubreaperBit) {
  // Set child subreaper bit.
  ASSERT_THAT(prctl(PR_SET_CHILD_SUBREAPER, 1), SyscallSucceeds());
  ScopedThread thread([&] {
    int is_subreaper = 0;
    ASSERT_THAT(prctl(PR_GET_CHILD_SUBREAPER, &is_subreaper),
                SyscallSucceeds());
    EXPECT_EQ(is_subreaper, 1);
  });
}

TEST(PrctlTest, ProcessesDoNotInheritChildSubreaperBit) {
  // Set child subreaper bit.
  ASSERT_THAT(prctl(PR_SET_CHILD_SUBREAPER, 1), SyscallSucceeds());

  const auto rest = [&] {
    int is_subreaper = 0;
    TEST_CHECK_SUCCESS(prctl(PR_GET_CHILD_SUBREAPER, &is_subreaper));
    TEST_CHECK(is_subreaper == 0);
  };

  EXPECT_THAT(InForkedProcess(rest), IsPosixErrorOkAndHolds(0));
}

static std::atomic<bool> got_sigchild;

void sigchild_handler(int sig, siginfo_t* siginfo, void* arg) {
  got_sigchild = true;
}

TEST(PrctlTest, OrphansReparentedToSubreaper) {
  // Set the subreaper bit.
  ASSERT_THAT(prctl(PR_SET_CHILD_SUBREAPER, 1), SyscallSucceeds());

  // Set up a signal handler to listen for reparented children.
  struct sigaction sa = {};
  sa.sa_sigaction = sigchild_handler;
  sigfillset(&sa.sa_mask);
  auto const sig_cleanup =
      ASSERT_NO_ERRNO_AND_VALUE(ScopedSigaction(SIGCHLD, sa));

  // Execute the test_app zombie_test, which will create an orphaned process
  // and expect that it is reaped.
  constexpr char kTestApp[] = "test/cmd/test_app/test_app";
  const std::string path = RunfilePath(kTestApp);
  int execve_errno;
  pid_t pid;
  auto exec_cleanup = ASSERT_NO_ERRNO_AND_VALUE(
      ForkAndExec(path, {path, "zombie_test"}, {}, &pid, &execve_errno));
  ASSERT_EQ(execve_errno, 0);

  // Wait for 2 children: the process we just started, and the orphan that will
  // be reparented to us.
  for (int i = 0; i < 2; i++) {
    int status;
    int wait_pid;
    ASSERT_THAT(wait_pid = RetryEINTR(waitpid)(-1, &status, 0),
                SyscallSucceeds());
    if (wait_pid == pid) {
      // Test app should have exited cleanly.
      EXPECT_EQ(status, 0);
    }
  }

  // We should have gotten a SIGCHILD for the reparented orphan.
  EXPECT_TRUE(got_sigchild);
}

}  // namespace

}  // namespace testing
}  // namespace gvisor

int main(int argc, char** argv) {
  gvisor::testing::TestInit(&argc, &argv);

  if (absl::GetFlag(FLAGS_prctl_no_new_privs_test_child)) {
    exit(gvisor::testing::kPrctlNoNewPrivsTestChildExitBase +
         prctl(PR_GET_NO_NEW_PRIVS, 0, 0, 0, 0));
  }

  return gvisor::testing::RunAllTests();
}
