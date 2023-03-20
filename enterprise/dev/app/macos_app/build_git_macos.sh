#!/usr/bin/env bash

symlink_dir() {
  local dir target target_sum file file_sum
  dir="${1}"
  target="${2}"
  pushd "${dir}" || exit 1
  # the git binary will be our reference checksum
  target_sum=$(sha256sum "${target}" | awk '{print $1}')
  while IFS= read -r file; do
    [[ "${file}" == "${target}" ]] && continue
    [ -f "${file}" ] || continue
    file_sum=$(sha256sum "${file}" | awk '{print $1}')
    [[ "${target_sum}" == "${file_sum}" ]] && ln -s -f "${target}" "${file}"
  done < <(ls -1)
  popd || exit 1
}

gettext_version="0.21.1"
git_version="2.39.2"
workdir="${PWD}"

while [ ${#} -gt 0 ]; do
  case "${1}" in
    --gettext-version)
      [ -z "${2}" ] && {
        echo "expected a gettext version" 1>&2
        exit 1
      }
      gettext_version="${2}"
      shift
      ;;
    --git-version)
      [ -z "${2}" ] && {
        echo "expected a git version" 1>&2
        exit 1
      }
      git_version="${2}"
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
      echo "$(basename "${BASH_SOURCE[0]}") [--workdir <directory to work in (defaults to PWD)>] [--git-version <version> (defaults to 2.39.2)] [--gettext-version <version> (defaults to 0.21.1)]" 1>&2
      exit 1
      ;;
  esac
  shift
done

[ -d "${workdir}" ] || {
  [ -d "$(dirname "${workdir}")" ] && mkdir -p "${workdir}"
}

cd "${workdir}" || exit 1

[ -d "gettext-${gettext_version}" ] || {
  echo "downloading gettext ${gettext_version}"
  curl -fsSLO "https://ftp.gnu.org/pub/gnu/gettext/gettext-${gettext_version}.tar.gz"
  tar -xzf gettext-${gettext_version}.tar.gz
}
pushd "gettext-${gettext_version}" || exit 1
for ARCH in arm64e x86_64; do
  # gettext takes a long time to build, so if the build directory exists,
  # assume it's been built and go on to the next one
  [ -d "${PWD}/build_${ARCH}" ] && continue
  arch -${ARCH} ./configure --prefix "${PWD}/build_${ARCH}" &&
    make clean &&
    arch -${ARCH} make &&
    arch -${ARCH} make install
done
# discard dynamic libraries so that the binary will build with static links to avoid requiring those libraries to be available on the target system
# it would be nice to do this in the loop, but it wasn't working for some reason
# it would also be nice to be able to tell the linker to use static libraries, instead,
# but those options also weren't working
find "${PWD}"/build_{arm64e,x86_64} -name '*.dylib' -exec rm {} +
popd || exit 1

[ -d "git-${git_version}" ] || {
  echo "downloading git ${get_version}"
  curl -fsSLO "https://github.com/git/git/archive/refs/tags/v${git_version}.tar.gz"
  tar -xzvf "v${git_version}.tar.gz"
}
pushd "git-${git_version}" || exit 1
for ARCH in arm64e x86_64; do
  [ -d "${PWD}/build_${ARCH}" ] && continue
  arch -${ARCH} make configure &&
    arch -${ARCH} ./configure --prefix "${PWD}/build_${ARCH}" "LDFLAGS=-L${workdir}/gettext-${gettext_version}/build_${ARCH}/lib" "CFLAGS=-I${workdir}/gettext-${gettext_version}/build_${ARCH}/include" &&
    arch -${ARCH} make &&
    arch -${ARCH} make install
done

# cleanup
# the build process creates a bunch of identical binaries
# turning them into symbolic links saves a lot of space
for ARCH in arm64e x86_64; do
  symlink_dir "build_${ARCH}/bin" "git"
  symlink_dir "build_${ARCH}/libexec/git-core" "../../bin/git"
  symlink_dir "build_${ARCH}/libexec/git-core" "../../bin/scalar"
  symlink_dir "build_${ARCH}/libexec/git-core" "git-remote-http"
done

# combine to get universal binaries
# use lipo because we have to build on macOS anyway
while IFS= read -r arm_path; do
  universal_path=${arm_path/build_arm64e/build_universal}
  [ -d "${arm_path}" ] && {
    mkdir -p "${universal_path}"
    continue
  }
  mkdir -p "$(dirname "${universal_path}")"
  lipod=false
  # find executable binaries that have only one arch in them
  [[ $(file -h "${arm_path}" | grep -c Mach-O) -eq 1 ]] && {
    intel_path=${arm_path/build_arm64e/build_x86_64}
    [ -s "${arm_path}" ] && [ -s "${intel_path}" ] && {
      lipo "${arm_path}" "${intel_path}" -create -output "${universal_path}" && lipod=true
    }
  }
  ${lipod} || cp -P "${arm_path}" "${universal_path}"
done < <(find build_arm64e)

# archive the bin and libexec directories, since those are what we need for the macOS app bundle
mkdir "git-${git_version}"
cp -R build_universal/bin "git-${git_version}/bin"
cp -R build_universal/libexec "git-${git_version}/libexec"
tar cvzf "${workdir}/git-universal-${git_version}.tar.gz" "git-${git_version}"
rm -rf "git-${git_version}"

popd || exit 1

echo "${workdir}/git-universal-${git_version}.tar.gz"
