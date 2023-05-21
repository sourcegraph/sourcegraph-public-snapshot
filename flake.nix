{
  description = "The Sourcegraph developer environment & packages Nix Flake";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem
      (system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          pkgs' = pkgs.lib.fold (a: b: b.extend a) pkgs (builtins.attrValues self.overlays);
        in
        {
          packages = {
            ctags = pkgs.callPackage ./dev/nix/ctags.nix { };
            comby = pkgs.callPackage ./dev/nix/comby.nix { };
            nodejs-16_x = pkgs.callPackage ./dev/nix/nodejs.nix { };
          } // pkgs.lib.optionalAttrs (pkgs.targetPlatform.system != "aarch64-linux") {
            p4-fusion = pkgs.callPackage ./dev/nix/p4-fusion.nix { };
          };

          devShells.default = pkgs'.callPackage ./shell.nix { };

          formatter = pkgs.nixpkgs-fmt;
        }) // {
      overlays = {
        ctags = final: prev: { universal-ctags = self.packages.${prev.system}.ctags; };
        comby = final: prev: { comby = self.packages.${prev.system}.comby; };
        nodejs-16_x = final: prev: { nodejs-16_x = self.packages.${prev.system}.nodejs-16_x; };
        p4-fusion = final: prev: { p4-fusion = self.packages.${prev.system}.p4-fusion; };
      };
    };
}
