{ pkgs, fetchurl }:
pkgs.nodejs_20.overrideAttrs (oldAttrs: {
  # don't override version here, as it won't be in binary cache
  # and building is super expensive
  # version = "16.19.0";

  passthru.pkgs = oldAttrs.passthru.pkgs // {
    pnpm = oldAttrs.passthru.pkgs.pnpm.override rec {
      # PLEASE UPDATE THE SHA512 BELOW OR NOTIFY ONE OF THE NIX USERS BEFORE MERGING A CHANGE TO THESE FILES
      version = "8.9.2";
      src = fetchurl {
        url = "https://registry.npmjs.org/pnpm/-/pnpm-${version}.tgz";
        sha512 = "sha512-udNf6RsqWFTa3EMDSj57LmdfpLVuIOjgnvB4+lU8GPiu1EBR57Nui43UNfl+sMRMT/O0T8fG+n0h4frBe75mHg==";
      };
    };
    # fetching typescript 5.2.2 from npm registry results in an archive with a missing package-lock.json, which nix
    # refuses to build without. We can generate this lock file with `npm install --package-lock-only` but at the time I didn't
    # want to vendor the file. Fortunately, nixpkgs-unstable has typescript at 5.2.2
    typescript = assert pkgs.nodePackages.typescript.version == "5.2.2"; pkgs.nodePackages.typescript;
  };
})
