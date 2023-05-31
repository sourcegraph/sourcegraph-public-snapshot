{ pkgs, fetchurl }:
pkgs.nodejs-16_x.overrideAttrs (oldAttrs: {
  # don't override version here, as it won't be in binary cache
  # and building is super expensive
  # version = "16.19.0";

  passthru.pkgs = oldAttrs.passthru.pkgs // {
    pnpm = oldAttrs.passthru.pkgs.pnpm.override rec {
      version = "8.1.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/pnpm/-/pnpm-${version}.tgz";
        sha512 = "sha512-e2H73wTRxmc5fWF/6QJqbuwU6O3NRVZC1G1WFXG8EqfN/+ZBu8XVHJZwPH6Xh0DxbEoZgw8/wy2utgCDwPu4Sg==";
      };
    };
    typescript = oldAttrs.passthru.pkgs.typescript.override rec {
      version = "4.9.5";
      src = fetchurl {
        url = "https://registry.npmjs.org/typescript/-/typescript-${version}.tgz";
        sha512 = "sha512-1FXk9E2Hm+QzZQ7z+McJiHL4NW1F2EzMu9Nq9i3zAaGqibafqYwCVU6WyWAuyQRRzOlxou8xZSyXLEN8oKj24g==";
      };
    };
  };
})
