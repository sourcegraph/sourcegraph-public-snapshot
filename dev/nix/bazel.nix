{ nixpkgs
, pkgs
, bazel_7
, lib
, substituteAll
, bash
, coreutils
, diffutils
, file
, findutils
, gawk
, gnugrep
, gnupatch
, gnused
, gnutar
, gzip
, python3
, unzip
, which
, zip
}: {
  bazel_7 = bazel_7.overrideAttrs (oldAttrs:
    let
      # yoinked from https://sourcegraph.com/github.com/NixOS/nixpkgs/-/blob/pkgs/development/tools/build-managers/bazel/bazel_7/default.nix?L77-120
      defaultShellUtils = [
        bash
        coreutils
        diffutils
        file
        findutils
        gawk
        gnugrep
        gnupatch
        gnused
        gnutar
        gzip
        python3
        unzip
        which
        zip
      ];
    in
    {
      # https://github.com/NixOS/nixpkgs/pull/262152#issuecomment-1879053113
      patches = (oldAttrs.patches or [ ]) ++ [
        (substituteAll {
          src = "${nixpkgs}/pkgs/development/tools/build-managers/bazel/bazel_6/actions_path.patch";
          actionsPathPatch = lib.makeBinPath defaultShellUtils;
        })
      ];
    });
}
