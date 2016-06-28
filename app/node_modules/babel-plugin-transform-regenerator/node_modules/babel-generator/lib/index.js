"use strict";

exports.__esModule = true;
exports.CodeGenerator = undefined;

var _classCallCheck2 = require("babel-runtime/helpers/classCallCheck");

var _classCallCheck3 = _interopRequireDefault(_classCallCheck2);

var _possibleConstructorReturn2 = require("babel-runtime/helpers/possibleConstructorReturn");

var _possibleConstructorReturn3 = _interopRequireDefault(_possibleConstructorReturn2);

var _inherits2 = require("babel-runtime/helpers/inherits");

var _inherits3 = _interopRequireDefault(_inherits2);

exports.default = function (ast, opts, code) {
  var gen = new Generator(ast, opts, code);
  return gen.generate();
};

var _detectIndent = require("detect-indent");

var _detectIndent2 = _interopRequireDefault(_detectIndent);

var _whitespace = require("./whitespace");

var _whitespace2 = _interopRequireDefault(_whitespace);

var _sourceMap = require("./source-map");

var _sourceMap2 = _interopRequireDefault(_sourceMap);

var _position = require("./position");

var _position2 = _interopRequireDefault(_position);

var _babelMessages = require("babel-messages");

var messages = _interopRequireWildcard(_babelMessages);

var _printer = require("./printer");

var _printer2 = _interopRequireDefault(_printer);

function _interopRequireWildcard(obj) { if (obj && obj.__esModule) { return obj; } else { var newObj = {}; if (obj != null) { for (var key in obj) { if (Object.prototype.hasOwnProperty.call(obj, key)) newObj[key] = obj[key]; } } newObj.default = obj; return newObj; } }

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

/**
 * Babel's code generator, turns an ast into code, maintaining sourcemaps,
 * user preferences, and valid output.
 */

var Generator = function (_Printer) {
  (0, _inherits3.default)(Generator, _Printer);

  function Generator(ast, opts, code) {
    (0, _classCallCheck3.default)(this, Generator);

    opts = opts || {};

    var comments = ast.comments || [];
    var tokens = ast.tokens || [];
    var format = Generator.normalizeOptions(code, opts, tokens);

    var position = new _position2.default();

    var _this = (0, _possibleConstructorReturn3.default)(this, _Printer.call(this, position, format));

    _this.comments = comments;
    _this.position = position;
    _this.tokens = tokens;
    _this.format = format;
    _this.opts = opts;
    _this.ast = ast;
    _this._inForStatementInitCounter = 0;

    _this.whitespace = new _whitespace2.default(tokens);
    _this.map = new _sourceMap2.default(position, opts, code);
    return _this;
  }

  /**
   * Normalize generator options, setting defaults.
   *
   * - Detects code indentation.
   * - If `opts.compact = "auto"` and the code is over 100KB, `compact` will be set to `true`.
   */

  Generator.normalizeOptions = function normalizeOptions(code, opts, tokens) {
    var style = "  ";
    if (code && typeof code === "string") {
      var _indent = (0, _detectIndent2.default)(code).indent;
      if (_indent && _indent !== " ") style = _indent;
    }

    var format = {
      auxiliaryCommentBefore: opts.auxiliaryCommentBefore,
      auxiliaryCommentAfter: opts.auxiliaryCommentAfter,
      shouldPrintComment: opts.shouldPrintComment,
      retainLines: opts.retainLines,
      comments: opts.comments == null || opts.comments,
      compact: opts.compact,
      minified: opts.minified,
      concise: opts.concise,
      quotes: opts.quotes || Generator.findCommonStringDelimiter(code, tokens),
      indent: {
        adjustMultilineComment: true,
        style: style,
        base: 0
      }
    };

    if (format.minified) {
      format.compact = true;
    }

    if (format.compact === "auto") {
      format.compact = code.length > 100000; // 100KB

      if (format.compact) {
        console.error("[BABEL] " + messages.get("codeGeneratorDeopt", opts.filename, "100KB"));
      }
    }

    if (format.compact) {
      format.indent.adjustMultilineComment = false;
    }

    return format;
  };

  /**
   * Determine if input code uses more single or double quotes.
   */


  Generator.findCommonStringDelimiter = function findCommonStringDelimiter(code, tokens) {
    var occurences = {
      single: 0,
      double: 0
    };

    var checked = 0;

    for (var i = 0; i < tokens.length; i++) {
      var token = tokens[i];
      if (token.type.label !== "string") continue;

      var raw = code.slice(token.start, token.end);
      if (raw[0] === "'") {
        occurences.single++;
      } else {
        occurences.double++;
      }

      checked++;
      if (checked >= 3) break;
    }
    if (occurences.single > occurences.double) {
      return "single";
    } else {
      return "double";
    }
  };

  /**
   * Generate code and sourcemap from ast.
   *
   * Appends comments that weren't attached to any node to the end of the generated output.
   */

  Generator.prototype.generate = function generate() {
    this.print(this.ast);
    this.printAuxAfterComment();

    return {
      map: this.map.get(),
      code: this.get()
    };
  };

  return Generator;
}(_printer2.default);

/**
 * We originally exported the Generator class above, but to make it extra clear that it is a private API,
 * we have moved that to an internal class instance and simplified the interface to the two public methods
 * that we wish to support.
 */

var CodeGenerator = exports.CodeGenerator = function () {
  function CodeGenerator(ast, opts, code) {
    (0, _classCallCheck3.default)(this, CodeGenerator);

    this._generator = new Generator(ast, opts, code);
  }

  CodeGenerator.prototype.generate = function generate() {
    return this._generator.generate();
  };

  return CodeGenerator;
}();