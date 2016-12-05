"use strict";

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.UnusedFilesWebpackPlugin = undefined;

var _extends = Object.assign || function (target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i]; for (var key in source) { if (Object.prototype.hasOwnProperty.call(source, key)) { target[key] = source[key]; } } } return target; };

var _createClass = function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; }();

var _path = require("path");

var _glob = require("glob");

var _glob2 = _interopRequireDefault(_glob);

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

function _defineProperty(obj, key, value) { if (key in obj) { Object.defineProperty(obj, key, { value: value, enumerable: true, configurable: true, writable: true }); } else { obj[key] = value; } return obj; }

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

var UnusedFilesWebpackPlugin = exports.UnusedFilesWebpackPlugin = function () {
  function UnusedFilesWebpackPlugin() {
    var options = arguments.length <= 0 || arguments[0] === undefined ? {} : arguments[0];

    _classCallCheck(this, UnusedFilesWebpackPlugin);

    this.options = _extends({
      pattern: "**/*.*"
    }, options, {
      failOnUnused: options.failOnUnused === true
    });

    this.globOptions = _extends({
      ignore: "node_modules/**/*"
    }, options.globOptions);
  }

  _createClass(UnusedFilesWebpackPlugin, [{
    key: "apply",
    value: function apply(compiler) {
      var _this = this;

      compiler.plugin("after-emit", function (compilation, done) {
        return _this._applyAfterEmit(compiler, compilation, done);
      });
    }
  }, {
    key: "_applyAfterEmit",
    value: function _applyAfterEmit(compiler, compilation, done) {
      var _this2 = this;

      var globOptions = this._getGlobOptions(compiler);
      var fileDepsMap = this._getFileDepsMap(compilation);
      var absolutePathResolver = function absolutePathResolver(it) {
        return (0, _path.join)(globOptions.cwd, it);
      };

      var handleError = function handleError(err) {
        if (compilation.bail) {
          done(err);
        } else {
          compilation.errors.push(err);
        }
      };

      (0, _glob2.default)(this.options.pattern, globOptions, function (err, files) {
        if (err) {
          handleError(err);
          return;
        }
        var unused = files.filter(function (filepath) {
          return !(absolutePathResolver(filepath) in fileDepsMap);
        });
        if (unused.length === 0) {
          done();
          return;
        }
        var error = new Error("\nUnusedFilesWebpackPlugin found some unused files:\n" + unused.join("\n"));

        if (_this2.options.failOnUnused) {
          handleError(error);
        } else {
          compilation.warnings.push(error);
          done();
        }
      });
    }
  }, {
    key: "_getGlobOptions",
    value: function _getGlobOptions(compiler) {
      return _extends({
        cwd: compiler.context
      }, this.globOptions);
    }
  }, {
    key: "_getFileDepsMap",
    value: function _getFileDepsMap(compilation) {
      var fileDepsBy = compilation.fileDependencies.reduce(function (acc, usedFilepath) {
        return _extends({}, acc, _defineProperty({}, usedFilepath, usedFilepath));
      }, {});

      var assets = compilation.assets;

      Object.keys(assets).forEach(function (assetRelpath) {
        var existsAt = assets[assetRelpath].existsAt;
        fileDepsBy[existsAt] = existsAt;
      });
      return fileDepsBy;
    }
  }]);

  return UnusedFilesWebpackPlugin;
}();

exports.default = UnusedFilesWebpackPlugin;