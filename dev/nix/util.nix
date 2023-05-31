{ lib }:
{
  # utility function to add some best-effort flags for emitting static objects instead of dynamic
  mkStatic = pkg:
    let
      auto = builtins.intersectAttrs pkg.override.__functionArgs { withStatic = true; static = true; enableStatic = true; enableShared = false; };
      overridden = pkg.overrideAttrs (oldAttrs: {
        dontDisableStatic = true;
      } // lib.optionalAttrs (!(oldAttrs.dontAddStaticConfigureFlags or false)) {
        configureFlags = (oldAttrs.configureFlags or [ ]) ++ [ "--disable-shared" "--enable-static" "--enable-shared=false" ];
      });
    in
    if pkg.pname == "openssl" then pkg.override { static = true; } else overridden.override auto;

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

  # removes packages from a list of packages by name.
  # Copied from https://sourcegraph.com/github.com/NixOS/nixpkgs@4d924a6b3376c5e3cae3ba8c971007bf736084c5/-/blob/nixos/lib/utils.nix?L219
  removePackagesByName = packages: packagesToRemove:
    let
      namesToRemove = map lib.getName packagesToRemove;
    in
    lib.filter (x: !(builtins.elem (lib.getName x) namesToRemove)) packages;
}
