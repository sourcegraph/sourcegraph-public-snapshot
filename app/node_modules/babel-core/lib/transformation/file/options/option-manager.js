"use strict";

exports.__esModule = true;

var _stringify = require("babel-runtime/core-js/json/stringify");

var _stringify2 = _interopRequireDefault(_stringify);

var _assign = require("babel-runtime/core-js/object/assign");

var _assign2 = _interopRequireDefault(_assign);

var _getIterator2 = require("babel-runtime/core-js/get-iterator");

var _getIterator3 = _interopRequireDefault(_getIterator2);

var _typeof2 = require("babel-runtime/helpers/typeof");

var _typeof3 = _interopRequireDefault(_typeof2);

var _classCallCheck2 = require("babel-runtime/helpers/classCallCheck");

var _classCallCheck3 = _interopRequireDefault(_classCallCheck2);

var _node = require("../../../api/node");

var context = _interopRequireWildcard(_node);

var _plugin2 = require("../../plugin");

var _plugin3 = _interopRequireDefault(_plugin2);

var _babelMessages = require("babel-messages");

var messages = _interopRequireWildcard(_babelMessages);

var _index = require("./index");

var _resolve = require("../../../helpers/resolve");

var _resolve2 = _interopRequireDefault(_resolve);

var _json = require("json5");

var _json2 = _interopRequireDefault(_json);

var _pathIsAbsolute = require("path-is-absolute");

var _pathIsAbsolute2 = _interopRequireDefault(_pathIsAbsolute);

var _pathExists = require("path-exists");

var _pathExists2 = _interopRequireDefault(_pathExists);

var _cloneDeepWith = require("lodash/cloneDeepWith");

var _cloneDeepWith2 = _interopRequireDefault(_cloneDeepWith);

var _clone = require("lodash/clone");

var _clone2 = _interopRequireDefault(_clone);

var _merge = require("../../../helpers/merge");

var _merge2 = _interopRequireDefault(_merge);

var _config = require("./config");

var _config2 = _interopRequireDefault(_config);

var _removed = require("./removed");

var _removed2 = _interopRequireDefault(_removed);

var _path = require("path");

var _path2 = _interopRequireDefault(_path);

var _fs = require("fs");

var _fs2 = _interopRequireDefault(_fs);

function _interopRequireWildcard(obj) { if (obj && obj.__esModule) { return obj; } else { var newObj = {}; if (obj != null) { for (var key in obj) { if (Object.prototype.hasOwnProperty.call(obj, key)) newObj[key] = obj[key]; } } newObj.default = obj; return newObj; } }

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

var existsCache = {}; /* eslint max-len: 0 */

var jsonCache = {};

var BABELIGNORE_FILENAME = ".babelignore";
var BABELRC_FILENAME = ".babelrc";
var PACKAGE_FILENAME = "package.json";

function exists(filename) {
  var cached = existsCache[filename];
  if (cached == null) {
    return existsCache[filename] = _pathExists2.default.sync(filename);
  } else {
    return cached;
  }
}

var OptionManager = function () {
  function OptionManager(log) {
    (0, _classCallCheck3.default)(this, OptionManager);

    this.resolvedConfigs = [];
    this.options = OptionManager.createBareOptions();
    this.log = log;
  }

  OptionManager.memoisePluginContainer = function memoisePluginContainer(fn, loc, i, alias) {
    for (var _iterator = OptionManager.memoisedPlugins, _isArray = Array.isArray(_iterator), _i = 0, _iterator = _isArray ? _iterator : (0, _getIterator3.default)(_iterator);;) {
      var _ref;

      if (_isArray) {
        if (_i >= _iterator.length) break;
        _ref = _iterator[_i++];
      } else {
        _i = _iterator.next();
        if (_i.done) break;
        _ref = _i.value;
      }

      var cache = _ref;

      if (cache.container === fn) return cache.plugin;
    }

    var obj = void 0;

    if (typeof fn === "function") {
      obj = fn(context);
    } else {
      obj = fn;
    }

    if ((typeof obj === "undefined" ? "undefined" : (0, _typeof3.default)(obj)) === "object") {
      var _plugin = new _plugin3.default(obj, alias);
      OptionManager.memoisedPlugins.push({
        container: fn,
        plugin: _plugin
      });
      return _plugin;
    } else {
      throw new TypeError(messages.get("pluginNotObject", loc, i, typeof obj === "undefined" ? "undefined" : (0, _typeof3.default)(obj)) + loc + i);
    }
  };

  OptionManager.createBareOptions = function createBareOptions() {
    var opts = {};

    for (var _key in _config2.default) {
      var opt = _config2.default[_key];
      opts[_key] = (0, _clone2.default)(opt.default);
    }

    return opts;
  };

  OptionManager.normalisePlugin = function normalisePlugin(plugin, loc, i, alias) {
    plugin = plugin.__esModule ? plugin.default : plugin;

    if (!(plugin instanceof _plugin3.default)) {
      // allow plugin containers to be specified so they don't have to manually require
      if (typeof plugin === "function" || (typeof plugin === "undefined" ? "undefined" : (0, _typeof3.default)(plugin)) === "object") {
        plugin = OptionManager.memoisePluginContainer(plugin, loc, i, alias);
      } else {
        throw new TypeError(messages.get("pluginNotFunction", loc, i, typeof plugin === "undefined" ? "undefined" : (0, _typeof3.default)(plugin)));
      }
    }

    plugin.init(loc, i);

    return plugin;
  };

  OptionManager.normalisePlugins = function normalisePlugins(loc, dirname, plugins) {
    return plugins.map(function (val, i) {
      var plugin = void 0,
          options = void 0;

      if (!val) {
        throw new TypeError("Falsy value found in plugins");
      }

      // destructure plugins
      if (Array.isArray(val)) {
        plugin = val[0];
        options = val[1];
      } else {
        plugin = val;
      }

      var alias = typeof plugin === "string" ? plugin : loc + "$" + i;

      // allow plugins to be specified as strings
      if (typeof plugin === "string") {
        var pluginLoc = (0, _resolve2.default)("babel-plugin-" + plugin, dirname) || (0, _resolve2.default)(plugin, dirname);
        if (pluginLoc) {
          plugin = require(pluginLoc);
        } else {
          throw new ReferenceError(messages.get("pluginUnknown", plugin, loc, i, dirname));
        }
      }

      plugin = OptionManager.normalisePlugin(plugin, loc, i, alias);

      return [plugin, options];
    });
  };

  OptionManager.prototype.addConfig = function addConfig(loc, key) {
    var json = arguments.length <= 2 || arguments[2] === undefined ? _json2.default : arguments[2];

    if (this.resolvedConfigs.indexOf(loc) >= 0) {
      return false;
    }

    var content = _fs2.default.readFileSync(loc, "utf8");
    var opts = void 0;

    try {
      opts = jsonCache[content] = jsonCache[content] || json.parse(content);
      if (key) opts = opts[key];
    } catch (err) {
      err.message = loc + ": Error while parsing JSON - " + err.message;
      throw err;
    }

    this.mergeOptions({
      options: opts,
      alias: loc,
      dirname: _path2.default.dirname(loc)
    });
    this.resolvedConfigs.push(loc);

    return !!opts;
  };

  /**
   * This is called when we want to merge the input `opts` into the
   * base options (passed as the `extendingOpts`: at top-level it's the
   * main options, at presets level it's presets options).
   *
   *  - `alias` is used to output pretty traces back to the original source.
   *  - `loc` is used to point to the original config.
   *  - `dirname` is used to resolve plugins relative to it.
   */

  OptionManager.prototype.mergeOptions = function mergeOptions(_ref2) {
    var _this = this;

    var rawOpts = _ref2.options;
    var extendingOpts = _ref2.extending;
    var alias = _ref2.alias;
    var loc = _ref2.loc;
    var dirname = _ref2.dirname;

    alias = alias || "foreign";
    if (!rawOpts) return;

    //
    if ((typeof rawOpts === "undefined" ? "undefined" : (0, _typeof3.default)(rawOpts)) !== "object" || Array.isArray(rawOpts)) {
      this.log.error("Invalid options type for " + alias, TypeError);
    }

    //
    var opts = (0, _cloneDeepWith2.default)(rawOpts, function (val) {
      if (val instanceof _plugin3.default) {
        return val;
      }
    });

    //
    dirname = dirname || process.cwd();
    loc = loc || alias;

    for (var _key2 in opts) {
      var option = _config2.default[_key2];

      // check for an unknown option
      if (!option && this.log) {
        var pluginOptsInfo = "Check out http://babeljs.io/docs/usage/options/ for more info";

        if (_removed2.default[_key2]) {
          this.log.error("Using removed Babel 5 option: " + alias + "." + _key2 + " - " + _removed2.default[_key2].message, ReferenceError);
        } else {
          this.log.error("Unknown option: " + alias + "." + _key2 + ". " + pluginOptsInfo, ReferenceError);
        }
      }
    }

    // normalise options
    (0, _index.normaliseOptions)(opts);

    // resolve plugins
    if (opts.plugins) {
      opts.plugins = OptionManager.normalisePlugins(loc, dirname, opts.plugins);
    }

    // add extends clause
    if (opts.extends) {
      var extendsLoc = (0, _resolve2.default)(opts.extends, dirname);
      if (extendsLoc) {
        this.addConfig(extendsLoc);
      } else {
        if (this.log) this.log.error("Couldn't resolve extends clause of " + opts.extends + " in " + alias);
      }
      delete opts.extends;
    }

    // resolve presets
    if (opts.presets) {
      // If we're in the "pass per preset" mode, we resolve the presets
      // and keep them for further execution to calculate the options.
      if (opts.passPerPreset) {
        opts.presets = this.resolvePresets(opts.presets, dirname, function (preset, presetLoc) {
          _this.mergeOptions({
            options: preset,
            extending: preset,
            alias: presetLoc,
            loc: presetLoc,
            dirname: dirname
          });
        });
      } else {
        // Otherwise, just merge presets options into the main options.
        this.mergePresets(opts.presets, dirname);
        delete opts.presets;
      }
    }

    // env
    var envOpts = void 0;
    var envKey = process.env.BABEL_ENV || process.env.NODE_ENV || "development";
    if (opts.env) {
      envOpts = opts.env[envKey];
      delete opts.env;
    }

    // Merge them into current extending options in case of top-level
    // options. In case of presets, just re-assign options which are got
    // normalized during the `mergeOptions`.
    if (rawOpts === extendingOpts) {
      (0, _assign2.default)(extendingOpts, opts);
    } else {
      (0, _merge2.default)(extendingOpts || this.options, opts);
    }

    // merge in env options
    this.mergeOptions({
      options: envOpts,
      extending: extendingOpts,
      alias: alias + ".env." + envKey,
      dirname: dirname
    });
  };

  /**
   * Merges all presets into the main options in case we are not in the
   * "pass per preset" mode. Otherwise, options are calculated per preset.
   */


  OptionManager.prototype.mergePresets = function mergePresets(presets, dirname) {
    var _this2 = this;

    this.resolvePresets(presets, dirname, function (presetOpts, presetLoc) {
      _this2.mergeOptions({
        options: presetOpts,
        alias: presetLoc,
        loc: presetLoc,
        dirname: _path2.default.dirname(presetLoc || "")
      });
    });
  };

  /**
   * Resolves presets options which can be either direct object data,
   * or a module name to require.
   */


  OptionManager.prototype.resolvePresets = function resolvePresets(presets, dirname, onResolve) {
    return presets.map(function (val) {
      if (typeof val === "string") {
        var presetLoc = (0, _resolve2.default)("babel-preset-" + val, dirname) || (0, _resolve2.default)(val, dirname);
        if (presetLoc) {
          var _val = require(presetLoc);
          onResolve && onResolve(_val, presetLoc);
          return _val;
        } else {
          throw new Error("Couldn't find preset " + (0, _stringify2.default)(val) + " relative to directory " + (0, _stringify2.default)(dirname));
        }
      } else if ((typeof val === "undefined" ? "undefined" : (0, _typeof3.default)(val)) === "object") {
        onResolve && onResolve(val);
        return val;
      } else {
        throw new Error("Unsupported preset format: " + val + ".");
      }
    });
  };

  OptionManager.prototype.addIgnoreConfig = function addIgnoreConfig(loc) {
    var file = _fs2.default.readFileSync(loc, "utf8");
    var lines = file.split("\n");

    lines = lines.map(function (line) {
      return line.replace(/#(.*?)$/, "").trim();
    }).filter(function (line) {
      return !!line;
    });

    this.mergeOptions({
      options: { ignore: lines },
      loc: loc
    });
  };

  OptionManager.prototype.findConfigs = function findConfigs(loc) {
    if (!loc) return;

    if (!(0, _pathIsAbsolute2.default)(loc)) {
      loc = _path2.default.join(process.cwd(), loc);
    }

    var foundConfig = false;
    var foundIgnore = false;

    while (loc !== (loc = _path2.default.dirname(loc))) {
      if (!foundConfig) {
        var configLoc = _path2.default.join(loc, BABELRC_FILENAME);
        if (exists(configLoc)) {
          this.addConfig(configLoc);
          foundConfig = true;
        }

        var pkgLoc = _path2.default.join(loc, PACKAGE_FILENAME);
        if (!foundConfig && exists(pkgLoc)) {
          foundConfig = this.addConfig(pkgLoc, "babel", JSON);
        }
      }

      if (!foundIgnore) {
        var ignoreLoc = _path2.default.join(loc, BABELIGNORE_FILENAME);
        if (exists(ignoreLoc)) {
          this.addIgnoreConfig(ignoreLoc);
          foundIgnore = true;
        }
      }

      if (foundIgnore && foundConfig) return;
    }
  };

  OptionManager.prototype.normaliseOptions = function normaliseOptions() {
    var opts = this.options;

    for (var _key3 in _config2.default) {
      var option = _config2.default[_key3];
      var val = opts[_key3];

      // optional
      if (!val && option.optional) continue;

      // aliases
      if (option.alias) {
        opts[option.alias] = opts[option.alias] || val;
      } else {
        opts[_key3] = val;
      }
    }
  };

  OptionManager.prototype.init = function init() {
    var opts = arguments.length <= 0 || arguments[0] === undefined ? {} : arguments[0];

    var filename = opts.filename;

    // resolve all .babelrc files
    if (opts.babelrc !== false) {
      this.findConfigs(filename);
    }

    // merge in base options
    this.mergeOptions({
      options: opts,
      alias: "base",
      dirname: filename && _path2.default.dirname(filename)
    });

    // normalise
    this.normaliseOptions(opts);

    return this.options;
  };

  return OptionManager;
}();

exports.default = OptionManager;


OptionManager.memoisedPlugins = [];
module.exports = exports["default"];