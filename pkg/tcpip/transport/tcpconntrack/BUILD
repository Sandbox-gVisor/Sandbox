load("//tools:defs.bzl", "go_library", "go_test")

package(
    default_applicable_licenses = ["//:license"],
    licenses = ["notice"],
)

go_library(
    name = "tcpconntrack",
    srcs = ["tcp_conntrack.go"],
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/tcpip/header",
        "//pkg/tcpip/seqnum",
    ],
)

go_test(
    name = "tcpconntrack_test",
    size = "small",
    srcs = ["tcp_conntrack_test.go"],
    deps = [
        ":tcpconntrack",
        "//pkg/refs",
        "//pkg/tcpip/header",
    ],
)
