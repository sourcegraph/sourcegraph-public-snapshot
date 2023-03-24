#!/usr/bin/env bash

### has to be run on macOS because there ase C code programs, so there's no other way to generate Mach-O binaries

exedir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

# shellcheck disable=SC1090 disable=SC1091
. "${exedir}/common_functions.sh"

workdir="${PWD}"
libyaml_version=0.2.5
jansson_version=2.14
pcre2_version=10.42
ctags_version=6.0.0

while [ ${#} -gt 0 ]; do
  case "${1}" in
    --libyaml-version)
      [ -z "${2}" ] && {
        echo "expected a libyaml version" 1>&2
        exit 1
      }
      libyaml_version="${2}"
      shift
      ;;
    --jansson-version)
      [ -z "${2}" ] && {
        echo "expected a jansson version" 1>&2
        exit 1
      }
      jansson_version="${2}"
      shift
      ;;
    --pcre2-version)
      [ -z "${2}" ] && {
        echo "expected a pcre2 version" 1>&2
        exit 1
      }
      pcre2_version="${2}"
      shift
      ;;
    --ctags-version)
      [ -z "${2}" ] && {
        echo "expected a ctags version" 1>&2
        exit 1
      }
      pcre2_version="${2}"
      shift
      ;;
    --workdir | --work-dir)
      [ -z "${2}" ] && {
        echo "expected a working directory" 1>&2
        exit 1
      }
      workdir="${2}"
      shift
      ;;
    --help)
      echo "$(basename "${BASH_SOURCE[0]}") [--workdir <directory to work in (defaults to PWD)>] [--libyaml-version <version> (defaults to 0.2.5)] [--jansson-version <version> (defaults to 2.14)] [--pcre2-version <version> (defaults to 10.42)] [--ctags-version <version> (defaults to 6.0.0)]" 1>&2
      exit 1
      ;;
  esac
  shift
done

cd "${workdir}" || exit 1

[ -d "libyaml-${libyaml_version}" ] || {
  curl -fsSL "https://github.com/yaml/libyaml/archive/refs/tags/${libyaml_version}.tar.gz" -o "libyaml-${libyaml_version}.tar.gz"
  tar -xzf "libyaml-${libyaml_version}.tar.gz"
}
pushd "libyaml-${libyaml_version}" || exit 1
./bootstrap
for ARCH in arm64e x86_64; do
  arch -${ARCH} ./configure --prefix="${PWD}/build_${ARCH}"
  arch -${ARCH} make clean
  arch -${ARCH} make
  arch -${ARCH} make install
  arch -${ARCH} make clean
done
popd || exit 1

[ -d "jansson-${jansson_version}" ] || {
  curl -fsSL "https://github.com/akheron/jansson/archive/refs/tags/v2.14.tar.gz" -o "jansson-${jansson_version}.tar.gz"
  tar -xzf "jansson-${jansson_version}.tar.gz"
}
pushd "jansson-${jansson_version}" || exit 1
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

[ -d "pcre2-pcre2-${pcre2_version}" ] || {
  curl -fsSL "https://github.com/PCRE2Project/pcre2/archive/refs/tags/pcre2-${pcre2_version}.tar.gz" -o "pcre2-pcre2-${pcre2_version}.tar.gz"
  tar -xzf "pcre2-pcre2-${pcre2_version}.tar.gz"
}
pushd "pcre2-pcre2-${pcre2_version}" || exit 1
./autogen.sh
for ARCH in arm64e x86_64; do
  arch -${ARCH} ./configure --prefix="${PWD}/build_${ARCH}"
  arch -${ARCH} make clean
  arch -${ARCH} make
  arch -${ARCH} make install
  arch -${ARCH} make clean
done
popd || exit 1

[ -d "ctags-${ctags_version}" ] || {
  curl -fsSL "https://github.com/universal-ctags/ctags/archive/refs/tags/v${ctags_version}.tar.gz" -o "ctags-${ctags_version}.tar.gz"
  tar -xzf "ctags-${ctags_version}.tar.gz"
}
pushd "ctags-${ctags_version}" || exit 1
./autogen.sh
# discard dynamic libraries so that the binary will build with static links to avoid requiring those libraries to be available on the target system
# it would be nice to do this in the loop, but it wasn't working for some reason
# it would also be nice to be able to tell the linker to use static libraries, instead,
# but those options also weren't working
find "${workdir}"/{"libyaml-${libyaml_version}","jansson-${jansson_version}","pcre2-pcre2-${pcre2_version}"} -name '*.dylib' -exec rm {} +
for ARCH in arm64e x86_64; do
  LDFLAGS=""
  CPPFLAGS=""
  for x in "libyaml-${libyaml_version}" "jansson-${jansson_version}" "pcre2-pcre2-${pcre2_version}"; do
    build="${workdir}/${x}/build_${ARCH}"
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
make_fat_binary universal-ctags build_{arm64e,x86_64}/bin/universal-ctags

tar cvzf "${workdir}/universal-ctags-universal-${ctags_version}.tar.gz" "universal-ctags"

popd || exit 1

[ -s "${workdir}/universal-ctags-universal-${ctags_version}.tar.gz" ] || {
  echo "failed building" 1>&2
  exit 1
}

echo "${workdir}/universal-ctags-universal-${ctags_version}.tar.gz"
