"Overrides rust_test, rust_binary and rust_library to automatically declare rust_clippy targets."

load("@rules_rust//rust:defs.bzl", "rust_clippy", _rust_binary = "rust_binary", _rust_library = "rust_library", _rust_test = "rust_test")

def rust_binary(**kwargs):
    _rust_binary(**kwargs)
    rust_clippy(
        name = "{}_clippy".format(kwargs["name"]),
        testonly = True,
        deps = [":{}".format(kwargs["name"])],
    )

def rust_test(**kwargs):
    _rust_test(**kwargs)
    rust_clippy(
        name = "{}_clippy".format(kwargs["name"]),
        testonly = True,
        deps = [":{}".format(kwargs["name"])],
    )

def rust_library(**kwargs):
    _rust_library(**kwargs)
    rust_clippy(
        name = "{}_clippy".format(kwargs["name"]),
        testonly = True,
        deps = [":{}".format(kwargs["name"])],
    )
