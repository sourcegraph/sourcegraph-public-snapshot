{ pkgs, pkgsStatic, pkgsMusl, hostPlatform, lib }:
let
  inherit (import ./util.nix { inherit lib; }) mkStatic unNixifyDylibs;
  combyBuilder = ocamlPkgs:
    (ocamlPkgs.comby.override {
      ocamlPackages = ocamlPkgs.ocamlPackages.overrideScope' (_: prev: {
        ocaml_pcre = prev.ocaml_pcre.override {
          pcre = mkStatic pkgsStatic.pcre;
        };
        ssl = prev.ssl.override {
          openssl = mkStatic pkgsStatic.openssl;
        };
      });
      sqlite = pkgsStatic.sqlite;
      # pkgsStatic.zlib.static doesn't exist on linux, but does on macos
      zlib = pkgsStatic.zlib.static or pkgsStatic.zlib;
      # `static = true` from mkStatic is currently broken on macos, noah to fix upstream
      libev = (mkStatic pkgsStatic.libev).override { static = false; };
      gmp = mkStatic pkgsStatic.gmp;
    });
in
if hostPlatform.isMacOS then
  unNixifyDylibs { inherit pkgs; } (combyBuilder pkgs)
else
# ocaml in pkgsStatic is problematic, so we use it from pkgsMusl instead and just
# supply pkgsStatic system libraries such as openssl etc
  (combyBuilder pkgsMusl).overrideAttrs (_: {
    postPatch = ''
      cat >> src/dune <<EOF
      (env (release (flags  :standard -ccopt -static)))
      EOF
    '';
  })
