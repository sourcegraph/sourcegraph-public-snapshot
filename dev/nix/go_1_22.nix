{ pkgs, fetchurl }:
pkgs.go_1_22.overrideAttrs rec {
  # PLEASE UPDATE THE SHA512 BELOW OR NOTIFY ONE OF THE NIX USERS BEFORE MERGING A CHANGE TO THESE FILES
  version = "1.22.4";
  src = fetchurl {
    url = "https://go.dev/dl/go${version}.src.tar.gz";
    hash = "sha256-/tcgZ45yinyjC6jR3tHKr+J9FgKPqwIyuLqOIgCPt4Q=";
  };
}
