# Experimental support for developing in nix. Please reach out to @keegan or @noah if
# you encounter any issues.
#
# Things it does differently:
#
# - Runs postgres under ~/.sourcegraph with a unix socket. No need to manage a
#   service. Must remember to run "pg_ctl stop" if you want to stop it.
#
# Status: everything works on linux. Go1.17 is currently broken on
# darwin. https://github.com/NixOS/nixpkgs/commit/9675a865c9c3eeec36c06361f7215e109925654c
{ pkgs }:
let
  # pkgs.universal-ctags installs the binary as "ctags", not "universal-ctags"
  # like zoekt expects.
  universal-ctags = pkgs.writeScriptBin "universal-ctags" ''
    #!${pkgs.stdenv.shell}
    exec ${pkgs.universal-ctags}/bin/ctags "$@"
  '';

  # We let bazelisk manage the bazel version since we actually need to run two
  # different versions thanks to aspect. Additionally bazelisk allows us to do
  # things like "bazel configure". So we just install a script called bazel
  # which calls bazelisk.
  #
  # Additionally bazel seems to break when CC and CXX is set to a nix managed
  # compiler on darwin. So the script unsets those.
  bazel-wrapper = pkgs.writeScriptBin "bazel" ''
    #!${pkgs.stdenv.shell}
    unset CC CXX
    exec ${pkgs.bazelisk}/bin/bazelisk "$@"
  '';
  bazelStatic = pkgs.bazel.overrideAttrs (oldAttrs: {
    # TODO: test whether putting this here doesn't trigger a build on non-linux
    preBuildPhase = oldAttrs.preBuildPhase + pkgs.lib.optionalString pkgs.hostPlatform.isLinux ''
      export BAZEL_LINKLIBS=-l%:libstdc++.a:-lm
      export BAZEL_LINKOPTS=-static-libstdc++:-static-libgcc
    '';
  });
  bazelStaticWrapped =
    if pkgs.hostPlatform.isLinux then
      pkgs.writeShellScriptBin "bazel" ''exec -a $0 ${bazelStatic}/bin/bazel-6.1.1-linux-${pkgs.hostPlatform.parsed.cpu.name} "$@"''
    else bazelStatic;
  ibazel = pkgs.writeShellScriptBin "ibazel" ''
    exec ${pkgs.bazel-watcher}/bin/ibazel -bazel_path=${if pkgs.hostPlatform.isLinux then bazelStaticWrapped else bazelStatic}/bin/bazel "$@"
  '';
  cargo-bazel = pkgs.rustPlatform.buildRustPackage rec {
    pname = "cargo-bazel";
    version = "0.8.0";
    sourceRoot = "source/crate_universe";
    doCheck = false;

    src = pkgs.fetchFromGitHub {
      owner = "bazelbuild";
      repo = "rules_rust";
      rev = "0.19.0";
      sha256 = "sha256-+tYfw12oELy+x7V8jtGWK0EiNElTwOteO6aUEMlWXio=";
    };

    patches = [
      ./dev/nix/001-rules-rust-cargo-bazel-env.patch
    ];

    cargoSha256 = "sha256-3zFqJrxkHM8MbYkEoThzOJGeFXj9ggTaI+zIL+Hy44I=";
  };
in
pkgs.mkShell {
  name = "sourcegraph-dev";

  # The packages in the `buildInputs` list will be added to the PATH in our shell
  nativeBuildInputs = with pkgs; [
    rnix-lsp

    # Our core DB.
    postgresql_13

    # Cache and some store data
    redis

    # Used by symbols and zoekt-git-index to extract symbols from sourcecode.
    universal-ctags

    # Build our backend.
    go_1_20

    # Lots of our tooling and go tests rely on git et al.
    git
    git-lfs
    parallel
    nssTools

    # CI lint tools you need locally
    shfmt
    shellcheck

    # Web tools. Need node 16.7 so we use unstable. Yarn should also be built against it.
    nodejs-16_x
    (nodejs-16_x.pkgs.pnpm.override {
      version = "7.28.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/pnpm/-/pnpm-7.28.0.tgz";
        sha512 = "sha512-nbuY07S2519jEjaV9KLjSFmOwh0b6KIViIdc/RCJkgco8SZa2+ikQQe4N3CfNn5By5BH0dKVbZ8Ox1Mw8wItSA==";
      };
    })
    nodePackages.typescript

    # Rust utils for syntax-highlighter service, currently not pinned to the same versions.
    cargo
    rustc
    rustfmt
    libiconv
    clippy

    # special sauce bazel stuff.
    bazelisk
    bazel-wrapper
    bazel-watcher
    bazelStaticWrapped
    bazel-buildtools
  ];

  # Startup postgres
  shellHook = ''
    set -h # command hashmap is not guaranteed to be enabled, but required by sg
    . ./dev/nix/shell-hook.sh
  '';

  # Fix for using Delve https://github.com/sourcegraph/sourcegraph/pull/35885
  hardeningDisable = [ "fortify" ];

  # By explicitly setting this environment variable we avoid starting up
  # universal-ctags via docker.
  CTAGS_COMMAND = "${universal-ctags}/bin/universal-ctags";

  RUST_SRC_PATH = "${pkgs.rust.packages.stable.rustPlatform.rustLibSrc}";

  DEV_WEB_BUILDER = "esbuild";

  # not sure if still needed, maybe
  LD_LIBRARY_PATH = pkgs.lib.makeLibraryPath [ pkgs.stdenv.cc.cc.lib pkgs.zlib ];

  CARGO_BAZEL_GENERATOR_URL = "file://${cargo-bazel}/bin/cargo-bazel";
}
