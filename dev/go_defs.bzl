load("@io_bazel_rules_go//go:def.bzl", _go_test="go_test")

def go_test(**kwargs):
  _go_test(race="on", **kwargs)
