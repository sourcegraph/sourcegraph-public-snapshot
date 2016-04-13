# npm-which

### Locate a program or locally installed node module executable

Use `npm-which` to locate executables which may be installed in the
local 'node_modules/.bin', or in a parent 'node_modules/.bin' directory.

`npm-which` runs in the context of an npm lifecycle script with its npm-modified PATH.

i.e. if you install a module that has an executable script using npm install, that module's executable will be picked up by `npm-which` from anywhere in the ./node_modules tree.

## Installation

```bash
> npm install -g npm-which
```

## Usage

### Command Line

```bash
> npm-which tape
/Users/timoxley/Projects/npm-which/node_modules/.bin/tape
```

This is the equivalent of running an npm script with the body: `which tape`.

### Example

```bash
# unless something is installed in a node_modules
# npm-which and which(1) will have the same output:

> which tape
/usr/local/bin/tape

> npm-which tape
/usr/local/bin/tape

# install tape local to current dir
# tape includes an executable 'tape'
> npm install tape
> ./node_modules/.bin/tape && echo 'found'
found

# vanilla which(1) still finds global tape
> which tape
/usr/local/bin/tape

# npm-which finds locally installed tape :)
> npm-which tape
/Users/timoxley/Projects/npm-which/node_modules/.bin/tape
```

### Programmatic

#### Asynchronous

```js
var which = require('npm-which')
which('tape', function(err, pathToTape) {
  if (err) return console.error(err.message)
  console.log(pathToTape) // /Users/.../node_modules/.bin/tape
})
```

#### Synchronous

```js
var which = require('npm-which')
var pathToTape = which.sync('tape')
console.log(pathToTape) // /Users/.../node_modules/.bin/tape
```

#### Options

Both async and sync versions take an optional options object:

* Set `options.env` if you wish to use something other than `process.env` (the default)
* Set `options.cwd` if you wish to use something other than `process.cwd()` (the default)

```js
which('tape', {cwd: '/some/other/path'}, function() {
  // ...
})
```

## Why

#### npm is slow to boot

* Shelling out to `npm bin` is very slow; it has to wait for all of npm to boot up – this often takes longer than the actual script you want to execute!

#### Hard-coding paths to modules is very fragile

* You can't rely on './node_modules' actually containing your module! The module may exist much higher in the directory hierarchy.
* `npm bin` returns the location of the `./node_modules/.bin` directory, but it does not take into account being called within the context of another module, also, npm slow.
* If the module does exist in a parent directory, then './node_modules/.bin' will be missing your module's executable.

## License

MIT
