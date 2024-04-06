load("@io_bazel_rules_go//go:def.bzl", _go_test = "go_test")

def go_test(**kwargs):
    # All go tests have their timeout set to short by default, unless specified otherwise.
    if "timeout" not in kwargs:
        kwargs["timeout"] = "short"

    # All go tests have the race detector turned on
    if "race" not in kwargs:
        kwargs["race"] = "on"

    # All go tests are tagged with "go" by default
    if "tags" in kwargs:
        if "go" not in kwargs["tags"]:
            kwargs["tags"].append("go")
    else:
        kwargs["tags"] = ["go"]

    _go_test(**kwargs)
