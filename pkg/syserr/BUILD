load("//tools:defs.bzl", "go_library")

package(
    default_applicable_licenses = ["//:license"],
    licenses = ["notice"],
)

go_library(
    name = "syserr",
    srcs = [
        "host_darwin.go",
        "host_linux.go",
        "netstack.go",
        "syserr.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/abi/linux/errno",
        "//pkg/errors",
        "//pkg/errors/linuxerr",
        "//pkg/safecopy",
        "//pkg/tcpip",
        "@org_golang_x_sys//unix:go_default_library",
    ],
)
