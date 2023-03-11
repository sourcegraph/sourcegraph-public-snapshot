#!/usr/bin/env bash

### has to be run on macOS because there ase C code programs, so there's no other way to generate Mach-O binaries

DIR="${1}"
[[ $# -lt 1 ]] || [[ -z "${1}" ]] && {
  DIR="${PWD}"
}

cd "${DIR}" || exit 1

[ -d libyaml-0.2.5 ] || {
  curl -fsSL "https://github.com/yaml/libyaml/archive/refs/tags/0.2.5.tar.gz" -o "libyaml-0.2.5.tar.gz"
  tar -xzf "libyaml-0.2.5.tar.gz"
}
pushd libyaml-0.2.5 || exit 1
./bootstrap
for ARCH in arm64e x86_64; do
  arch -${ARCH} ./configure --prefix="${PWD}/build_${ARCH}"
  arch -${ARCH} make clean
  arch -${ARCH} make
  arch -${ARCH} make install
  arch -${ARCH} make clean
done
popd || exit 1

[ -d jansson-2.14 ] || {
  curl -fsSL "https://github.com/akheron/jansson/archive/refs/tags/v2.14.tar.gz" -o "jansson-2.14.tar.gz"
  tar -xzf "jansson-2.14.tar.gz"
}
pushd jansson-2.14 || exit 1
autoreconf -i
autoupdate
for ARCH in arm64e x86_64; do
  arch -${ARCH} ./configure --prefix="${PWD}/build_${ARCH}"
  arch -${ARCH} make clean
  arch -${ARCH} make
  arch -${ARCH} make install
  arch -${ARCH} make clean
done
popd || exit 1

[ -d pcre2-pcre2-10.42 ] || {
  curl -fsSL "https://github.com/PCRE2Project/pcre2/archive/refs/tags/pcre2-10.42.tar.gz" -o "pcre2-pcre2-10.42.tar.gz"
  tar -xzf "pcre2-pcre2-10.42.tar.gz"
}
pushd pcre2-pcre2-10.42 || exit 1
./autogen.sh
for ARCH in arm64e x86_64; do
  arch -${ARCH} ./configure --prefix="${PWD}/build_${ARCH}"
  arch -${ARCH} make clean
  arch -${ARCH} make
  arch -${ARCH} make install
  arch -${ARCH} make clean
done
popd || exit 1

[ -d ctags-6.0.0/ ] || {
  curl -fsSL "https://github.com/universal-ctags/ctags/archive/refs/tags/v6.0.0.tar.gz" -o "ctags-6.0.0.tar.gz"
  tar -xzf "ctags-6.0.0.tar.gz"
}
pushd ctags-6.0.0/ || exit 1
./autogen.sh
# discard dynamic libraries so that the binary will build with static links to avoid requiring those libraries to be available on the target system
# it would be nice to do this in the loop, but it wasn't working for some reason
# it would also be nice to be able to tell the linker to use static libraries, instead,
# but those options also weren't working
find "${DIR}"/{libyaml-0.2.5,jansson-2.14,pcre2-pcre2-10.42} -name '*.dylib' -exec rm {} +
for ARCH in arm64e x86_64; do
  LDFLAGS=""
  CPPFLAGS=""
  for x in libyaml-0.2.5 jansson-2.14 pcre2-pcre2-10.42; do
    build="${DIR}/${x}/build_${ARCH}"
    [ -d "${build}/lib" ] || [ -d "${build}/include" ] || {
      echo "ERROR: ${build}/lib DNE"
      continue
    }
    [ -n "${LDFLAGS}" ] && LDFLAGS="${LDFLAGS} "
    [ -n "${CPPFLAGS}" ] && CPPFLAGS="${CPPFLAGS} "
    LDFLAGS="${LDFLAGS}-L${build}/lib"
    CPPFLAGS="${CPPFLAGS}-I${build}/include"
  done
  arch -${ARCH} ./configure "CPPFLAGS=${CPPFLAGS}" "LDFLAGS=${LDFLAGS}" --prefix "${PWD}/build_${ARCH}" --program-prefix=universal- --enable-json &&
    arch -${ARCH} make clean &&
    arch -${ARCH} make &&
    rm -rf "${PWD}/build_${ARCH}" &&
    arch -${ARCH} make install &&
    arch -${ARCH} make clean
done
lipo build_{arm64e,x86_64}/bin/universal-ctags -create -output universal-ctags
popd || exit 1

[ -s ctags-6.0.0/universal-ctags ] || {
  echo "failed building" 1>&2
  exit 1
}

echo "${PWD}/ctags-6.0.0/universal-ctags"
