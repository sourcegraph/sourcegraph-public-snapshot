# flow-bin [![Build Status](https://travis-ci.org/flowtype/flow-bin.svg?branch=master)](https://travis-ci.org/flowtype/flow-bin)

> Binary wrapper for [Flow](http://flowtype.org) - A static type checker for JavaScript

Only OS X and Linux (64-bit) binaries are currently [provided](http://flowtype.org/docs/getting-started.html#_).


## CLI

```
$ npm install --global flow-bin
```

```
$ flow --help
```


## API

```
$ npm install --save flow-bin
```

```js
const execFile = require('child_process').execFile;
const flow = require('flow-bin');

execFile(flow, ['check'], (err, stdout) => {
	console.log(stdout);
});
```


## License

flow-bin is BSD-licensed. We also provide an additional patent grant.
