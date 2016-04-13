# decompress-unzip [![Build Status](http://img.shields.io/travis/kevva/decompress-unzip.svg?style=flat)](https://travis-ci.org/kevva/decompress-unzip)

> zip decompress plugin

## Install

```sh
$ npm install --save decompress-unzip
```

## Usage

```js
var Decompress = require('decompress');
var zip = require('decompress-unzip');

var decompress = new Decompress()
	.src('foo.zip')
	.dest('dest')
	.use(zip({strip: 1}));

decompress.run(function (err, files) {
	if (err) {
		throw err;
	}

	console.log('Files extracted successfully!'); 
});
```

You can also use this plugin with [gulp](http://gulpjs.com):

```js
var gulp = require('gulp');
var zip = require('decompress-unzip');

gulp.task('default', function () {
	return gulp.src('foo.zip')
		.pipe(zip({strip: 1}))
		.pipe(gulp.dest('dest'));
});
```

## Options

### strip

Type: `Number`  
Default: `0`

Equivalent to `--strip-components` for tar.

## License

MIT © [Kevin Mårtensson](https://github.com/kevva)
