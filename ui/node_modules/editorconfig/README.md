# EditorConfig JavaScript Core

[![Build Status](https://travis-ci.org/editorconfig/editorconfig-core-js.png?branch=master)](https://travis-ci.org/editorconfig/editorconfig-core-js)

The EditorConfig JavaScript core will provide the same functionality as the
[EditorConfig C Core][] and [EditorConfig Python Core][].


## Installation

You need [node][] to use this package.

To install this package (system-wide):

    $ npm install editorconfig

To install the package system-wide:

    $ npm install -g editorconfig


## Usage

Usage as a Node library:

    $ node
    > var editorconfig = require('./editorconfig');
    undefined
    > editorconfig.parse('/home/zoidberg/humans/anatomy.md');
    { charset: 'utf-8',
        insert_final_newline: 'true',
        end_of_line: 'lf',
        tab_width: '8',
        trim_trailing_whitespace: 'sometimes' }


Usage as a command line tool:

```
$ ./bin/editorconfig

    Usage: editorconfig [OPTIONS] FILEPATH1 [FILEPATH2 FILEPATH3 ...]

    EditorConfig Node.js Core Version 0.11.4-development

    FILEPATH can be a hyphen (-) if you want path(s) to be read from stdin.

    Options:

        -h, --help     output usage information
        -V, --version  output the version number
        -f <path>      Specify conf filename other than ".editorconfig"
        -b <version>   Specify version (used by devs to test compatibility)
```

Example:

    $ ./bin/editorconfig /home/zoidberg/humans/anatomy.md
    charset=utf-8
    insert_final_newline=true
    end_of_line=lf
    tab_width=8
    trim_trailing_whitespace=sometimes


## Development

To install dependencies for this package run this in the package directory:

    $ npm install

Next, run:

    $ npm link

The global editorconfig will now point to the files in your development
repository instead of a globally-installed version from npm. You can now use
editorconfig directly to test your changes.

If you ever update from the central repository and there are errors, it might
be because you are missing some dependencies. If that happens, just run npm
link again to get the latest dependencies.

To test the command line interface:

    $ editorconfig <filepath>


# Testing

[CMake][] must be installed to run the tests.

To run the tests:

    $ npm test

To run the tests with increased verbosity (for debugging test failures):

    $ npm run-script test-verbose

[EditorConfig C Core]: https://github.com/editorconfig/editorconfig-core
[EditorConfig Python Core]: https://github.com/editorconfig/editorconfig-core-py
[node]: http://nodejs.org/
[cmake]: http://www.cmake.org
