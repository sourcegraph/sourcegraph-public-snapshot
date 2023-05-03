# Experimental support for developing in nix. Please reach out to @keegan or @noah if
# you encounter any issues.
#
# Things it does differently:
#
# - Runs postgres under ~/.sourcegraph with a unix socket. No need to manage a
#   service. Must remember to run "pg_ctl stop" if you want to stop it.
# - Builds bazel statically (on linux) and configures it to use toolchains for nodejs
#   and rust to use the ones provided by rules_nixpkgs
#
# Status: everything works on linux & darwin.
{ pkgs }:
let
  # pkgs.universal-ctags installs the binary as "ctags", not "universal-ctags"
  # like zoekt expects.
  universal-ctags = pkgs.writeShellScriptBin "universal-ctags" ''
    exec ${pkgs.universal-ctags}/bin/ctags "$@"
  '';

  # On darwin, we let bazelisk manage the bazel version since we actually need to run two
  # different versions thanks to aspect. Additionally bazelisk allows us to do
  # things like "bazel configure". So we just install a script called bazel
  # which calls bazelisk.
  #
  # On linux, we use a from-source statically built bazel (due to libstdc++ woes) for all commands
  # besides 'configure', where we transparently defer to bazelisk (which defers to aspect cli).
  #
  # Additionally bazel seems to break when CC and CXX is set to a nix managed
  # compiler on darwin. So the script unsets those.
  bazel-wrapper = pkgs.writeShellScriptBin "bazel" (if pkgs.hostPlatform.isMacOS then ''
    unset CC CXX
    exec ${pkgs.bazelisk}/bin/bazelisk "$@"
  '' else ''
    if [ "$1" == "configure" ]; then
      exec ${pkgs.bazelisk}/bin/bazelisk "$@"
    fi
    exec ${bazel-static}/bin/bazel "$@"
  '');
  bazel-static = pkgs.bazel_6.overrideAttrs (oldAttrs: {
    preBuildPhase = oldAttrs.preBuildPhase + ''
      export BAZEL_LINKLIBS=-l%:libstdc++.a:-lm
      export BAZEL_LINKOPTS=-static-libstdc++:-static-libgcc
    '';
  });
  bazel-watcher = pkgs.writeShellScriptBin "ibazel" ''
    exec ${pkgs.bazel-watcher}/bin/ibazel \
      ${pkgs.lib.optionalString pkgs.hostPlatform.isLinux "-bazel_path=${bazel-static}/bin/bazel"} "$@"
  '';
  # custom cargo-bazel so we can pass down LD_LIBRARY_PATH, see definition of LD_LIBRARY_PATH below
  # for more info.
  cargo-bazel = pkgs.rustPlatform.buildRustPackage rec {
    pname = "cargo-bazel";
    version = "0.8.0";
    sourceRoot = "source/crate_universe";
    doCheck = false;

    buildInputs = [ ] ++ pkgs.lib.optional pkgs.hostPlatform.isMacOS [
      pkgs.darwin.Security
    ];

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
    # nix language server
    nil

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
    bazel-buildtools
  ];

  # Startup postgres, redis & set nixos specific stuff
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

  # Needed for rules_rust provisioned rust tools when running `bazel(isk) configure`, still need
  # nixpkgs_rust_configure for actual compilation step
  LD_LIBRARY_PATH = pkgs.lib.makeLibraryPath [ pkgs.stdenv.cc.cc.lib pkgs.zlib ];

  # Tell rules_rust to use our custom cargo-bazel.
  CARGO_BAZEL_GENERATOR_URL = "file://${cargo-bazel}/bin/cargo-bazel";

  # bazel complains when the bazel version differs even by a patch version to whats defined in .bazelversion,
  # so we tell it to h*ck off here.
  # https://sourcegraph.com/github.com/bazelbuild/bazel@1a4da7f331c753c92e2c91efcad434dc29d10d43/-/blob/scripts/packages/bazel.sh?L23-28
  USE_BAZEL_VERSION =
    if pkgs.hostPlatform.isMacOS then "" else pkgs.bazel_6.version;
}
