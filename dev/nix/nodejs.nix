{ pkgs, fetchurl }:
pkgs.nodejs-16_x.overrideAttrs (oldAttrs: {
  # don't override version here, as it won't be in binary cache
  # and building is super expensive
  # version = "16.19.0";

  passthru.pkgs = oldAttrs.passthru.pkgs // {
    pnpm = oldAttrs.passthru.pkgs.pnpm.override rec {
      # PLEASE UPDATE THE SHA512 BELOW OR NOTIFY ONE OF THE NIX USERS BEFORE MERGING A CHANGE TO THESE FILES
      version = "8.3.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/pnpm/-/pnpm-${version}.tgz";
        sha512 = "sha512-wRS8ap/SPxBqbUMzcUNkoA0suLqk9BqMlvi8dM2FRuhwUDgqVGYLc5jQ6Ww3uqVc+84zJvN2GYmTWCubaoWPtQ==";
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
