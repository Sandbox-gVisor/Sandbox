load("//tools:defs.bzl", "go_library")

package(
    default_applicable_licenses = ["//:license"],
    licenses = ["notice"],
)

go_library(
    name = "xdp",
    srcs = [
        "endpoint.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/buffer",
        "//pkg/sync",
        "//pkg/tcpip",
        "//pkg/tcpip/header",
        "//pkg/tcpip/link/qdisc/fifo",
        "//pkg/tcpip/link/rawfile",
        "//pkg/tcpip/link/stopfd",
        "//pkg/tcpip/stack",
        "//pkg/xdp",
        "@org_golang_x_sys//unix:go_default_library",
    ],
)
