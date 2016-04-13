# npm-installed [![Build Status](http://img.shields.io/travis/kevva/npm-installed.svg?style=flat)](https://travis-ci.org/kevva/npm-installed)

> Find programs installed by npm

## Install

```sh
$ npm install --save npm-installed
```

## Usage

```js
var npmInstalled = require('npm-installed');

npmInstalled('imagemin', function (err, path) {
	if (err) {
		throw err;
	}

	console.log(path);
	//=> /home/sirjohndoe/.npm-packages/bin/imagemin
});

npmInstalled.sync('imagemin');
//=> /home/sirjohndoe/.npm-packages/bin/imagemin
```

## License

MIT © [Kevin Mårtensson](https://github.com/kevva)
