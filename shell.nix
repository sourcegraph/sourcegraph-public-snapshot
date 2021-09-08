# Experimental support for developing in nix. Please reach out to @keegan if
# you encounter any issues.
#
# Things it does differently:
#
# - Runs postgres under ~/.sourcegraph with a unix socket. No need to manage a
#   service. Must remember to run "pg_ctl stop" if you want to stop it.
#
# Status: go test ./... and yarn works

{ pkgs ? import <nixpkgs> { }, ... }:

let
  # pkgs.universal-ctags installs the binary as "ctags", not "universal-ctags"
  # like zoekt expects.
  universal-ctags = pkgs.writeScriptBin "universal-ctags" ''
    #!${pkgs.stdenv.shell}
    exec ${pkgs.universal-ctags}/bin/ctags "$@"
  '';

  # go1.17 is not yet available in nixpkgs and is held up due to go requiring
  # macOS 10.13 or later. So we use official go releases.
  go_1_17 =
    pkgs.callPackage "${<nixpkgs>}/pkgs/development/compilers/go/binary.nix" {
      version = "1.17";
      hashes = {
        linux-amd64 =
          "6bf89fc4f5ad763871cf7eac80a2d594492de7a818303283f1366a7f6a30372d";
        darwin-amd64 =
          "355bd544ce08d7d484d9d7de05a71b5c6f5bc10aa4b316688c2192aeb3dacfd1";
      };
    };

  # need unstable to get the latest version of node. We pin a very specific
  # commit to make this reproducable.
  unstable = import (pkgs.fetchFromGitHub {
    owner = "NixOS";
    repo = "nixpkgs";
    rev = "f3706ab27f99b2ffdaeb6dd03ee6e2f26511c6db";
    sha256 = "1fb3z0y08y1jjhzffsg4qa5y9mk434s167n55avcwbqqjwd7kj1c";
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
    universal-ctags

    # Build our backend.
    go_1_17

    # Lots of our tooling and go tests rely on git.
    pkgs.git

    # cgo dependency for symbols. TODO build with nix?
    pkgs.pcre
    pkgs.sqlite
    pkgs.pkg-config

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

  # Official go release expects GOROOT to be /usr/local/go. While we are using
  # the official release we need to point it to the correct place and disable
  # CGO.
  GOROOT = "${go_1_17}/share/go";
  CGO_ENABLED = "0";
}
