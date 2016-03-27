# is-path-global [![Build Status](https://travis-ci.org/kevva/is-path-global.svg?branch=master)](https://travis-ci.org/kevva/is-path-global)

> Check if a path is in PATH


## Install

```
$ npm install --save is-path-global
```


## Usage

```js
var isPathGlobal = require('is-path-global');

isPathGlobal('/bin/sh');
//=> true

isPathGlobal('/home/sirjohndoe');
//=> false
```


## License

MIT © [Kevin Mårtensson](https://github.com/kevva)
