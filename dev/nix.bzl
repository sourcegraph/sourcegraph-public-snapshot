load("@io_tweag_rules_nixpkgs//core:nixpkgs.bzl", "nixpkgs_local_repository")
load("@io_tweag_rules_nixpkgs//toolchains/nodejs:nodejs.bzl", "nixpkgs_nodejs_configure")
load("@io_tweag_rules_nixpkgs//toolchains/rust:rust.bzl", "nixpkgs_rust_configure")

def nix_deps():
    nixpkgs_local_repository(
        name = "nixpkgs",
        nix_flake_lock_file = "//:flake.lock",
    )

    nixpkgs_nodejs_configure(
        name = "nixpkgs_nodejs",
        attribute_path = "nodejs-16_x",
        register = False,
        repository = "@nixpkgs",
    )

    nixpkgs_rust_configure(
        name = "nixpkgs_rust",
        default_edition = "2021",
        register = False,
        repository = "@nixpkgs",
    )
