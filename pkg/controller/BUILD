package(default_visibility = ["//pkg:__subpackages__"])

load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "controller.go",
    ],
    deps = [
        "//pkg/model/config:go_default_library",
        "//pkg/model/osb:go_default_library",
        "@com_github_golang_glog//:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["controller_test.go"],
    library = ":go_default_library",
    deps = [
        "@com_github_davecgh_go_spew//spew:go_default_library",
        "@com_github_golang_mock//gomock:go_default_library",
        "@io_istio_api//:go_default_library",
    ],
)
