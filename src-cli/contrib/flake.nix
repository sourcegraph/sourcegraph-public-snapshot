{
  description = "Sourcegraph CLI";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    flake-compat = {
      url = "github:edolstra/flake-compat";
      flake = false;
    };
  };

  outputs = { self, nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      rec {
        packages.src-cli = with pkgs; buildGoModule.override { go = pkgs.go_1_19; } {
          pname = "src-cli";
          version = self.shortRev or "dirty";
          src = ../.;

          subPackages = [ "cmd/src" ];

          vendorSha256 = "sha256-NMLrBYGscZexnR43I4Ku9aqzJr38z2QAnZo0RouHFrc=";
        };
        packages.default = packages.src-cli;
      }
    );
}
