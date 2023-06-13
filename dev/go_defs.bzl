load("@io_bazel_rules_go//go:def.bzl", _go_test="go_test")

def go_test(**kwargs):
    # All go tests have their timeout set to short by default, unless specified otherwise.
    if "timeout" not in kwargs:
        kwargs["timeout"] = "short"

    _go_test(race="on", **kwargs)
