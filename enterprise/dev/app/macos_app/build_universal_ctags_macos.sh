#!/usr/bin/env bash

cd "${1:-${HOME}/Downloads}" || exit 1

[ -d libyaml ] || git clone https://github.com/yaml/libyaml.git
pushd libyaml || exit 1
git reset --hard
git clean -dfxq
git pull
./bootstrap
for ARCH in arm64e x86_64
do
    arch -${ARCH} ./configure --prefix="${PWD}/build_${ARCH}"
    arch -${ARCH} make clean
    arch -${ARCH} make
    arch -${ARCH} make install
    arch -${ARCH} make clean
done
popd || exit 1

[ -d jansson ] || git clone https://github.com/akheron/jansson.git
pushd jansson || exit 1
git reset --hard
git clean -dfxq
git pull
autoreconf -i
autoupdate
for ARCH in arm64e x86_64
do
    arch -${ARCH} ./configure --prefix="${PWD}/build_${ARCH}"
    arch -${ARCH} make clean
    arch -${ARCH} make
    arch -${ARCH} make install
    arch -${ARCH} make clean
done
popd || exit 1

[ -d pcre2 ] || git clone https://github.com/PCRE2Project/pcre2.git
pushd pcre2 || exit 1
git reset --hard
git clean -dfxq
git pull
./autogen.sh
for ARCH in arm64e x86_64
do
    arch -${ARCH} ./configure --prefix="${PWD}/build_${ARCH}"
    arch -${ARCH} make clean
    arch -${ARCH} make
    arch -${ARCH} make install
    arch -${ARCH} make clean
done
popd || exit 1

[ -d ctags ] || git clone https://github.com/universal-ctags/ctags.git
pushd ctags || exit 1
git reset --hard
git clean -dfxq
git pull
./autogen.sh
# discard dynamic libraries so that the binary will build with static links to avoid requiring those libraries to be available on the target system
# it would be nice to do this in the loop, but it wasn't working for some reason
# it would also be nice to be able to tell the linker to use static libraries, instead,
# but those options also weren't working
find $HOME/Downloads/{libyaml,jansson,pcre2} -name '*.dylib' -exec rm {} +
for ARCH in arm64e x86_64
do
    LDFLAGS=""
    CPPFLAGS=""
    for x in libyaml jansson pcre2
    do
        build="${HOME}/Downloads/${x}/build_${ARCH}"
        [ -d "${build}/lib" ] || [ -d "${build}/include" ] || {
            echo "ERROR: ${build}/lib DNE"
            continue
        }
        [ -n "${LDFLAGS}" ] && LDFLAGS="${LDFLAGS} "
        [ -n "${CPPFLAGS}" ] && CPPFLAGS="${CPPFLAGS} "
        LDFLAGS="${LDFLAGS}-L${build}/lib"
        CPPFLAGS="${CPPFLAGS}-I${build}/include"
    done
    arch -${ARCH} ./configure "CPPFLAGS=${CPPFLAGS}" "LDFLAGS=${LDFLAGS}" --prefix "${PWD}/build_${ARCH}" --program-prefix=universal- --enable-json && \
    arch -${ARCH} make clean && \
    arch -${ARCH} make && \
    rm -rf "${PWD}/build_${ARCH}" && \
    arch -${ARCH} make install && \
    arch -${ARCH} make clean
done
lipo build_{arm64e,x86_64}/bin/universal-ctags -create -output universal-ctags
popd || exit 1

[ -s ctags/universal-ctags ] || {
    echo "failed building" 1>&2
    exit 1
}

echo "${PWD}/ctags/universal-ctags"
