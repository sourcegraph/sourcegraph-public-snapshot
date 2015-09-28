#!/usr/bin/env node

var concat = require('concat-stream')
var fs = require('fs')
var hyperquest = require('hyperquest')
var cp = require('child_process')
var split = require('split')
var through = require('through2')

var url = 'https://api.github.com/repos/nodejs/io.js/contents'
var dirs = [
  '/test/parallel',
  '/test/pummel'
]

cp.execSync('rm -r node/*.js', { cwd: __dirname + '/../test' })
cp.execSync('rm -r node-es6/*.js', { cwd: __dirname + '/../test' })

var httpOpts = {
  headers: {
    'User-Agent': null
    // auth if github rate-limits you...
    // 'Authorization': 'Basic ' + Buffer('username:password').toString('base64'),
  }
}

dirs.forEach(function (dir) {
  var req = hyperquest(url + dir, httpOpts)
  req.pipe(concat(function (data) {
    if (req.response.statusCode !== 200) {
      throw new Error(url + dir + ': ' + data.toString())
    }
    downloadBufferTests(dir, JSON.parse(data))
  }))
})

function downloadBufferTests (dir, files) {
  files.forEach(function (file) {
    if (!/test-buffer.*/.test(file.name)) return

    var path
    if (file.name === 'test-buffer-iterator.js') {
      path = __dirname + '/../test/node-es6/' + file.name
    } else if (file.name === 'test-buffer-fakes.js') {
      // These teses only apply to node, where they're calling into C++ and need to
      // ensure the prototype can't be faked, or else there will be a segfault.
      return
    } else {
      path = __dirname + '/../test/node/' + file.name
    }

    hyperquest(file.download_url, httpOpts)
      .pipe(split())
      .pipe(testfixer(file.name))
      .pipe(fs.createWriteStream(path))
  })
}

function testfixer (filename) {
  var firstline = true

  return through(function (line, enc, cb) {
    line = line.toString()

    if (firstline) {
      // require buffer explicitly
      var preamble = 'var Buffer = require(\'../../\').Buffer;\n' +
        'if (process.env.OBJECT_IMPL) Buffer.TYPED_ARRAY_SUPPORT = false;'
      if (/use strict/.test(line)) line += '\n' + preamble
      else line + preamble + '\n' + line
      firstline = false
    }

    // comment out require('common')
    line = line.replace(/((var|const) common = require.*)/, 'var common = {};')

    // require browser buffer
    line = line.replace(/(.*)require\('buffer'\)(.*)/, '$1require(\'../../\')$2')

    // smalloc is only used for kMaxLength
    line = line.replace(
      /require\('smalloc'\)/,
      '{ kMaxLength: process.env.OBJECT_IMPL ? 0x3fffffff : 0x7fffffff }'
    )

    // comment out console logs
    line = line.replace(/(.*console\..*)/, '// $1')

    // we can't reliably test typed array max-sizes in the browser
    if (filename === 'test-buffer-big.js') {
      line = line.replace(/(.*new Int8Array.*RangeError.*)/, '// $1')
      line = line.replace(/(.*new ArrayBuffer.*RangeError.*)/, '// $1')
      line = line.replace(/(.*new Float64Array.*RangeError.*)/, '// $1')
    }

    // https://github.com/iojs/io.js/blob/v0.12/test/parallel/test-buffer.js#L38
    // we can't run this because we need to support
    // browsers that don't have typed arrays
    if (filename === 'test-buffer.js') {
      line = line.replace(/b\[0\] = -1;/, 'b[0] = 255;')
    }

    // https://github.com/iojs/io.js/blob/v0.12/test/parallel/test-buffer.js#L1138
    // unfortunately we can't run this as it touches
    // node streams which do an instanceof check
    // and crypto-browserify doesn't work in old
    // versions of ie
    if (filename === 'test-buffer.js') {
      line = line.replace(/^(\s*)(var crypto = require.*)/, '$1// $2')
      line = line.replace(/(crypto.createHash.*\))/, '1 /*$1*/')
    }

    cb(null, line + '\n')
  })
}
