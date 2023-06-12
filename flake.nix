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
          pkgs' = import nixpkgs { inherit system; overlays = builtins.attrValues self.overlays; };
        in
        {
          # We set legacyPackages to our custom static binaries so command
          # like "nix build .#p4-fusion" work.
          legacyPackages = pkgs';

          packages = {
            ctags = pkgs.callPackage ./dev/nix/ctags.nix { };
            comby = pkgs.callPackage ./dev/nix/comby.nix { };
            nodejs-16_x = pkgs.callPackage ./dev/nix/nodejs.nix { };
          }
          # so we don't get `packages.aarch64-linux.p4-fusion` in nix `flake show` output
          // pkgs.lib.optionalAttrs (pkgs.targetPlatform.system != "aarch64-linux") {
            p4-fusion = pkgs.callPackage ./dev/nix/p4-fusion.nix { };
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
        nodejs-16_x = final: prev: { nodejs-16_x = self.packages.${prev.system}.nodejs-16_x; };
        p4-fusion = final: prev: { p4-fusion = self.packages.${prev.system}.p4-fusion; };
      };
    };
}
