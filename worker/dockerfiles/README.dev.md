Instructions for testing toolchain changes
==============

If you are working on toolchain srclib-LANG, in order to test your changes
with the `src` you can do the following:

- **COPY** your working source code repository to `toolchains/sourcegraph.com/sourcegraph/srclib-LANG`. Unfortunately it's impossible to symlink it, see https://github.com/docker/docker/issues/1676
- make changes in your code
- run `docker build -t sourcegraph/srclib-LANG -f ./Dockerfile.srclib-LANG .`
