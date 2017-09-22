package(default_visibility = ["//pkg:__subpackages__"])

load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "controller.go",
        "file.go",
    ],
    deps = [
        "//pkg/model/config:go_default_library",
        "//pkg/model/osb:go_default_library",
        "@com_github_golang_glog//:go_default_library",
    ],
)
