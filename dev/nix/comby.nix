{ pkgs, pkgsStatic, pkgsMusl, hostPlatform, lib }:
let
  inherit (import ./util.nix { inherit lib; }) makeStatic unNixifyDylibs;
  isMacOS = hostPlatform.isMacOS;
  combyBuilder = ocamlPkgs: systemPkgs:
    (ocamlPkgs.comby.override {
      ocamlPackages = ocamlPkgs.ocamlPackages.overrideScope' (_: prev: {
        ocaml_pcre = prev.ocaml_pcre.override {
          pcre = makeStatic systemPkgs.pcre;
        };
        ssl = prev.ssl.override {
          openssl = makeStatic systemPkgs.openssl;
        };
      });
      sqlite = systemPkgs.sqlite;
      zlib = systemPkgs.zlib;
      libev = (makeStatic systemPkgs.libev).override { static = false; };
      gmp = makeStatic systemPkgs.gmp;
    });
in
if isMacOS then
  unNixifyDylibs pkgs (combyBuilder pkgs pkgs.pkgsStatic)
else
  (combyBuilder pkgs.pkgsMusl pkgs.pkgsStatic).overrideAttrs (_: {
    postPatch = ''
      cat >> src/dune <<EOF
      (env (release (flags  :standard -ccopt -static)))
      EOF
    '';
  })
