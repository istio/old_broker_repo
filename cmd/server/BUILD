package(default_visibility = ["//visibility:public"])

load("@io_bazel_rules_go//go:def.bzl", "go_binary")

go_binary(
    name = "go_default_library",
    srcs = [
        "main.go",
    ],
    linkstamp = "istio.io/broker/pkg/version",
    deps = [
        "//cmd/shared:go_default_library",
        "//pkg/server:go_default_library",
        "@com_github_spf13_cobra//:go_default_library",
    ],
)
