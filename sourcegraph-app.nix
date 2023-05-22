{ nixpkgs, lib, utils }: let
    allSystems = [ "aarch64-darwin" ];

    forAllSystems = f: nixpkgs.lib.genAttrs allSystems (system: f {
      pkgs = import nixpkgs { inherit system; };
    });
  in
  forAllSystems ( { pkgs }: {
    sourcegraph-app = pkgs.buildBazelPackage {
    bazel = pkgs.bazel_6;
    name = "sourcegraph-nix";
    bazelTargets = ["//enterprise/cmd/sourcegraph:sourcegraph"];
    src = ./.;
    fetchAttrs = {
        preBuild = ''
          echo ${pkgs.bazel_6.version} > .bazelversion
        '';
        sha256 = "";
      };
    buildAttrs = {
        outputs = [ "out" ];

        USE_BAZEL_VERSION =
          if pkgs.hostPlatform.isMacOS then "" else pkgs.bazel_6.version;

        nativeBuildInputs = [ pkgs.libtool ];

        preBuild = ''
          echo ${pkgs.bazel_6.version} > .bazelversion
        '';

        bazelFlags = [
        "--announce-rc"
        ];
      };
  };
  })
