load("//tools:defs.bzl", "go_library")

package(
    default_applicable_licenses = ["//:license"],
    licenses = ["notice"],
)

go_library(
    name = "callbacks",
    srcs = [
        "callback_config.go",
        "util.go"
    ],
    visibility = ["//pkg/sentry:internal"],
    imports = [

    ],
    deps = [
        "//pkg/sentry/arch",
        "@github_com_dop251_goja//:go_default_library",
        "@github_com_dop251_goja//parser:go_default_library",
    ],
)
