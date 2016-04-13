var fs = require('fs')
var path = require('path')

module.exports = function (options, cb) {
  return findPrefix(options.cwd, options.isSync, cb)
}

function readdir (p, isSync, fn) {
  if (isSync) {
    try {
      var val = fs.readdirSync(p)
    } catch (err) {
      return fn(err)
    }

    return fn(null, val)
  }

  return fs.readdir(p, fn)
}

// try to find the most reasonable prefix to use
function findPrefix (p, isSync, cb_) {
  function cb (er, p) {
    if (isSync) return cb_(er, p)
    process.nextTick(function () {
      cb_(er, p)
    })
  }

  p = path.resolve(p)
  // if there's no node_modules folder, then
  // walk up until we hopefully find one.
  // if none anywhere, then use cwd.
  var walkedUp = false
  while (path.basename(p) === 'node_modules') {
    p = path.dirname(p)
    walkedUp = true
  }
  if (walkedUp) return cb(null, p)

  findPrefix_(p, p, isSync, cb)
}

function findPrefix_ (p, original, isSync, cb) {
  if (p === '/' ||
      (process.platform === 'win32' && p.match(/^[a-zA-Z]:(\\|\/)?$/))) {
    return cb(null, original)
  }

  readdir(p, isSync, function (er, files) {
    // an error right away is a bad sign.
    // unless the prefix was simply a non
    // existent directory.
    if (er && p === original) {
      if (er.code === 'ENOENT') return cb(null, original)
      return cb(er)
    }

    // walked up too high or something.
    if (er) return cb(null, original)

    if (files.indexOf('node_modules') !== -1 ||
        files.indexOf('package.json') !== -1) {
      return cb(null, p)
    }

    var d = path.dirname(p)
    if (d === p) return cb(null, original)

    return findPrefix_(d, original, isSync, cb)
  })
}
