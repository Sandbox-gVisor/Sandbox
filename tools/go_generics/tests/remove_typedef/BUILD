load("//tools/go_generics/tests:defs.bzl", "go_generics_test")

package(default_applicable_licenses = ["//:license"])

go_generics_test(
    name = "remove_typedef",
    inputs = ["input.go"],
    output = "output.go",
    types = {
        "T": "U",
    },
)

# @unused
glaze_ignore = [
    "input.go",
    "output.go",
]
