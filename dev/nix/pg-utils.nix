{ stdenv
, fetchurl
, pkg-config
, tzdata
, lib
, path
, substituteAll
, darwin
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

  # needed in order for libpq to be statically linked
  postPatch = ''
    sed -r 's/^(.*all-lib.*[ \t:])[a-z0-9-]+shared\S*/\1/' -i src/Makefile.shlib
  '';

  # for some reason, `make -C src/bin` wasnt being stable for me, but the install variant was,
  # so we essentially do that building in the installing phase instead.
  dontBuild = true;
  doCheck = false;

  installPhase = ''
    make -C src/bin install
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
