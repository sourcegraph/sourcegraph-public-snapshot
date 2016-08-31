"use strict";

var _slicedToArray = function () { function sliceIterator(arr, i) { var _arr = []; var _n = true; var _d = false; var _e = undefined; try { for (var _i = arr[Symbol.iterator](), _s; !(_n = (_s = _i.next()).done); _n = true) { _arr.push(_s.value); if (i && _arr.length === i) break; } } catch (err) { _d = true; _e = err; } finally { try { if (!_n && _i["return"]) _i["return"](); } finally { if (_d) throw _e; } } return _arr; } return function (arr, i) { if (Array.isArray(arr)) { return arr; } else if (Symbol.iterator in Object(arr)) { return sliceIterator(arr, i); } else { throw new TypeError("Invalid attempt to destructure non-iterable instance"); } }; }();

var _path = require("path");

var _tape = require("tape");

var _tape2 = _interopRequireDefault(_tape);

var _memoryFs = require("memory-fs");

var _memoryFs2 = _interopRequireDefault(_memoryFs);

var _webpack = require("webpack");

var _webpack2 = _interopRequireDefault(_webpack);

var _index = require("../index");

var _index2 = _interopRequireDefault(_index);

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

var EXPECTED_FILENAME_LIST = ["CHANGELOG.md", "README.md", "src/__tests__/index.spec.js", "package.json"];

(0, _tape2.default)("UnusedFilesWebpackPlugin", function (t) {
  var compiler = (0, _webpack2.default)({
    context: (0, _path.resolve)(__dirname, "../../"),
    entry: {
      UnusedFilesWebpackPlugin: (0, _path.resolve)(__dirname, "../index.js")
    },
    output: {
      path: __dirname },
    plugins: [new _index2.default()]
  });
  compiler.outputFileSystem = new _memoryFs2.default();

  compiler.run(function (err, stats) {
    t.equal(err, null);

    var warnings = stats.compilation.warnings;

    t.equal(warnings.length, 1);

    var _warnings = _slicedToArray(warnings, 1);

    var unusedFilesError = _warnings[0];

    t.equal(unusedFilesError instanceof Error, true);

    var message = unusedFilesError.message;

    var containsExpected = EXPECTED_FILENAME_LIST.every(function (filename) {
      return message.match(filename);
    });
    t.equal(containsExpected, true);

    t.end();
  });
});