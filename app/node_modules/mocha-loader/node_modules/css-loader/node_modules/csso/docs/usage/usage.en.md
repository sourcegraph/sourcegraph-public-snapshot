## In the browser

Use a [browser vesion of CSSO](http://css.github.com/csso/csso.html) to minimize your code.

**CSSO is not guaranteed to work in browsers. Preferred way to use this tool is to run it from the command line or via npm modules.**

## From the command line

Run `bin/csso` (when installed from git), [NodeJS](http://nodejs.org) 0.8.x is required.

Run `csso` (when installed from npm).

Usage:

    csso
        shows usage information
    csso <filename>
        minimizes the CSS in <filename> and outputs the result to stdout
    csso <in_filename> <out_filename>
    csso -i <in_filename> -o <out_filename>
    csso --input <in_filename> --output <out_filename>
        minimizes the CSS in <in_filename> and outputs the result to <out_filename>
    csso -off
    csso --restructure-off
        turns structure minimization off
    csso -h
    csso --help
        shows usage information
    csso -v
    csso --version
        shows the version number

Example:

    $ echo ".test { color: red; color: green }" > test.css
    $ csso test.css
    .test{color:green}

## As an npm module

Sample (`test.js`):
```js
    var csso = require('csso'),
        css = '.test, .test { color: rgb(255, 255, 255) }';

    console.log(csso.justDoIt(css));
```
Output (`> node test.js`):
```css
    .test{color:#fff}
```
Use `csso.justDoIt(css, true)` to turn structure minimization off.
