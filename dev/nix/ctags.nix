{ pkgs, pkgsStatic, fetchFromGitHub, lib }:
let
  inherit (import ./util.nix { inherit lib; }) mkStatic unNixifyDylibs removePackagesByName;
  pcre2 = mkStatic pkgsStatic.pcre2;
  libyaml = mkStatic pkgsStatic.libyaml;
  jansson = pkgsStatic.jansson.overrideAttrs (oldAttrs: {
    cmakeFlags = [ "-DJANSSON_BUILD_SHARED_LIBS=OFF" ];
  });
in
unNixifyDylibs pkgs ((pkgsStatic.universal-ctags.override {
  # static python is a hassle, and its only used for docs here so we dont care about
  # it being static or not
  inherit (pkgs) python3;
  inherit pcre2 libyaml jansson;
}).overrideAttrs (oldAttrs: {
  version = "5.9.20220403.0";
  src = fetchFromGitHub {
    owner = "universal-ctags";
    repo = "ctags";
    rev = "f95bb3497f53748c2b6afc7f298cff218103ab90";
    sha256 = "sha256-pd89KERQj6K11Nue3YFNO+NLOJGqcMnHkeqtWvMFk38=";
  };
  enableParallelBuilding = true; # must be enabled on a per-package basis
  # pkgsStatic moves `buildInputs` into `propagatedBuildInputs`, and we don't want sandbox support so we remove
  # libseccomp so it autoconfigures to no sandbox
  propagatedBuildInputs = (removePackagesByName (oldAttrs.propagatedBuildInputs or [ ]) [ pkgsStatic.libseccomp ]);
  # 1) check-genfile tries to perform git operations, but .git is removed by default (and leaveDotGit not recommended)
  #    due to being non-deterministic https://github.com/NixOS/nixpkgs/issues/8567
  # 2) tutil tests may be for libseccomp builds only:
  #    `make: *** No rule to make target 'tutil'.  Stop`
  checkFlags = lib.remove "tutil" (oldAttrs.checkFlags ++ [ "-o" "check-genfile" ]);
  # don't include libintl/gettext on macos
  dontAddExtraLibs = true;
  postFixup = (oldAttrs.postFixup or "") + ''
    ln -s $out/bin/ctags $out/bin/universal-ctags
  '';
}))
