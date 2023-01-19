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

    # Used by symbols and zoekt-git-index to extract symbols from
    # sourcecode.
    universal-ctags

    # Build our backend.
    go_1_19

    # Lots of our tooling and go tests rely on git et al.
    git
    git-lfs
    parallel
    nssTools

    # CI lint tools you need locally
    shfmt
    shellcheck
    golangci-lint

    # Web tools. Need node 16.7 so we use unstable. Yarn should also be built
    # against it.
    nodejs-16_x
    (nodePackages.pnpm.override {
      nodejs = nodejs-16_x;
      version = "7.24.2";
      src = fetchurl {
        url = "https://registry.npmjs.org/pnpm/-/pnpm-7.24.2.tgz";
        sha512 = "sha512-XDTYvZf3xF/kaX0pcdh9GWpak9tV5uDGuNCjkN1SFa0UE350mJGpszmM/j2rVyfoOOFzVR73GNdN3Purd4rXlg==";
      };
    })
    nodePackages.typescript

    # Rust utils for syntax-highlighter service,
    # currently not pinned to the same versions.
    cargo
    rustc
    rustfmt
    libiconv
    clippy
  ];

  # Startup postgres
  shellHook = ''
    . ./dev/nix/shell-hook.sh
  '';

  # Fix for using Delve https://github.com/sourcegraph/sourcegraph/pull/35885
  hardeningDisable = [ "fortify" ];

  # By explicitly setting this environment variable we avoid starting up
  # universal-ctags via docker.
  CTAGS_COMMAND = "${universal-ctags}/bin/universal-ctags";

  RUST_SRC_PATH = "${pkgs.rust.packages.stable.rustPlatform.rustLibSrc}";

  DEV_WEB_BUILDER = "esbuild";
}
