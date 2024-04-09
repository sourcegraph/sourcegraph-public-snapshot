{ stdenv
, fetchurl
, pkg-config
, tzdata
, lib
, path
, substituteAll
, darwin
, gnugrep
, patchelf
}:
stdenv.mkDerivation rec {
  name = "postgresql";
  version = "12.18";

  src = fetchurl {
    url = "mirror://postgresql/source/v${version}/postgresql-${version}.tar.bz2";
    hash = "sha256-T5kZcl2UHOmGjgf+HtHTqGdIWZtIM4ZUdYOSi3TDkYo=";
  };

  hardeningEnable = [ ];

  outputs = [ "out" ];

  LC_ALL = "C";

  nativeBuildInputs = [
    pkg-config
  ];

  # needed in order for libpq to be statically linked.
  # results in the following diff:
  # 277c277
  # < .PHONY: all-lib all-static-lib all-shared-lib
  # ---
  # > .PHONY: all-lib all-static-lib
  # 279c279
  # < all-lib: all-shared-lib
  # ---
  # > all-lib:
  # 445,446c445,446
  # < .PHONY: install-lib install-lib-static install-lib-shared installdirs-lib
  # < install-lib: install-lib-shared
  # ---
  # > .PHONY: install-lib install-lib-static  installdirs-lib
  # > install-lib:
  postPatch = ''
    sed -r 's/^(.*all-lib.*[ \t:])[a-z0-9-]+shared\S*/\1/' -i src/Makefile.shlib
  '';

  # for some reason, `make -C src/bin` wasnt being stable for me, but the install variant was,
  # so we essentially do that building in the installing phase instead.
  dontBuild = true;

  installPhase = ''
    make -C src/bin install
  '';

  # guard against dynamically linking against anything besides libc
  doInstallCheck = true;
  installCheckPhase = ''
    ${patchelf}/bin/patchelf --print-needed $out/bin/pg_dump | \
      ${gnugrep}/bin/grep -v 'libc.so.6' && echo 'unexpected dynamic library dependency found, only libc is expected' && exit 1 \
      || exit 0
  '';

  # update linker to not point to the nix one, but to one that will work on
  # most other distros such as Ubuntu
  postFixup = lib.optionalString stdenv.isLinux ''
    for bin in $out/bin/*; do
      ${patchelf}/bin/patchelf --set-interpreter /lib64/ld-linux-x86-64.so.2 "$bin"
    done
  '';

  # Want the minimal amount of things involved, so we don't have to deal with
  # statically linking them all. We'll discover which we need as time goes on : )
  configureFlags = [
    "USE_DEV_URANDOM=1"
    "--without-openssl"
    "--without-libxml"
    "--without-icu"
    "--sysconfdir=/etc"
    "--prefix=$(out)/"
    "--with-system-tzdata=${tzdata}/share/zoneinfo"
    "--disable-strong-random"
    "--without-readline"
    "--without-zlib"
  ];

  # some patches taken from https://sourcegraph.com/github.com/NixOS/nixpkgs/-/blob/pkgs/servers/sql/postgresql/generic.nix,
  # only removed those obviously not needed, but haven't vetted the rest.
  patches = [
    "${path}/pkgs/servers/sql/postgresql/patches/disable-resolve_symlinks.patch"
    "${path}/pkgs/servers/sql/postgresql/patches/less-is-more.patch"
    "${path}/pkgs/servers/sql/postgresql/patches/hardcode-pgxs-path.patch"
    "${path}/pkgs/servers/sql/postgresql/patches/specify_pkglibdir_at_runtime.patch"
    "${path}/pkgs/servers/sql/postgresql/patches/findstring.patch"

    (substituteAll {
      src = "${path}/pkgs/servers/sql/postgresql/locale-binary-path.patch";
      locale = "${if stdenv.isDarwin then darwin.adv_cmds else lib.getBin stdenv.cc.libc}/bin/locale";
    })

  ] ++ lib.optionals stdenv.isLinux [
    "${path}/pkgs/servers/sql/postgresql/patches/socketdir-in-run.patch"
  ];

  disallowedReferences = [ stdenv.cc ];
}
