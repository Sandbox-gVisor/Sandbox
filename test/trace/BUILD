load("//tools:defs.bzl", "go_library", "go_test")

package(licenses = ["notice"])

go_test(
    name = "trace_test",
    size = "small",
    srcs = ["trace_test.go"],
    data = [
        "//runsc",
        "//test/trace/workload",
    ],
    library = ":trace",
    tags = [
        "local",
        "manual",
    ],
    deps = [
        "//pkg/sentry/seccheck",
        "//pkg/sentry/seccheck/checkers/remote/test",
        "//pkg/sentry/seccheck/points:points_go_proto",
        "//pkg/test/testutil",
        "//runsc/boot",
        "@org_golang_google_protobuf//proto:go_default_library",
    ],
)

go_library(
    name = "trace",
    srcs = ["trace.go"],
)