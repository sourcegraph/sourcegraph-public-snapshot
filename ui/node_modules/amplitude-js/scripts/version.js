var fs = require('fs');
var path = require('path');
var package = require('../package');
var component = require('../component');
var previous = require('../src/version');

var version = package.version;

var cwd = process.cwd();

function replaceVersion(filepath) {
  var filename = path.join(cwd, filepath);
  fs.writeFileSync(filename, fs.readFileSync(filename, 'utf-8').split(previous).join(version));
  console.log('Updated ', filepath);
}

console.log('Updating to version ' + version);

component.version = version;
fs.writeFileSync(path.join(cwd, 'component.json'), JSON.stringify(component, null, 2) + '\n');
console.log('Updated component.json');

var files = [
  'README.md',
  path.join('src', 'amplitude-snippet.js'),
  path.join('src', 'version.js'),
];
files.map(replaceVersion);

console.log('Updated version from', previous, 'to', version);
