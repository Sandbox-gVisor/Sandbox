load("//tools:defs.bzl", "go_library", "go_test")

package(default_applicable_licenses = ["//:license"])

licenses(["notice"])

go_library(
    name = "fdchannel",
    srcs = ["fdchannel_unsafe.go"],
    visibility = ["//visibility:public"],
    deps = [
        "@org_golang_x_sys//unix:go_default_library",
    ],
)

go_test(
    name = "fdchannel_test",
    size = "small",
    srcs = ["fdchannel_test.go"],
    library = ":fdchannel",
    deps = [
        "//pkg/sync",
        "@org_golang_x_sys//unix:go_default_library",
    ],
)
