lintspaces-cli
==============

Simple as pie CLI for the node-lintspaces module. Supports all the usual
lintspaces args that the Grunt, Gulp and vanilla Node.js module support.

## Install
```
$ npm install -g lintspaces-cli
```


## Help Output
```
eshortiss@Evans-MacBook-Pro:~/lintspaces --help

  Usage: lintspaces [options]

  Options:

    -h, --help                      output usage information
    -V, --version                   output the version number
    -n, --newline                   Require newline at end of file.
    -g, --guessindentation          Tries to guess the indention of a line depending on previous lines
    -b, --skiptrailingonblank       Skip blank lines in trailingspaces check.
    -it, --trailingspacestoignores  Ignore trailing spaces in ignores
    -l, --maxnewlines <n>           Specify max number of newlines between blocks.
    -t, --trailingspaces            Tests for useless whitespaces (trailing whitespaces) at each lineending of all files.
    -d, --indentation <s>           Check indentation is "tabs" or "spaces".
    -s, --spaces <n>                Used in conjunction with -d to set number of spaces.
    -i, --ignores <items>           Comma separated list of ignores.
    -e, --editorconfig <s>          Use editorconfig specified at this file path for settings.
```

## Example Commands

Check all JavaScript files in directory for trailing spaces and newline at the
end of file:

```
lintspaces -n -t ./*.js
```

Check that 2 spaces are used as indent:

```
lintspaces -nt -s 2 -d spaces ./*.js
```

## Changelog

* 0.1.1 - Support for Node.js <=4.0.0 (thank you @gurdiga)

* 0.1.0 - Initial stable release

* < 0.1.0 - Dark ages...

## Contributors
* [Vlad Gurdiga](https://github.com/gurdiga)
