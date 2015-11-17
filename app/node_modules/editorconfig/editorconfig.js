var fs = require('fs');
var path = require('path');

var minimatch = require('./lib/fnmatch');
var iniparser = require('./lib/ini');
var Version = require('./lib/version');
var pkg = require('./package.json');


var knownProps = ['end_of_line', 'indent_style', 'indent_size',
    'insert_final_newline', 'trim_trailing_whitespace', 'charset'];


function fnmatch(filepath, glob) {
  var matchOptions = {matchBase: true, dot: true, noext: true};
  glob = glob.replace(/\*\*/g, '{*,**/**/**}');
  return minimatch(filepath, glob, matchOptions);
}


function getConfigFileNames(filepath, configname) {
  var old_dirname = filepath;
  var dirname = old_dirname;
  var paths = [];
  do {
    paths.push(path.join(dirname, configname || ".editorconfig"));
    old_dirname = dirname;
    dirname = path.dirname(old_dirname);
  } while(dirname != old_dirname);
  return paths;
}

function processMatches(matches, version) {

  // Set indent_size to "tab" if indent_size is unspecified and
  // indent_style is set to "tab".
  if ("indent_style" in matches && matches.indent_style === "tab" &&
    !("indent_size" in matches) && version.gte(new Version(0, 10))) {
    matches.indent_size = "tab";
  }

  // Set tab_width to indent_size if indent_size is specified and
  // tab_width is unspecified
  if ("indent_size" in matches && !("tab_width" in matches) &&
  matches.indent_size !== "tab")
    matches.tab_width = matches.indent_size;

  // Set indent_size to tab_width if indent_size is "tab"
  if("indent_size" in matches && "tab_width" in matches &&
  matches.indent_size === "tab")
    matches.indent_size = matches.tab_width;

  return matches;
}


function processOptions(options) {
  options = options || {};
  options.version = new Version(options.version || pkg.version);
  return options;
}


function parseFromFiles(filepath, configs, options) {
  var matches = {};

  configs.reverse().forEach(function (file) {
    var pathPrefix = path.dirname(file.name);
    var config = file.contents;
    config.forEach(function (section) {
      var fullGlob;
      var glob = section[0];
      var options = section[1];
      if (!glob) return;
      if (glob.indexOf('/') === -1) {
        fullGlob = path.join(pathPrefix, "**/" + glob);
      } else if (glob.indexOf('/') === 0) {
        fullGlob = path.join(pathPrefix, glob.substring(1));
      } else {
        fullGlob = path.join(pathPrefix, glob);
      }
      if (fnmatch(filepath, fullGlob)) {
        for (var key in options) {
          var value = options[key];
          if (knownProps.indexOf(key) !== -1) {
            value = value.toLowerCase();
          }
          try {
            value = JSON.parse(value);
          } catch(e){}
          matches[key.toLowerCase()] = value;
        }
      }
    });
  });

  return processMatches(matches, options.version);

}


function getConfigsForFiles(files) {
  var configs = [];
  for (var i = 0; i < files.length; i++) {
    files[i].contents = iniparser.parseString(files[i].contents);
    configs.push(files[i]);
    if (/^true$/i.test(files[i].contents[0][1].root)) break;
  }
  return configs;
}


function readConfigFiles(filepaths) {
  var files = [];
  filepaths.forEach(function (configFilePath) {
    if (fs.existsSync(configFilePath)) {
      files.push({
        name: configFilePath,
        contents: fs.readFileSync(configFilePath, 'utf-8')
      });
    }
  });
  return files;
}


module.exports.parseFromFiles = function(filepath, files, options) {
  filepath = path.resolve(filepath);
  options = processOptions(options);
  return parseFromFiles(filepath, getConfigsForFiles(files), options);
};


module.exports.parse = function(filepath, options) {
  filepath = path.resolve(filepath);
  options = processOptions(options);
  var filepaths = getConfigFileNames(path.dirname(filepath), options.config);
  var files = readConfigFiles(filepaths);
  return parseFromFiles(filepath, getConfigsForFiles(files), options);
};
