load("//tools:defs.bzl", "go_library")

package(
    default_applicable_licenses = ["//:license"],
    licenses = ["notice"],
)

go_library(
    name = "globals",
    srcs = [
        "globals_visitor.go",
        "scope.go",
    ],
    stateify = False,
    visibility = ["//tools/go_generics:__pkg__"],
)
