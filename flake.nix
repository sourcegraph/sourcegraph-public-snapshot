{
  description = "The Sourcegraph developer environment & packages Nix Flake";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    let
      xcompileTargets = {
        "aarch64-darwin" = nixpkgs.legacyPackages.aarch64-darwin.pkgsx86_64Darwin;
        "x86_64-darwin" = import nixpkgs { system = "x86_64-darwin"; crossSystem = nixpkgs.lib.systems.examples.aarch64-darwin; };
      };
    in
    flake-utils.lib.eachDefaultSystem
      (system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          pkgs' = import nixpkgs { inherit system; overlays = builtins.attrValues self.overlays; };
          pkgsX = xcompileTargets.${system} or null;
          xcompilify = f: with nixpkgs; lib.foldl'
            (acc: val: acc // (lib.mapAttrs'
              # rename with -${xtarget} if an xtarget package
              (name: value:
                lib.nameValuePair
                  (name + lib.optionalString (value.system != system) "-${value.system}")
                  value
              )
              val))
            { }
            (builtins.map (val: f val) ([ pkgs ] ++ lib.optional (pkgsX != null) pkgsX))
          ;
        in
        {
          legacyPackages = pkgs';

          packages = (xcompilify (pkgs: {
            ctags = pkgs.callPackage ./dev/nix/ctags.nix { };
            comby = pkgs.callPackage ./dev/nix/comby.nix { };
            p4-fusion = pkgs.callPackage ./dev/nix/p4-fusion.nix { };
          })) // {
            nodejs-16_x = pkgs.callPackage ./dev/nix/nodejs.nix { };
          };

          # We use pkgs (not pkgs') intentionally to avoid doing extra work of
          # building static comby/universal-ctags in our development
          # environments.
          devShells.default = pkgs.callPackage ./shell.nix { };

          formatter = pkgs.nixpkgs-fmt;
        }) // {
      overlays = {
        ctags = final: prev: {
          universal-ctags = self.packages.${prev.system}.ctags;
        };
        comby = final: prev: { comby = self.packages.${prev.system}.comby; };
        nodejs-16_x = final: prev: { nodejs-16_x = self.packages.${prev.system}.nodejs-16_x; };
        p4-fusion = final: prev: { p4-fusion = self.packages.${prev.system}.p4-fusion; };
      };
    };
}
