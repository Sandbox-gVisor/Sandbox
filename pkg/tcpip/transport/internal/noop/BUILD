load("//tools:defs.bzl", "go_library")

package(
    default_applicable_licenses = ["//:license"],
    licenses = ["notice"],
)

go_library(
    name = "noop",
    srcs = ["endpoint.go"],
    visibility = ["//pkg/tcpip/transport/raw:__pkg__"],
    deps = [
        "//pkg/tcpip",
        "//pkg/tcpip/stack",
        "//pkg/waiter",
    ],
)
