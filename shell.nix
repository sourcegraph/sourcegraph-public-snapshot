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

  # The version required of node is ahead of what is available in the
  # registry, so we build a custom version.
  node16_7 =
    pkgs.callPackage "${<nixpkgs>}/pkgs/development/web/nodejs/nodejs.nix" {
      python = pkgs.python3;
    } {
      enableNpm = true;
      version = "16.7.0";
      sha256 = "0drd7zyadjrhng9k0mspz456j3pmr7kli5dd0kx8grbqsgxzv1gs";
    };

  # Build yarn against the node we use.
  yarn = pkgs.yarn.override { nodejs = node16_7; };

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
    pkgs.go

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

    # Web tools
    node16_7
    yarn
    pkgs.nodePackages.typescript
  ];

  # Startup postgres
  shellHook = ''
    . ./dev/nix/shell-hook.sh
  '';

  # By explicitly setting this environment variable we avoid starting up
  # universal-ctags via docker.
  CTAGS_COMMAND = "${universal-ctags}/bin/universal-ctags";
}
