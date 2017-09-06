load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["config.go"],
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/model/config:go_default_library",
        "//pkg/testing/mock/proto:go_default_library",
        "@com_github_golang_protobuf//proto:go_default_library",
    ],
)
