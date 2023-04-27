{ nixpkgs, lib, utils }:
lib.genAttrs utils.lib.defaultSystems (system:
  let
    inherit (import ./util.nix { inherit (nixpkgs) lib; }) makeStatic unNixifyDylibs;
    pkgs = nixpkgs.legacyPackages.${system};
    isMacOS = nixpkgs.legacyPackages.${system}.hostPlatform.isMacOS;
    combyBuilder = ocamlPkgs: systemDepsPkgs:
      (ocamlPkgs.comby.override {
        sqlite = systemDepsPkgs.sqlite;
        zlib = if isMacOS then systemDepsPkgs.zlib.static else systemDepsPkgs.zlib;
        libev = (makeStatic (systemDepsPkgs.libev)).override { static = false; };
        gmp = makeStatic systemDepsPkgs.gmp;
        ocamlPackages = ocamlPkgs.ocamlPackages.overrideScope' (self: super: {
          ocaml_pcre = super.ocaml_pcre.override {
            pcre = makeStatic systemDepsPkgs.pcre;
          };
          ssl = super.ssl.override {
            openssl = (makeStatic systemDepsPkgs.openssl).override { static = true; };
          };
        });
      });
  in
  if isMacOS then {
    comby = unNixifyDylibs pkgs (combyBuilder pkgs pkgs.pkgsStatic);
  } else {
    comby = (combyBuilder pkgs.pkgsMusl pkgs.pkgsStatic).overrideAttrs (_: {
      postPatch = ''
        cat >> src/dune <<EOF
        (env (release (flags  :standard -ccopt -static)))
        EOF
      '';
    });
  }
)
