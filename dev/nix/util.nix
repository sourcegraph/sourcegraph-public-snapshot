{ lib }:
{
  # utility function to add some best-effort flags for emitting static objects instead of dynamic
  makeStatic = pkg:
    let
      auto = builtins.intersectAttrs pkg.override.__functionArgs { withStatic = true; static = true; enableStatic = true; enableShared = false; };
      overridden = pkg.overrideAttrs (oldAttrs: {
        dontDisableStatic = true;
      } // lib.optionalAttrs (!(oldAttrs.dontAddStaticConfigureFlags or false)) {
        configureFlags = (oldAttrs.configureFlags or [ ]) ++ [ "--disable-shared" "--enable-static" "--enable-shared=false" ];
      });
    in
    overridden.override auto;

  # doesn't actually change anything in practice, just makes otool -L not display nix store paths for libiconv and libxml.
  # they exist in macos dydl cache anyways, so where they point to is irrelevant. worst case, this will let you catch earlier
  # when a library that should be statically linked or that isnt in dydl cache is dynamically linked.
  unNixifyDylibs = pkgs: drv:
    drv.overrideAttrs (oldAttrs: {
      postFixup = with pkgs; (oldAttrs.postFixup or "") + lib.optionalString pkgs.hostPlatform.isMacOS ''
        for bin in $(${findutils}/bin/find $out/bin -type f); do
          for lib in $(otool -L $bin | ${coreutils}/bin/tail -n +2 | ${coreutils}/bin/cut -d' ' -f1 | ${gnugrep}/bin/grep nix); do
            install_name_tool -change "$lib" "@rpath/$(basename $lib)" $bin
          done
        done
      '';
    });
}
