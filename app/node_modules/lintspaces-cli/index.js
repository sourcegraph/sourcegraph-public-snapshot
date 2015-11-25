#!/usr/bin/env node

var fs = require('fs')
  , util = require('util')
  , path = require('path')
  , colors = require('colors') // It's colours I tell ya. Colours!!!
  , program = require('commander')
  , Validator = require('lintspaces')
  , version = require('./package.json').version
  , validator = null
  , targetFiles = null
  , files = null;


/**
 * Split comma separated list into array.
 * @param {String}
 */
function list (l) {
  return l.split(',');
}

/**
 * Check does the provided editorconfig exist
 * @param {String}
 */
function resolveEditorConfig (e) {
  if (e) {
    e = path.resolve(e)

    if (!fs.existsSync(e)) {
      console.log('Error: Specified .editorconfig "%s" doesn\'t exist'.red, e);
      process.exit(1);
    }

    return e;
  }

  return e;
}


program.version(version)
  .option('-n, --newline', 'Require newline at end of file.')
  .option('-g, --guessindentation', 'Tries to guess the indention of a line ' +
    'depending on previous lines.')
  .option('-b, --skiptrailingonblank', 'Skip blank lines in trailingspaces ' +
    'check.')
  .option('-it, --trailingspacestoignores', 'Ignore trailing spaces in ' +
    'ignores.')
  .option('-l, --maxnewlines <n>', 'Specify max number of newlines between' +
    ' blocks.', parseInt)
  .option('-t, --trailingspaces', 'Tests for useless whitespaces' +
    ' (trailing whitespaces) at each lineending of all files.')
  .option('-d, --indentation <s>', 'Check indentation is "tabs" or "spaces".')
  .option('-s, --spaces <n>', 'Used in conjunction with -d to set number of ' +
    'spaces.', parseInt)
  .option('-i, --ignores <items>', 'Comma separated list of ignores.', list)
  .option('-e, --editorconfig <s>', 'Use editorconfig specified at this ' +
   'file path for settings.', resolveEditorConfig)
  .parse(process.argv);


// Setup validator with user options
validator = new Validator({
  newline: program.newline,
  newlineMaximum: program.maxnewlines,
  trailingspaces: program.trailingspaces,
  indentation: program.indentation,
  spaces: program.spaces,
  ignores: program.ignores,
  editorconfig: program.editorconfig,
  indentationGuess: program.guessindentation,
  trailingspacesSkipBlanks: program.skiptrailingonblank,
  trailingspacesToIgnores: program.trailingspacesToIgnores
});


// Get files from args to support **/* syntax. Probably not the best way...
targetFiles = process.argv.slice(2).filter(fs.existsSync.bind(fs));


// Run validation
for (var file in targetFiles) {
  validator.validate(path.resolve(targetFiles[file]));
}
files = validator.getInvalidFiles();


// Output results
for (var file in files) {
  var curFile = files[file];
  console.warn(util.format('\nFile: %s', file).red.underline);

  for (var line in curFile) {
    var curLine = curFile[line];

    for(var err in curLine) {
      var curErr = curLine[err]
        , msg = ''
        , errMsg = curErr.type;

      if (errMsg.toLowerCase() === 'warning') {
        errMsg = errMsg.red;
      } else {
        errMsg = errMsg.green;
      }

      msg = util.format('Line: %s %s [%s]', line, curErr.message, errMsg);

      console.warn(msg);
    }
  }
}


// Give error exit code if required
if (Object.keys(files).length) {
  process.exit(1);
}
