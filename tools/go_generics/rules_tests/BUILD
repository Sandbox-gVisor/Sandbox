load("//tools:defs.bzl", "go_test")
load("//tools/go_generics:defs.bzl", "go_template", "go_template_instance")

package(
    default_applicable_licenses = ["//:license"],
    licenses = ["notice"],
)

go_template_instance(
    name = "instance",
    out = "instance_test.go",
    consts = {
        "n": "20",
        "m": "\"test\"",
        "o": "math.MaxUint64",
    },
    imports = {
        "math": "math",
    },
    package = "template_test",
    template = ":test_template",
    types = {
        "t": "int",
    },
)

go_template(
    name = "test_template",
    srcs = [
        "template.go",
    ],
    opt_consts = [
        "n",
        "m",
        "o",
    ],
    opt_types = ["t"],
)

go_test(
    name = "template_test",
    srcs = [
        "instance_test.go",
        "template_test.go",
    ],
)
