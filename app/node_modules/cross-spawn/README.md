# cross-spawn [![Build Status](https://travis-ci.org/IndigoUnited/node-cross-spawn.svg?branch=master)](https://travis-ci.org/IndigoUnited/node-cross-spawn) [![Build status](https://ci.appveyor.com/api/projects/status/ckt1qfvke6mc6pe5/branch/master?svg=true)](https://ci.appveyor.com/project/satazor/node-cross-spawn/branch/master)

A cross platform solution to node's spawn and spawnSync.


## Installation

`$ npm install cross-spawn`

If you're not using the `spawnSync`, you can use [cross-spawn-async](http://github.com/IndigoUnited/node-cross-spawn-async) which doesn't require a build toolchain, see [#18](https://github.com/IndigoUnited/node-cross-spawn/pull/18)


## Why

Node has issues when using spawn on Windows:

- It ignores [PATHEXT](https://github.com/joyent/node/issues/2318)
- It does not support [shebangs](http://pt.wikipedia.org/wiki/Shebang)
- It does not allow you to run `del` or `dir`
- It does not properly escape arguments with spaces or special characters

All these issues are handled correctly by `cross-spawn`.
There are some known modules, such as [win-spawn](https://github.com/ForbesLindesay/win-spawn), that try to solve this but they are either broken or provide faulty escaping of shell arguments.


## Usage

Exactly the same way as node's [`spawn`](https://nodejs.org/api/child_process.html#child_process_child_process_spawn_command_args_options) or [`spawnSync`](https://nodejs.org/api/child_process.html#child_process_child_process_spawnsync_command_args_options), so it's a drop in replacement.

```javascript
var spawn = require('cross-spawn');

// Spawn NPM asynchronously
var process = spawn('npm', ['list', '-g', '-depth' '0'], { stdio: 'inherit' });

// Spawn NPM synchronously
var results = spawn.sync('npm', ['list', '-g', '-depth' '0'], { stdio: 'inherit' });
```


## Tests

`$ npm test`


## License

Released under the [MIT License](http://www.opensource.org/licenses/mit-license.php).
