"use strict"

var fs = require('fs')
var path = require('path')
var which = require('which')
var findPrefix = require('./find-prefix')

var PATH = getPATHKey()
var SEPARATOR = getPATHSeparator()

/**
 * Get new $PATH setting with additional paths supplied by the npm.
 *
 * @param Object options Config options Object.
 * @param Object options.env Environment to use. Default: process.env
 * @param String options.wd Working directory. Default: process.cwd()
 * @param Function fn callback function.
 */

function getPath(options, fn) {
  var wd = options.cwd = options.cwd || process.cwd()
  var env = options.env = options.env || process.env
  var pathArr = getPathArr(options)

  whichNpm(options, function(err, npmPath) {
    if (err) return fn(err)
    findPrefix(options, function (err, prefixPath) {
      if (!err && prefixPath) {
        // ignore err if cannot find prefix
        pathArr.unshift(path.join(prefixPath, "node_modules", ".bin"))
      }
      // we also unshift the bundled node-gyp-bin folder so that
      // the bundled one will be used for installing things.
      pathArr.unshift(path.join(path.dirname(npmPath), "node-gyp-bin"))
      if (env[PATH]) pathArr.push(env[PATH])
      fn(null, pathArr.join(SEPARATOR))
    })
  })
}

/**
 * Async wrapper around `getPath`.
 */

function getPathAsync(options, fn) {
  // options is optional
  if (options instanceof Function) fn = options, options = {}
  // if no fn, execute as sync
  if (!(fn instanceof Function)) return getPathSync(options)
  options = options || {}
  options.isSync = false
  return getPath(options, fn)
}

/**
 * Sync wrapper around `getPath`.
 */

function getPathSync(options) {
  options = options || {}
  options.isSync = true
  var thePath = undefined
  // sync magic: if sync true, callback is executed sync
  // therefore we can set thePath from inside it before returning
  getPath(options, function(err, foundPath) {
    if (err) throw err
    thePath = foundPath
  })
  return thePath
}

/**
 * Change environment to include npm path adjustments.
 *
 * @param Object options Config options Object.
 * @param Object options.env Environment to use. Default: process.env
 * @param String options.wd Working directory. Default: process.cwd()
 * @param Function fn callback function.
 */

function setPathAsync(options, fn) {
  // options is optional
  if (options instanceof Function) fn = options, options = {}
  // if no fn, execute as sync
  if (!(fn instanceof Function)) return setPathSync(options)

  getPathAsync(options, function(err, newPath) {
    if (err) return fn(err)
    fn(null, options.env[PATH] = newPath)
  })
}

/**
 * Sync version of `setPathAsync`
 */

function setPathSync(options) {
  options = options || {}
  var newPath = getPathSync(options)
  return options.env[PATH] = newPath
}

/**
 * Generate simple parts of the npm path. Basically everything that doesn't
 * depend on potentially async operations.
 *
 * @return Array
 */

function getPathArr(options) {
  var wd = options.cwd
  var pathArr = []
  var p = wd.split(path.sep + "node_modules" + path.sep)
  var acc = path.resolve(p.shift())

  // first add the directory containing the `node` executable currently
  // running, so that any lifecycle script that invoke "node" will execute
  // this same one.
  pathArr.unshift(path.dirname(process.execPath))

  p.forEach(function (pp) {
    pathArr.unshift(path.join(acc, "node_modules", ".bin"))
    acc = path.join(acc, "node_modules", pp)
  })
  pathArr.unshift(path.join(acc, "node_modules", ".bin"))
  return pathArr
}

/**
 * Use callback-style signature but toggle sync execution if `isSync` is true.
 * If options.npm is supplied, this will simply provide npm/bin/npm-cli.
 */

function whichNpm(options, fn) {
  var npmCli = options.npm && path.join(options.npm, 'bin', 'npm-cli.js')

  if (options.isSync) {
    fn(null, fs.realpathSync(
      npmCli || which.sync('npm')
    ))
    return
  }

  if (options.npm) {
    process.nextTick(function() {
      fn(null, npmCli)
    })
    return
  }

  which('npm', function(err, npmPath) {
    if (err) return fn(err)
    fs.realpath(npmPath, fn)
  })
}

/**
 * Get key to use as $PATH in environment
 */

function getPATHKey() {
  var PATH = 'PATH'

  // windows calls it's path "Path" usually, but this is not guaranteed.
  if (process.platform === "win32") {
    PATH = "Path"
    Object.keys(process.env).forEach(function (e) {
      if (e.match(/^PATH$/i)) {
        PATH = e
      }
    })
  }
  return PATH
}

/**
 * Get $PATH separator based on environment
 */

function getPATHSeparator() {
  return process.platform === "win32" ? ";" : ":"
}

module.exports = setPathAsync
module.exports.get = getPathAsync
module.exports.get.sync = getPathSync
module.exports.getSync = getPathSync

module.exports.set = setPathAsync
module.exports.set.sync = setPathSync
module.exports.setSync = setPathSync

module.exports.PATH = PATH
module.exports.SEPARATOR = SEPARATOR
