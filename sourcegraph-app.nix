{ nixpkgs, lib, utils }: let
    allSystems = [ "aarch64-darwin" ];

    forAllSystems = f: nixpkgs.lib.genAttrs allSystems (system: f {
      pkgs = import nixpkgs { inherit system; };
    });
  in
  forAllSystems ( { pkgs }: {
    sourcegraph-app = pkgs.stdenv.mkDerivation rec {
    name = "sourcegraph-app";
    src = ".";
    pname =  name;
    version = "1.0.0+dev";

    nativeBuildInputs = [
      pkgs.bazel
    ];

    buildPhase = ''
      bazel build //enterprise/cmd/sourcegraph
    '';
  };
  })
