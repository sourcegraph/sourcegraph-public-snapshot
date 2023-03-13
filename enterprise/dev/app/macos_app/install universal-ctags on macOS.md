# install universal-ctags on macOS

https://github.com/universal-ctags/ctags

It would be nice if we could do it the same way as on Linux: install dependencies, autogen.sh, configure, make, make install.

But when I try that, I get an error about "undefined symbols for architecture arm64"

However, I can compile and install it using homebrew, which does exactly the same steps, but it succeeds;
possibly because it sets up the environemnt differently?
Not sure; I'll just leverage homebrew to compile it.
(I tried poaching from homebrew's environment in order to run the make manually, but didn't succeed.)

The homebrew steps to take are

`brew tap universal-ctags/universal-ctags`

`brew fetch --HEAD universal-ctags/universal-ctags/universal-ctags`

`brew edit universal-ctags/universal-ctags/universal-ctags`

change the line

`system "./configure", "--prefix=#{prefix}", *opts`

to

`system "./configure", "--prefix=#{prefix}", "--program-prefix=universal-", "--enable-json", *opts`

Followed by

`brew install --HEAD universal-ctags/universal-ctags/universal-ctags`

which results in the executable binary `$(brew --prefix)/bin/universal-ctags`

Note that I did try `brew edit universal-ctags`, adding `, "--program-prefix=universal-", "--enable-json"` to the **install** `./configure` line, followed by `brew install --overwrite --build-from-source universal-ctags`, with a `brew uninstall` beforehand thrown in for good measure, but while I could see that it was building from source, it appeared to ignore the extra options. Editing `universal-ctags/universal-ctags/universal-ctags`, as awkward as that appears to be, resulted in a binary that does incorporate the options.

We could, I suppose, fork [the ctags brew formula repo](https://github.com/universal-ctags/homebrew-universal-ctags), add the required `--enable-json` flag (and the "universal-" prefix, also, I suppose, although that may not be necessary), and then we can simply `brew install --HEAD sourcegraph/universal-ctags/universal-ctags`

# UPDATE 2023-02-13

got a manual build to work on macOS:

```bash
cd ${HOME}/Downloads
[ -d ctags ] || git clone https://github.com/universal-ctags/ctags.git
cd ctags
git pull
./autogen.sh
./configure --prefix "${HOME}/Downloads/ctags/build" --program-prefix=universal- --enable-json
make
make install
```

# multi-arch build

Run on macOS - tested on M1

see `build_universal_ctags_macos.sh`

requires `autoconf`

## build libyaml

```bash
cd ${HOME}/Downloads
[ -d libyaml ] || git clone https://github.com/yaml/libyaml.git
cd libyaml
git pull
./bootstrap
for ARCH in arm64e x86_64
do
    arch -${ARCH} ./configure --prefix=${PWD}/build_${ARCH}
    arch -${ARCH} make clean
    arch -${ARCH} make
    arch -${ARCH} make install
    arch -${ARCH} make clean
done
```

## build jansson

requires `autoreconf`

```bash
cd ${HOME}/Downloads
[ -d jansson ] || git clone https://github.com/akheron/jansson.git
cd jansson
git reset --hard
git clean -dfxq
git pull
autoreconf -i
autoupdate
for ARCH in arm64e x86_64
do
    arch -${ARCH} ./configure --prefix=${PWD}/build_${ARCH}
    arch -${ARCH} make clean
    arch -${ARCH} make
    arch -${ARCH} make install
    arch -${ARCH} make clean
done
```

## build PCRE2

```bash
cd ${HOME}/Downloads
[ -d pcre2 ] || git clone https://github.com/PCRE2Project/pcre2.git
cd pcre2
git reset --hard
git clean -dfxq
git pull
./autogen.sh
for ARCH in arm64e x86_64
do
    arch -${ARCH} ./configure --prefix=${PWD}/build_${ARCH}
    arch -${ARCH} make clean
    arch -${ARCH} make
    arch -${ARCH} make install
    arch -${ARCH} make clean
done
```

## build ctags

```bash
cd ${HOME}/Downloads
[ -d ctags ] || git clone https://github.com/universal-ctags/ctags.git
cd ctags
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
```

Examine the binaries using `otool -L build_arm64e/bin/universal-ctags build_x86_64/bin/universal-ctags universal-ctags` if desired. That will show if there are still dynamic libraries beyond the system ones.

The end result of all that is a universal binary in `${PWD}/universal-ctags`. Put that where it needs to be.
