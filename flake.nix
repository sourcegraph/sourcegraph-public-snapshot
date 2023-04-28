{
  description = "The Sourcegraph developer environment & packages Nix Flake";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
    utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, utils }:
    with nixpkgs.lib; with utils.lib; {
      devShells = genAttrs defaultSystems (system:
        let
          pkgs = import nixpkgs { inherit system; overlays = [ self.overlays.ctags ]; };
        in
        {
          default = pkgs.callPackage ./shell.nix { };
        }
      );

      # Pin a specific version of universal-ctags to the same version as in cmd/symbols/ctags-install-alpine.sh.
      overlays.ctags = (import ./dev/nix/ctags.nix { inherit nixpkgs utils; inherit (nixpkgs) lib; }).overlay;

      packages = fold recursiveUpdate { } [
        ((import ./dev/nix/ctags.nix { inherit nixpkgs utils; inherit (nixpkgs) lib; }).packages)
        (import ./dev/nix/p4-fusion.nix { inherit nixpkgs utils; inherit (nixpkgs) lib; })
        (import ./dev/nix/comby.nix { inherit nixpkgs utils; inherit (nixpkgs) lib; })
      ];
    };
}
