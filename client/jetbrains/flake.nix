{
  description = "The Sourcegraph Jetbrains Client developer environment Nix Flake";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }: flake-utils.lib.eachDefaultSystem (system:
    let
      pkgs = nixpkgs.legacyPackages.${system};
    in
    {
      devShells.default = pkgs.mkShell {
        buildInputs = with pkgs; [
          gradle
          jdk11

          nodejs_20
          nodejs_20.pkgs.pnpm
          nodejs_20.pkgs.typescript
        ];
      };
    });
}
