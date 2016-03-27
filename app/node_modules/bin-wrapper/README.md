# bin-wrapper [![Build Status](http://img.shields.io/travis/kevva/bin-wrapper.svg?style=flat)](https://travis-ci.org/kevva/bin-wrapper)

> Binary wrapper that makes your programs seamlessly available as local dependencies

## Install

```sh
$ npm install --save bin-wrapper
```

## Usage

```js
var BinWrapper = require('bin-wrapper');

var base = 'https://github.com/imagemin/gifsicle-bin/raw/master/vendor';
var bin = new BinWrapper()
	.src(base + '/osx/gifsicle', 'darwin')
	.src(base + '/linux/x64/gifsicle', 'linux', 'x64')
	.src(base + '/win/x64/gifsicle.exe', 'win32', 'x64')
	.dest(path.join('vendor'))
	.use(process.platform === 'win32' ? 'gifsicle.exe' : 'gifsicle')
	.version('>=1.71');

bin.run(['--version'], function (err) {
	if (err) {
		throw err;
	}

	console.log('gifsicle is working');
});
```

Get the path to your binary with `bin.path()`:

```js
console.log(bin.path()); // => path/to/vendor/gifsicle
```

## API

### new BinWrapper(opts)

Creates a new `BinWrapper` instance. The available options are:

* `global`: Whether to check for global binaries or not. Defaults to `false`.
* `progress`: Show a progress bar when downloading files. Defaults to `true`.
* `skip`: Whether to skip checking if the binary works or not. Defaults to `false`.
* `strip`: Strip a number of leading paths from file names on extraction. Defaults to `1`.

### .src(url, os, arch)

Adds a source to download.

#### url

Type: `String`

Accepts a URL pointing to a file to download.

#### os

Type: `String`

Tie the source to a specific OS.

#### arch

Type: `String`

Tie the source to a specific arch.

### .dest(dest)

Type: `String`

Accepts a path which the files will be downloaded to.

### .use(bin)

Type: `String`

Define which file to use as the binary.

### .path()

Get the full path to your binary.

### .version(range)

Type: `String`

Define a [semver range](https://github.com/isaacs/node-semver#ranges) to check 
the binary against.

### .run(cmd, cb)

Runs the search for the binary. If no binary is found it will download the file 
using the URL provided in `.src()`.

#### cmd

Type: `Array`

Command to run the binary with. If it exits with code `0` it means that the 
binary is working.

#### cb(err)

Type: `Function`

Returns nothing but a possible error.

## License

MIT © [Kevin Mårtensson](http://kevinmartensson.com)
