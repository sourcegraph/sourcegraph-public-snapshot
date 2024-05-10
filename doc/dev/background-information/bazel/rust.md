# Bazel for Rust

Also checkout the [FAQ](faq.md) for common issues and solutions.

## TL;DR

- Commands:
  - `sg bazel configure rustdeps` after changing workspace members or dependencies.
    - Setting `CARGO_BAZEL_ISOLATED=0` can be set if repinning is too slow, if doing it frequently in local development.
  - `CARGO_BAZEL_REPIN=<crate name> sg bazel configure rustdeps` after adding/updating a dependency.

## Overview

The rules interfacing Rust are named [`rules_rust`](https://github.com/bazelbuild/rules_rust/) and they provide all the plumbing to call the Rust compiler and run the tests.

Bazel and Rust works slightly differently to Bazel and Go. Unlike with Go, `BUILD.bazel` files are not updated with `sg bazel configure` (they must be updated/configured/created by hand), and instead of there being one `BUILD.bazel` file per directory, it's one per workspace member.
There exists [gazelle_rust](https://github.com/Calsign/gazelle_rust), a plugin for [Gazelle](https://github.com/bazelbuild/bazel-gazelle) which is invoked via `sg bazel configure`, that may address the first point but we have decided not to invest in using it at the time of writing.

## Rules for Rust

The rules you'll see for Rust are [`rust_binary`](https://bazelbuild.github.io/rules_rust/defs.html#rust_binary), [`rust_library`](https://bazelbuild.github.io/rules_rust/defs.html#rust_library), [`rust_test`](https://bazelbuild.github.io/rules_rust/defs.html#rust_test) and [`rust_proc_macro`](https://bazelbuild.github.io/rules_rust/defs.html#rust_proc_macro).

Each rule type will have a certain set of common attribute values between them for dependency and proc-macro resolution, alongside others such as `name` and rule-specific attributes with values specific to those particular instances that are each briefly explained in the docs for the relevant rules, found at the links above.

`rust_binary` and `rust_library` targets have the following displayed attribute values in common:

```python
aliases = aliases(),
proc_macro_deps = all_crate_deps(
  proc_macro = True,
),
deps = all_crate_deps(
  normal = True,
),
```

while `rust_test` targets have the following displayed attribute values in common:

```python
aliases = aliases(
  normal_dev = True,
  proc_macro_dev = True,
),
proc_macro_deps = all_crate_deps(
  proc_macro_dev = True,
),
deps = all_crate_deps(
  normal_dev = True,
),
```
