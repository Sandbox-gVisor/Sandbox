# setup_container contains a shim binary that runs within the test container
# for syscall tests with container=True.

load("//tools:defs.bzl", "cc_binary")

package(
    default_applicable_licenses = ["//:license"],
    licenses = ["notice"],
)

cc_binary(
    name = "setup_container",
    testonly = 1,
    srcs = ["setup_container.cc"],
    visibility = ["//test/runner:__subpackages__"],
    deps = [
        "//test/syscalls/linux:socket_netlink_util",
        "//test/util:capability_util",
        "//test/util:posix_error",
        "//test/util:socket_util",
    ],
)
