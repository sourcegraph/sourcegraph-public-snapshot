{ pkgs
, lib
, clang13Stdenv
, fetchzip
, fetchFromGitHub
, cmake
, http-parser
, libcxx
, libcxxabi
, libiconv
, libressl
, libssh2
, patchelf
, pcre
, pkg-config
, zlib
, darwin
, hostPlatform
}:
let
  # utility function to add some best-effort flags for emitting static objects instead of dynamic
  makeStatic = pkg: pkg.overrideAttrs (oldAttrs: {
    configureFlags = (oldAttrs.configureFlags or [ ]) ++ [ "--without-shared" "--disable-shared" "--enable-static" ];
  });
  http-parser-static = ((makeStatic http-parser).overrideAttrs (oldAttrs: {
    # http-parser makefile is a bit incomplete, so fill in the gaps here
    # to move the static object and header files to the right location
    # https://github.com/nodejs/http-parser/issues/310
    buildFlags = [ "package" ];
    installTargets = "package";
    postInstall = ''
      install -D libhttp_parser.a $out/lib/libhttp_parser.a
      install -D  http_parser.h $out/include/http_parser.h
      ls -la $out/lib $out/include
    '';
  }));
  libiconv-static = (libiconv.override { enableStatic = true; enableShared = false; });
  libressl-static = (libressl.override { buildShared = false; });
  libssh2-libressl = ((makeStatic libssh2).override { openssl = libressl-static; });
  pcre-static = (makeStatic pcre).dev;
in
clang13Stdenv.mkDerivation rec {
  name = "p4-fusion";
  version = "v1.12";

  srcs = [
    (fetchFromGitHub {
      inherit name;
      owner = "salesforce";
      repo = "p4-fusion";
      rev = "3ee482466464c18e6a635ff4f09cd75a2e1bfe0f";
      hash = "sha256-rUXuBoXuOUanWxutd7dNgjn2vLFvHQ0IgCIn9vG5dgs=";
    })
    (
      if hostPlatform.isMacOS then
        fetchzip
          {
            name = "helix-core-api";
            url = "https://cdist2.perforce.com/perforce/r21.1/bin.macosx1015x86_64/p4api.tgz";
            hash = "sha256-KctrQcglwEHav+9m7ipw0fX4dds079/TVFlKONYlQeQ=";
          }
      else if hostPlatform.isLinux then
        fetchzip
          {
            name = "helix-core-api";
            url = "https://cdist2.perforce.com/perforce/r21.1/bin.linux26x86_64/p4api-glibc2.3-openssl1.0.2.tgz";
            hash = "sha256-gzmL0EMAC3vhSZXJjqRDadJqtLiMGX6lCK2DCFNAnus=";
          }
      else throw "unsupported platform ${clang13Stdenv.targetPlatform.parsed.kernel.name}"
    )
  ];

  sourceRoot = name;

  nativeBuildInputs = [
    patchelf
    pkg-config
    cmake
  ];

  buildInputs = [
    libcxx
    libcxxabi
    zlib.static
    zlib.dev
    http-parser-static
    pcre-static
    # openssl 1.0.2 is considered insecure, and p4-fusion won't compile against openssl >1.1 or higher
    # so we're using libressl instead
    libressl-static
    libssh2-libressl
  ] ++ lib.optional hostPlatform.isMacOS [
    # iconv is bundled with glibc and apparently only needed for osx
    # https://sourcegraph.com/github.com/salesforce/p4-fusion@3ee482466464c18e6a635ff4f09cd75a2e1bfe0f/-/blob/vendor/libgit2/README.md?L178:3
    libiconv-static
    darwin.apple_sdk.frameworks.CFNetwork
    darwin.apple_sdk.frameworks.Cocoa
  ];

  # because the world of openssl is an API versioning nightmare, we have to do some patching/stubbing here
  postUnpack =
    # helix-core-api references this from openssl, but its not defined in libressl (and deprecated in openssl anyways)
    ''
      echo 'extern "C" void SSL_COMP_free_compression_methods(void) { }' > $sourceRoot/p4-fusion/libressl.cc
    '' + lib.optionalString
      hostPlatform.isLinux
      # HMAC_CTX_cleanup was removed in libressl 3.5, so we update the #if guard in the code here to
      # include the stub for libressl versions >=3.5, but only on Linux apparently?
      # https://sourcegraph.com/github.com/salesforce/p4-fusion@3ee482466464c18e6a635ff4f09cd75a2e1bfe0f/-/blob/vendor/libgit2/deps/ntlmclient/crypt_openssl.c?L47
      # https://github.com/libgit2/libgit2/pull/6157#issuecomment-1039111648
      ''
        substituteInPlace $sourceRoot/vendor/libgit2/deps/ntlmclient/crypt_openssl.c \
          --replace \
            "#if (OPENSSL_VERSION_NUMBER >= 0x10100000L && !LIBRESSL_VERSION_NUMBER) || defined(CRYPT_OPENSSL_DYNAMIC)" \
            "#if ((OPENSSL_VERSION_NUMBER >= 0x10100000L && !defined(LIBRESSL_VERSION_NUMBER)) || LIBRESSL_VERSION_NUMBER >= 0x3050000f) || defined(CRYPT_OPENSSL_DYNAMIC)"
      '';

  # copy helix-core-api stuff into the expected directories
  preBuild = let dir = if hostPlatform.isMacOS then "mac" else "linux"; in
    ''
      mkdir -p $NIX_BUILD_TOP/$sourceRoot/vendor/helix-core-api/${dir}
      cp -R $NIX_BUILD_TOP/helix-core-api/* $NIX_BUILD_TOP/$sourceRoot/vendor/helix-core-api/${dir}
    '';

  cmakeFlags = [
    # we want to statically link
    "-DBUILD_SHARED_LIBS=OFF"
    # no harm to enable probably
    "-DUSE_SSH=ON"
    # enable for libgit2
    "-DTHREADSAFE=ON"
    # salesforce don't link against GSSAPI in CI, so I won't either
    "-DUSE_GSSAPI=OFF"
    # prefer nix-provided http-parser instead of bundled
    "-DUSE_HTTP_PARSER=system"
  ];

  postInstall = ''
    mkdir -p "$out/bin"
    cp p4-fusion/p4-fusion "$out/bin/p4-fusion"
  '';
}
