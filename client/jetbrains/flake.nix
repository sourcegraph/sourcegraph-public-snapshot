{
  description = "The Sourcegraph Jetbrains Client developer environment Nix Flake";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }: flake-utils.lib.eachDefaultSystem (system:
    let
      pkgs = nixpkgs.legacyPackages.${system};
      libraries = pkgs.lib.makeLibraryPath (with pkgs; with xorg; [
        libXtst
        libXext
        libX11
        libXrender
        libXi
        freetype
        fontconfig.lib
        zlib
        libsecret
      ]);
      gradle-wrapped = pkgs.writeShellScriptBin "gradle" ''
        export LD_LIBRARY_PATH=${libraries}
        exec ${pkgs.gradle.override {
            javaToolchains = [ "${pkgs.jdk8}/lib/openjdk" "${pkgs.jdk11}/lib/openjdk" "${pkgs.jdk17}/lib/openjdk" ];
          }}/bin/gradle "$@"
      '';
    in
    {
      devShells.default = pkgs.mkShell {
        nativeBuildInputs = with pkgs; [
          gradle-wrapped
          nodejs_20
          nodejs_20.pkgs.pnpm
          nodejs_20.pkgs.typescript
        ];
      };
    });
}
