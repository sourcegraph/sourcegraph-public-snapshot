# strip-dirs 

[![NPM version](https://badge.fury.io/js/strip-dirs.svg)](https://www.npmjs.org/package/strip-dirs)
[![Build Status](https://travis-ci.org/shinnn/node-strip-dirs.svg?branch=master)](https://travis-ci.org/shinnn/node-strip-dirs)
[![Dependency Status](https://david-dm.org/shinnn/node-strip-dirs.svg)](https://david-dm.org/shinnn/node-strip-dirs)
[![devDependency Status](https://david-dm.org/shinnn/node-strip-dirs/dev-status.svg)](https://david-dm.org/shinnn/node-strip-dirs#info=devDependencies)

Remove leading directory components from a path, like [tar(1)](http://linuxcommand.org/man_pages/tar1.html)'s `--strip-components` option

```javascript
var stripDirs = require('strip-dirs');

stripDirs('foo/bar/baz', 1); //=> 'bar/baz'
stripDirs('foo/bar/baz', 2); //=> 'baz'
stripDirs('foo/bar/baz', 999); //=> 'baz'
```

## Installation

[Install with npm](https://www.npmjs.org/doc/cli/npm-install.html). (Make sure you have installed [Node](http://nodejs.org/))

```
npm install --save strip-dirs
```

## API

```javascript
var stripDirs = require('strip-dirs');
```

### stripDirs(*path*, *count* [, *option*])

*path*: `String` (A relative path)  
*count*: `Number` (0, 1, 2, ...)  
*option*: `Object`  
Return: `String`

It removes directory components from the beginning of the *path* by *count*.

```javascript
var stripDirs = require('strip-dirs');

stripDirs('foo/bar', 1); //=> 'bar'
stripDirs('foo/bar/baz', 2); //=> 'bar'
stripDirs('foo/././/bar/./', 1); //=> 'bar'
stripDirs('foo/bar', 0); //=> 'foo/bar'

stripDirs('/foo/bar', 1) // throw an error because the path is an absolute path
```

#### option.narrow

Type: `Boolean`  
Default: `false`

By default, it keeps the last path component when path components are fewer than the *count*.

If this option is enabled, it throws an error in such case.

```javascript
stripDirs('foo/bar/baz', 9999); //=> 'baz'

stripDirs('foo/bar/baz', 9999, {narrow: true}); // throws an error
```

## CLI

You can use this module as `strip-dirs` command by installing it globally.

```
npm install -g strip-dirs
```

### Usage

```
strip-dirs <string> --count(or -c) <number> [--narrow(or -n)]
```

Or, use with pipe(`|`):

```
echo <string> | strip-dirs --count(or -c) <number> [--narrow(or -n)]
```

### Flags

```
--count,  -c: Number of directories to strip from the path
--narrow, -n: Disallow surplus count of directory level
```

## License

Copyright (c) 2014 [Shinnosuke Watanabe](https://github.com/shinnn)

Licensed under [the MIT LIcense](./LICENSE).
