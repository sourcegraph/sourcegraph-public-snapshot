{ lib }:
{
  # utility function to add some best-effort flags for emitting static objects instead of dynamic
  mkStatic = drv:
    assert lib.assertMsg (lib.isDerivation drv) "mkStatic expects a derivation, got ${builtins.typeOf drv}";
    assert lib.assertMsg (drv ? "overrideAttrs") "mkStatic expects an overridable derivation";

    let
      auto = builtins.intersectAttrs drv.override.__functionArgs { withStatic = true; static = true; enableStatic = true; enableShared = false; };
      overridden = drv.overrideAttrs (oldAttrs: {
        dontDisableStatic = true;
      } // lib.optionalAttrs (!(oldAttrs.dontAddStaticConfigureFlags or false)) {
        configureFlags = (oldAttrs.configureFlags or [ ]) ++ [ "--disable-shared" "--enable-static" "--enable-shared=false" ];
      });
    in
    if drv.pname == "openssl" then drv.override { static = true; } else overridden.override auto;

  # doesn't actually change anything in practice, just makes otool -L not display nix store paths for libiconv and libxml.
  # they exist in macos dydl cache anyways, so where they point to is irrelevant. worst case, this will let you catch earlier
  # when a library that should be statically linked or that isnt in dydl cache is dynamically linked.
  unNixifyDylibs = { pkgs }: drv:
    assert lib.assertMsg (lib.isDerivation drv) "unNixifyDylibs expects a derivation, got ${builtins.typeOf drv}";
    assert lib.assertMsg (drv ? "overrideAttrs") "unNixifyDylibs expects an overridable derivation";

    drv.overrideAttrs (oldAttrs: lib.optionalAttrs pkgs.hostPlatform.isMacOS {
      nativeBuildInputs = (oldAttrs.nativeBuildInputs or [ ]) ++
        map (drv: drv.__spliced.buildHost or drv)
          (with (pkgs.__splicedPackages or pkgs); [
            findutils
            darwin.cctools
            coreutils
            gnugrep
          ]);

      postFixup = (oldAttrs.postFixup or "") + ''
        for bin in $(find $out/bin -type f); do
          for lib in $(otool -L $bin | tail -n +2 | cut -d' ' -f1 | grep nix); do
            echo "patching dylib from "$lib" to "@rpath/$(basename $lib)""
            install_name_tool -change "$lib" "@rpath/$(basename $lib)" $bin
          done
        done
      '';
    });

  # returns a set of unsuffixed derivations for a native target and suffixed derivations for an optional cross-compile target
  # returned by applying `f` to the passed native & cross-compile package sets.
  xcompilify = { pkgs, pkgsX }: f: lib.foldl # merge all outputs into a single set
    (acc: pkgSet: acc //
      # rename with -${xtarget} if an xtarget package
      (lib.mapAttrs'
        (name: drv:
          assert lib.assertMsg (lib.isDerivation drv) "expected derivation, got ${builtins.typeOf drv}"; {
            name = (name + lib.optionalString
              # cant use drv.stdenv.buildPlatform.system here, as
              # aarch64-darwin.pkgsx86_64Darwin.stdenv.buildPlatform.system == aarch64-darwin.pkgsx86_64Darwin.stdenv.hostPlatform.system
              (pkgs.system != drv.stdenv.hostPlatform.system) "-${drv.stdenv.hostPlatform.system}"
            );
            value = drv;
          })
        pkgSet)
    )
    { }
    # call f with native pkgs (and, if non-null, cross-compile pkgsx)
    (builtins.map (pkgs: f pkgs) ([ pkgs ] ++ lib.optional (pkgsX != null) pkgsX));
}
