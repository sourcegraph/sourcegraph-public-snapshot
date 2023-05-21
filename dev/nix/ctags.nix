{ pkgs, pkgsStatic, lib }:
let
  inherit (import ./util.nix { inherit lib; }) makeStatic unNixifyDylibs;
  pcre2 = makeStatic pkgsStatic.pcre2;
  libyaml = makeStatic pkgsStatic.libyaml;
  jansson = pkgsStatic.jansson.overrideAttrs (oldAttrs: {
    cmakeFlags = [ "-DJANSSON_BUILD_SHARED_LIBS=OFF" ];
  });
in
unNixifyDylibs pkgs ((pkgs.universal-ctags.override {
  # static python is a hassle, and its only used for docs here so we dont care about
  # it being static or not
  inherit (pkgs) python3;
  inherit pcre2 libyaml jansson;
}).overrideAttrs (oldAttrs: {
  version = "5.9.20220403.0";
  src = pkgs.fetchFromGitHub {
    owner = "universal-ctags";
    repo = "ctags";
    rev = "f95bb3497f53748c2b6afc7f298cff218103ab90";
    sha256 = "sha256-pd89KERQj6K11Nue3YFNO+NLOJGqcMnHkeqtWvMFk38=";
  };
  # disable checks, else we get `make[1]: *** No rule to make target 'optlib/cmake.c'.  Stop.`
  doCheck = false;
  checkFlags = [ ];
  # don't include libintl/gettext
  dontAddExtraLibs = true;
  postFixup = (oldAttrs.postFixup or "") + ''
    ln -s $out/bin/ctags $out/bin/universal-ctags
  '';
}))
