Libsass
=======

by Aaron Leung ([@akhleung]) and Hampton Catlin ([@hcatlin])

[![Linux CI](https://travis-ci.org/sass/libsass.png?branch=master)](https://travis-ci.org/sass/libsass)
[![Windows CI](https://ci.appveyor.com/api/projects/status/github/sass/libsass?svg=true)](https://ci.appveyor.com/project/sass/libsass/branch/master)
[![Bountysource](https://www.bountysource.com/badge/tracker?tracker_id=283068)](https://www.bountysource.com/trackers/283068-libsass?utm_source=283068&utm_medium=shield&utm_campaign=TRACKER_BADGE)
[![Coverage Status](https://img.shields.io/coveralls/sass/libsass.svg)](https://coveralls.io/r/sass/libsass?branch=feature%2Ftest-travis-ci-3)
[![Join us](https://libsass-slack.herokuapp.com/badge.svg)](https://libsass-slack.herokuapp.com/)

https://github.com/sass/libsass

Libsass is just a library, but if you want to RUN libsass,
then go to https://github.com/sass/sassc or
https://github.com/sass/ruby-libsass or
[find your local implementer](https://github.com/sass/libsass/wiki/Implementations).

LibSass requires GCC 4.6+ or Clang/LLVM. If your OS is older, this version may not compile.

On Windows, you need MinGW with GCC 4.6+ or VS 2013 Update 4+. It is also possible to build LibSass with Clang/LLVM on Windows.

About
-----

Libsass is a C/C++ port of the Sass CSS precompiler. The original version was written in Ruby, but this version is meant for efficiency and portability.

This library strives to be light, simple, and easy to build and integrate with a variety of platforms and languages.

Developing
----------

As you may have noticed, the libsass repo itself has
no executables and no tests. Oh noes! How can you develop???

Well, luckily, SassC is the official binary wrapper for
libsass and is *always* kept in sync. SassC uses a git submodule
to include libsass. When developing libsass, its best to actually
check out SassC and develop in that directory with the SassC spec
and tests there.

We even run Travis tests for SassC!

Tests
-------

Since libsass is a pure library, tests are run through the [SassSpec](https://github.com/sass/sass-spec) project using the [SassC](http://github.com/sass/sassc) driver.

To run tests against libsass while developing, you can run `./script/spec`. This will clone SassC and Sass-Spec under the project folder and then run the Sass-Spec test suite. You may want to update the clones to ensure you have the latest version.

Library Usage
-------------

While libsass is a library primarily written in C++, it provides a simple
C interface which should be used by most implementers. Libsass does not do
much on its own without an implementer. This can be a command line tool, like
[sassc](https://github.com/sass/sassc) or a [binding](https://github.com/sass/libsass/wiki/Implementations)
to your favorite programing language.

If you want to build or interface with libsass, we recommend to check out the
wiki pages about [building libsass](https://github.com/sass/libsass/wiki/Building-Libsass) and
the [C-API documentation](https://github.com/sass/libsass/wiki/API-Documentation).

About Sass
----------

Sass is a CSS pre-processor language to add on exciting, new,
awesome features to CSS. Sass was the first language of its kind
and by far the most mature and up to date codebase.

Sass was originally created by the co-creator of this library,
Hampton Catlin ([@hcatlin]). The extension and continuing evolution
of the language has all been the result of years of work by Natalie
Weizenbaum ([@nex3]) and Chris Eppstein ([@chriseppstein]).

For more information about Sass itself, please visit http://sass-lang.com

Contribution Agreement
----------------------

Any contribution to the project are seen as copyright assigned to Hampton Catlin, a
human on the planet earth. Your contribution warrants that you have the right to
assign copyright on your work. The intention here is to ensure that the project
remains totally free (liberal, like).

Our MIT license is designed to be as simple, and liberal as possible.

[@hcatlin]: https://github.com/hcatlin
[@akhleung]: https://github.com/akhleung
[@chriseppstein]: https://github.com/chriseppstein
[@nex3]: https://github.com/nex3

sass2scss was originally written by [Marcel Greter](@mgreter)
and he happily agreed to have it merged into the project.

[sass_interface.h]: sass_interface.h
