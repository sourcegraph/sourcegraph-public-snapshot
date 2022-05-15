# Experimental support for developing in nix. Please reach out to @keegan if
# you encounter any issues.
#
# Things it does differently:
#
# - Runs postgres under ~/.sourcegraph with a unix socket. No need to manage a
#   service. Must remember to run "pg_ctl stop" if you want to stop it.
#
# Status: everything works on linux. Go1.17 is currently broken on
# darwin. https://github.com/NixOS/nixpkgs/commit/9675a865c9c3eeec36c06361f7215e109925654c

let
  # Pin a specific version of universal-ctags to the same version as in cmd/symbols/ctags-install-alpine.sh.
  ctags-overlay = (self: super: {
    universal-ctags = super.universal-ctags.overrideAttrs (old: {
      version = "5.9.20220403.0";
      src = super.fetchFromGitHub {
        owner = "universal-ctags";
        repo = "ctags";
        rev = "f95bb3497f53748c2b6afc7f298cff218103ab90";
        sha256 = "sha256-pd89KERQj6K11Nue3YFNO+NLOJGqcMnHkeqtWvMFk38=";
      };
      # disable checks, else we get `make[1]: *** No rule to make target 'optlib/cmake.c'.  Stop.`
      doCheck = false;
      checkFlags = [ ];
    });
  });
  # Pin a specific version of nixpkgs to ensure we get the same packages.
  pkgs = import
    (fetchTarball {
      url =
        "https://github.com/NixOS/nixpkgs/archive/cbe587c735b734405f56803e267820ee1559e6c1.tar.gz";
      sha256 = "0jii8slqbwbvrngf9911z3al1s80v7kk8idma9p9k0d5fm3g4z7h";
    })
    { overlays = [ ctags-overlay ]; };
  # pkgs.universal-ctags installs the binary as "ctags", not "universal-ctags"
  # like zoekt expects.
  universal-ctags = pkgs.writeScriptBin "universal-ctags" ''
    #!${pkgs.stdenv.shell}
    exec ${pkgs.universal-ctags}/bin/ctags "$@"
  '';

in
pkgs.mkShell {
  name = "sourcegraph-dev";

  # The packages in the `buildInputs` list will be added to the PATH in our shell
  nativeBuildInputs = with pkgs; [
    rnix-lsp

    # Our core DB.
    postgresql_13

    # Cache and some store data
    redis

    # Used by symbols and zoekt-git-index to extract symbols from
    # sourcecode.
    universal-ctags

    # Build our backend.
    go_1_18

    # Lots of our tooling and go tests rely on git et al.
    git
    parallel

    # CI lint tools you need locally
    shfmt
    shellcheck
    golangci-lint

    # Web tools. Need node 16.7 so we use unstable. Yarn should also be built
    # against it.
    nodejs-16_x
    (yarn.override { nodejs = nodejs-16_x; })
    nodePackages.typescript

    cargo
    rustc
    rustfmt
    libiconv
  ];

  # Startup postgres
  shellHook = ''
    . ./dev/nix/shell-hook.sh
  '';

  # By explicitly setting this environment variable we avoid starting up
  # universal-ctags via docker.
  CTAGS_COMMAND = "${universal-ctags}/bin/universal-ctags";

  RUST_SRC_PATH = "${pkgs.rust.packages.stable.rustPlatform.rustLibSrc}";
}
