{
  description = "The Sourcegraph developer environment & packages Nix Flake";

  nixConfig = {
    extra-substituters = [ "https://sourcegraph-noah.cachix.org" ];
    extra-trusted-public-keys = [ "sourcegraph-noah.cachix.org-1:rTTKnyuUmJuGt/UAXUpdOCOXDAfaO1AYy+/jSre3XgA=" ];
  };

  inputs = {
    nixpkgs.url = "nixpkgs/nixpkgs-unstable";
    nixpkgs-bazel.url = "github:Strum355/nixpkgs/bazel-7.1.0";
    # separate nixpkgs pin for more stable changes to binaries we build
    nixpkgs-stable.url = "github:NixOS/nixpkgs/e78d25df6f1036b3fa76750ed4603dd9d5fe90fc";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, nixpkgs-stable, flake-utils, nixpkgs-bazel }:
    let
      xcompileTargets = with nixpkgs-stable.lib.systems.examples; {
        "aarch64-darwin" = nixpkgs-stable.legacyPackages.aarch64-darwin.pkgsx86_64Darwin;
        "x86_64-darwin" = import nixpkgs-stable { system = "x86_64-darwin"; crossSystem = aarch64-darwin; };
      };
      inherit (import ./dev/nix/util.nix { inherit (nixpkgs) lib; }) xcompilify;
    in
    flake-utils.lib.eachDefaultSystem
      (system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          pkgsShell = import nixpkgs { inherit system; overlays = with self.overlays; [ nodejs-20_x bazel_7 go_1_22 ]; };
          pkgsBins = nixpkgs-stable.legacyPackages.${system};
          pkgsAll = import nixpkgs { inherit system; overlays = builtins.attrValues self.overlays; };
          pkgsX = xcompileTargets.${system} or null;
        in
        {
          legacyPackages = pkgsAll;

          packages = xcompilify { inherit pkgsX; pkgs = pkgsBins; }
            (p: {
              ctags = p.callPackage ./dev/nix/ctags.nix { };
              comby = p.callPackage ./dev/nix/comby.nix { };
              p4-fusion = p.callPackage ./dev/nix/p4-fusion.nix { };
            }) // {
            # doesnt need the same stability as those above
            nodejs-20_x = pkgs.callPackage ./dev/nix/nodejs.nix { };
            bazel_7 = nixpkgs-bazel.legacyPackages.${system}.callPackage ./dev/nix/bazel.nix { };
            pg-utils = (if pkgs.hostPlatform.isMacOS then pkgs.callPackage else pkgs.pkgsStatic.callPackage) ./dev/nix/pg-utils.nix {
              # tzdata fails to build on pkgsStatic, and pkgsMusl isnt supported on macos
              tzdata = if pkgs.hostPlatform.isMacOS then pkgs.tzdata else pkgs.pkgsMusl.tzdata;
            };
            go_1_22 = pkgs.callPackage ./dev/nix/go_1_22.nix { };
          };

          # We use pkgsShell (not pkgsAll) intentionally to avoid doing extra work of
          # building static comby/universal-ctags in our development
          # environments.
          devShells.default = pkgsShell.callPackage ./shell.nix { };

          formatter = pkgs.nixpkgs-fmt;
        }) // {
      overlays = {
        ctags = final: prev: { universal-ctags = self.packages.${prev.system}.ctags; };
        comby = final: prev: { comby = self.packages.${prev.system}.comby; };
        nodejs-20_x = final: prev: { nodejs-20_x = self.packages.${prev.system}.nodejs-20_x; };
        p4-fusion = final: prev: { p4-fusion = self.packages.${prev.system}.p4-fusion; };
        bazel_7 = final: prev: { bazel_7 = self.packages.${prev.system}.bazel_7; };
        pg-utils = final: prev: { pg-utils = self.packages.${prev.system}.pg-utils; };
        go_1_22 = final: prev: { go_1_22 = self.packages.${prev.system}.go_1_22; };
      };
    };
}
