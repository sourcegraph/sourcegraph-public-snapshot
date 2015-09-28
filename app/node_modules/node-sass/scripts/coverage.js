/*!
 * node-sass: scripts/coverage.js
 */

require('../lib/extensions');

var bin = require('path').join.bind(null, __dirname, '..', 'node_modules', '.bin'),
    spawn = require('child_process').spawn;

/**
 * Run test suite
 *
 * @api private
 */

function suite() {
  process.env.NODESASS_COV = 1;

  var coveralls = spawn(bin('coveralls'));

  var args = [bin('_mocha')].concat(['--reporter', 'mocha-lcov-reporter']);
  var mocha = spawn(process.sass.runtime.execPath, args, {
    env: process.env
  });

  mocha.on('error', function(err) {
    console.error(err);
    process.exit(1);
  });

  mocha.stderr.setEncoding('utf8');
  mocha.stderr.on('data', function(err) {
    console.error(err);
    process.exit(1);
  });

  mocha.stdout.pipe(coveralls.stdin);
}

/**
 * Generate coverage files
 *
 * @api private
 */

function coverage() {
  var jscoverage = spawn(bin('jscoverage'), ['lib', 'lib-cov']);

  jscoverage.on('error', function(err) {
    console.error(err);
    process.exit(1);
  });

  jscoverage.stderr.setEncoding('utf8');
  jscoverage.stderr.on('data', function(err) {
    console.error(err);
    process.exit(1);
  });

  jscoverage.on('close', suite);
}

/**
 * Run
 */

coverage();
