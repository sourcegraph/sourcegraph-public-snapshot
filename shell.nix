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

{ pkgs ? import <nixpkgs> { }, ... }:

let
  # pkgs.universal-ctags installs the binary as "ctags", not "universal-ctags"
  # like zoekt expects.
  universal-ctags = pkgs.writeScriptBin "universal-ctags" ''
    #!${pkgs.stdenv.shell}
    exec ${pkgs.universal-ctags}/bin/ctags "$@"
  '';

  # need unstable to get the latest version of node. We pin a very specific
  # commit to make this reproducable.
  unstable = import (pkgs.fetchFromGitHub {
    owner = "NixOS";
    repo = "nixpkgs";
    rev = "9675a865c9c3eeec36c06361f7215e109925654c";
    sha256 = "1agsmz77bwdpga9p35ayw4pmwacpa4m31d43c6zdksr7qkknyavx";
  }) { };

in pkgs.mkShell {
  name = "sourcegraph-dev";

  # The packages in the `buildInputs` list will be added to the PATH in our shell
  nativeBuildInputs = with pkgs; [
    # Our core DB.
    pkgs.postgresql_13

    # Cache and some store data
    pkgs.redis

    # Used by symbols and zoekt-git-index to extract symbols from
    # sourcecode.
    pkgs.universal-ctags

    # Build our backend.
    unstable.go_1_17

    # Lots of our tooling and go tests rely on git et al.
    pkgs.git
    pkgs.parallel

    # monitors src files to restart dev services
    pkgs.watchman

    # CI lint tools you need locally
    pkgs.shfmt
    pkgs.shellcheck

    # Web tools. Need node 16.7 so we use unstable. Yarn should also be built
    # against it.
    unstable.nodejs-16_x
    (unstable.yarn.override { nodejs = unstable.nodejs-16_x; })
    unstable.nodePackages.typescript
  ];

  # Startup postgres
  shellHook = ''
    . ./dev/nix/shell-hook.sh
  '';

  # By explicitly setting this environment variable we avoid starting up
  # universal-ctags via docker.
  CTAGS_COMMAND = "${universal-ctags}/bin/universal-ctags";
}
