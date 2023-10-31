{
  description = "The Sourcegraph developer environment & packages Nix Flake";

  inputs = {
    nixpkgs.url = "nixpkgs/nixpkgs-unstable";
    # separate nixpkgs pin for more stable changes to binaries we build
    nixpkgs-stable.url = "github:NixOS/nixpkgs/e78d25df6f1036b3fa76750ed4603dd9d5fe90fc";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, nixpkgs-stable, flake-utils }:
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
          pkgs = import nixpkgs { inherit system; overlays = [self.overlays.nodejs-20_x]; };
          pkgsBins = nixpkgs-stable.legacyPackages.${system};
          pkgs' = import nixpkgs { inherit system; overlays = builtins.attrValues self.overlays; };
          pkgsX = xcompileTargets.${system} or null;
        in
        {
          legacyPackages = pkgs';

          packages = xcompilify { inherit pkgsX; pkgs = pkgsBins; }
            (p: {
              ctags = p.callPackage ./dev/nix/ctags.nix { };
              comby = p.callPackage ./dev/nix/comby.nix { };
              p4-fusion = p.callPackage ./dev/nix/p4-fusion.nix { };
            }) // {
            # doesnt need the same stability as those above
            nodejs-20_x = pkgs.callPackage ./dev/nix/nodejs.nix { };
          };

          # We use pkgs (not pkgs') intentionally to avoid doing extra work of
          # building static comby/universal-ctags in our development
          # environments.
          devShells.default = pkgs.callPackage ./shell.nix { };

          formatter = pkgs.nixpkgs-fmt;
        }) // {
      overlays = {
        ctags = final: prev: { universal-ctags = self.packages.${prev.system}.ctags; };
        comby = final: prev: { comby = self.packages.${prev.system}.comby; };
        nodejs-20_x = final: prev: { nodejs-20_x = self.packages.${prev.system}.nodejs-20_x; };
        p4-fusion = final: prev: { p4-fusion = self.packages.${prev.system}.p4-fusion; };
      };
    };
}
