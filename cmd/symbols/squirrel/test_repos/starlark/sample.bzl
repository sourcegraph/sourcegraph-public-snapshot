#                          vvvvvvvvvvvvv bzl_library ref
load("//:bzl_library.bzl", "bzl_library", custom_bzl_library = "bzl_library")

#   vvvvv _impl def
def _impl():
    pass

#                  vvvvv _impl ref
bzl_library(impl = _impl)
custom_bzl_library(impl = _impl)
