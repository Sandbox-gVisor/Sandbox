load("//tools:defs.bzl", "go_library")

package(
    default_applicable_licenses = ["//:license"],
    licenses = ["notice"],
)

go_library(
    name = "callbacks",
    srcs = [
        "callback_config.go",
        "rw_hooks.go",
    ],
    visibility = ["//pkg/sentry:internal"],
    imports = [
        "gvisor.dev/gvisor/pkg/hostarch",
        "gvisor.dev/gvisor/pkg/sentry/kernel"
    ],
    deps = [
        "//pkg/sentry/kernel",
        "//pkg/hostarch",
    ],
)
