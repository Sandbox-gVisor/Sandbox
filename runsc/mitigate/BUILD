load("//tools:defs.bzl", "go_library", "go_test")

package(
    default_applicable_licenses = ["//:license"],
    licenses = ["notice"],
)

go_library(
    name = "mitigate",
    srcs = [
        "mitigate.go",
        "mock.go",
    ],
    visibility = [
        "//runsc:__subpackages__",
    ],
    deps = ["@in_gopkg_yaml_v2//:go_default_library"],
)

go_test(
    name = "mitigate_test",
    size = "small",
    srcs = ["mitigate_test.go"],
    library = ":mitigate",
    deps = ["@com_github_google_go_cmp//cmp:go_default_library"],
)
