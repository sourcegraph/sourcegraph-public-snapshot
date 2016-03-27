"use strict"

var which = require('which')
var npmPath = require('npm-path')

module.exports = function(cmd, options, fn) {
  if (options instanceof Function) fn = options, options = null
  options = options || {}
  options.cwd = options.cwd || process.cwd()
  options.env = options.env || process.env
  npmPath.get(options, function(err, newPath) {
    if (err) return fn(err)
    var oldPath = process.env[npmPath.PATH]
    process.env[npmPath.PATH] = newPath
    which(cmd, function(err, result) {
      process.env[npmPath.PATH] = oldPath
      fn(err, result)
    })
  })
}

module.exports.sync = function(cmd, options) {
  options = options || {}
  options.cwd = options.cwd || process.cwd()
  options.env = options.env || process.env
  var err = null
  try {
    var oldPath = process.env[npmPath.PATH]
    var newPath = npmPath.getSync(options)
    process.env[npmPath.PATH] = newPath
    var result = which.sync(cmd)
    return result
  } catch(e) {
    err = e
  } finally {
    process.env[npmPath.PATH] = oldPath
    if (err) throw err
  }
  return result
}
