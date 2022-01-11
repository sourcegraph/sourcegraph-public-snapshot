# Experimental support for developing in nix. Please reach out to @keegan if
# you encounter any issues.
#
# Things it does differently:
#
# - Runs postgres under ~/.sourcegraph with a unix socket. No need to manage a
#   service. Must remember to run "pg_ctl stop" if you want to stop it.
#
# Status: everything works on linux. Go1.17 is currently broken on
# darwin. https://github.com/NixOS/nixpkgs/commit/9675a865c9c3eeec36c06361f7215e109925654c

# Pin a specific version of nixpkgs to ensure we get the same packages.
{ pkgs ? import (fetchTarball {
  url =
    "https://github.com/NixOS/nixpkgs/archive/2b914ee8e20c7082b18a550bd93e1e7b384adc0f.tar.gz";
  sha256 = "0qnhrm4ywci89kvl35vwd85ldid3g1z74gqsj1b4hw06186dvcnp";
}) { }, ... }:

let
  # pkgs.universal-ctags installs the binary as "ctags", not "universal-ctags"
  # like zoekt expects.
  universal-ctags = pkgs.writeScriptBin "universal-ctags" ''
    #!${pkgs.stdenv.shell}
    exec ${pkgs.universal-ctags}/bin/ctags "$@"
  '';

in pkgs.mkShell {
  name = "sourcegraph-dev";

  # The packages in the `buildInputs` list will be added to the PATH in our shell
  nativeBuildInputs = with pkgs; [
    # Our core DB.
    postgresql_13

    # Cache and some store data
    redis

    # Used by symbols and zoekt-git-index to extract symbols from
    # sourcecode.
    universal-ctags

    # Build our backend.
    go_1_17

    # Lots of our tooling and go tests rely on git et al.
    git
    parallel

    # monitors src files to restart dev services
    watchman

    # CI lint tools you need locally
    shfmt
    shellcheck
    golangci-lint

    # Web tools. Need node 16.7 so we use unstable. Yarn should also be built
    # against it.
    nodejs-16_x
    (yarn.override { nodejs = nodejs-16_x; })
    nodePackages.typescript
  ];

  # Startup postgres
  shellHook = ''
    . ./dev/nix/shell-hook.sh
  '';

  # By explicitly setting this environment variable we avoid starting up
  # universal-ctags via docker.
  CTAGS_COMMAND = "${universal-ctags}/bin/universal-ctags";
}
