package(default_visibility = ["//visibility:public"])

load("@io_bazel_rules_go//proto:go_proto_library.bzl", "go_proto_library")

go_proto_library(
    name = "go_default_library",
    srcs = ["fake_config.proto"],
)

load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["fake_config.pb.go"],
    deps = ["@com_github_golang_protobuf//proto:go_default_library"],
)

filegroup(
    name = "go_default_library_protos",
    srcs = ["fake_config.proto"],
    visibility = ["//visibility:public"],
)
