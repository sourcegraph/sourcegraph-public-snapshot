var fs = require('fs-extra');
var path = require('path');
var package = require('../package');
var exec = require('child_process').exec;

var version = package.version;
var cwd = process.cwd();

var file = path.join(cwd, 'dist', 'amplitude-' + version + '.js');
var minfile = path.join(cwd, 'dist', 'amplitude-' + version + '-min.js');
var mingzfile = path.join(cwd, 'dist', 'amplitude-' + version + '-min.gz.js');

fs.copySync(path.join(cwd, 'amplitude.js'), file);
fs.copySync(path.join(cwd, 'amplitude.min.js'), minfile);
exec('gzip < ' + minfile + ' > ' + mingzfile);
