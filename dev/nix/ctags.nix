{ pkgs
, lib
, autoreconfHook
, perl
, pkg-config
, coreutils
, pkgsStatic
, fetchFromGitHub
, buildPackages
}:
let
  inherit (import ./util.nix { inherit lib; }) mkStatic unNixifyDylibs;
  pcre2 = mkStatic pkgsStatic.pcre2;
  libyaml = mkStatic pkgsStatic.libyaml;
  jansson = pkgsStatic.jansson.overrideAttrs (oldAttrs: {
    cmakeFlags = [ "-DJANSSON_BUILD_SHARED_LIBS=OFF" ];
  });
  stdenv = pkgsStatic.stdenv;
in
# yoinked from github.com/nixos/nixpkgs
unNixifyDylibs { inherit pkgs; } (stdenv.mkDerivation rec {
  pname = "universal-ctags";
  version = "6.0.0";

  src = fetchFromGitHub {
    owner = "universal-ctags";
    repo = "ctags";
    rev = "v${version}";
    sha256 = "sha256-XlqBndo8g011SDGp3zM7S+AQ0aCp6rpQlqJF6e5Dd6w=";
  };

  depsBuildBuild = [
    buildPackages.stdenv.cc
  ];

  nativeBuildInputs = [
    autoreconfHook
    perl
    pkg-config
  ];

  buildInputs = [
    libyaml
    pcre2
    jansson
    pkgsStatic.libxml2
  ]
  ++ lib.optional stdenv.isDarwin pkgsStatic.libiconv;

  configureFlags = [ "--enable-tmpdir=/tmp" ];

  dontAddExtraLibs = true;

  patches = [
    "${pkgs.path}/pkgs/development/tools/misc/universal-ctags/000-nixos-specific.patch"
  ];

  postPatch = ''
    substituteInPlace Tmain/utils.sh \
      --replace /bin/echo ${coreutils}/bin/echo

    patchShebangs misc/*
  '';

  # we create two symbolic links
  # 1. ctags-$version: Used bazel in dev/tool_deps.bzl. With the version it allows us to pin the ctags version bazel uses
  # 2. ctags: gets referenced by shell.nix and there is no version to ensure we always point to the latest.
  postFixup = ''
    ln -s $out/bin/ctags $out/bin/universal-ctags-$version
    ln -s $out/bin/ctags $out/bin/universal-ctags
  '';

  doCheck = true;

  checkTarget = [
    "tlib"
    "tmain"
    "units"
  ];
  # disable check-genfile, this attempts to run some git commands
  # which arent supported as we dont have/include .git
  checkFlags = [
    "-o"
    "check-genfile"
  ];

  # must be enabled on a per-package basis
  enableParallelBuilding = true;
  enableParallelChecking = true;

  meta = with lib; {
    homepage = "https://docs.ctags.io/en/latest/";
    description = "A maintained ctags implementation";
    longDescription = ''
      Universal Ctags (abbreviated as u-ctags) is a maintained implementation of
      ctags. ctags generates an index (or tag) file of language objects found in
      source files for programming languages. This index makes it easy for text
      editors and other tools to locate the indexed items.
    '';
    license = licenses.gpl2Plus;
    maintainers = [ maintainers.AndersonTorres ];
    platforms = platforms.all;
    mainProgram = "ctags";
    priority = 1; # over the emacs implementation
  };
})
