#!/usr/bin/env node


'use strict';

var _path = require('path');

var _path2 = _interopRequireDefault(_path);

var _gaze = require('gaze');

var _gaze2 = _interopRequireDefault(_gaze);

var _glob = require('glob');

var _glob2 = _interopRequireDefault(_glob);

var _yargs = require('yargs');

var _yargs2 = _interopRequireDefault(_yargs);

var _chalk = require('chalk');

var _chalk2 = _interopRequireDefault(_chalk);

var _dtsCreator = require('./dtsCreator');

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

var yarg = _yargs2.default.usage('Create .css.d.ts from CSS modules *.css files.\nUsage: $0 [options] <input directory>').example('$0 src/styles').example('$0 src -o dist').example('$0 -p styles/**/*.icss -w').detectLocale(false).demand(['_']).alias('o', 'outDir').describe('o', 'Output directory').alias('p', 'pattern').describe('p', 'Glob pattern with css files').alias('w', 'watch').describe('w', 'Watch input directory\'s css files or pattern').boolean('w').alias('h', 'help').help('h').version(function () {
  return require('../package.json').version;
});
var argv = yarg.argv;
var creator = void 0;

var writeFile = function writeFile(f) {
  creator.create(f, null, !!argv.w).then(function (content) {
    return content.writeFile();
  }).then(function (content) {
    console.log('Wrote ' + _chalk2.default.green(content.outputFilePath));
    content.messageList.forEach(function (message) {
      console.warn(_chalk2.default.yellow('[Warn] ' + message));
    });
  }).catch(function (reason) {
    return console.error(_chalk2.default.red('[Error] ' + reason));
  });
};

var main = function main() {
  var rootDir = void 0,
      searchDir = void 0;
  if (argv.h) {
    yarg.showHelp();
    return;
  }

  if (argv._ && argv._[0]) {
    searchDir = argv._[0];
  } else if (argv.p) {
    searchDir = './';
  } else {
    yarg.showHelp();
    return;
  }
  var filesPattern = _path2.default.join(searchDir, argv.p || '**/*.css');
  rootDir = process.cwd();
  creator = new _dtsCreator.DtsCreator({ rootDir: rootDir, searchDir: searchDir, outDir: argv.o });

  if (!argv.w) {
    (0, _glob2.default)(filesPattern, null, function (err, files) {
      if (err) {
        console.error(err);
        return;
      }
      if (!files || !files.length) return;
      files.forEach(writeFile);
    });
  } else {
    console.log('Watch ' + filesPattern + '...');
    (0, _gaze2.default)(filesPattern, function (err, files) {
      this.on('changed', writeFile);
      this.on('added', writeFile);
    });
  }
};

main();