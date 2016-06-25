"use strict";

exports.__esModule = true;

var _classCallCheck2 = require("babel-runtime/helpers/classCallCheck");

var _classCallCheck3 = _interopRequireDefault(_classCallCheck2);

var _repeat = require("lodash/repeat");

var _repeat2 = _interopRequireDefault(_repeat);

var _trimEnd = require("lodash/trimEnd");

var _trimEnd2 = _interopRequireDefault(_trimEnd);

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

/**
 * Buffer for collecting generated output.
 */

var Buffer = function () {
  function Buffer(position, format) {
    (0, _classCallCheck3.default)(this, Buffer);

    this.printedCommentStarts = {};
    this.parenPushNewlineState = null;
    this.position = position;
    this._indent = format.indent.base;
    this.format = format;
    this.buf = "";

    // Maintaining a reference to the last char in the buffer is an optimization
    // to make sure that v8 doesn't "flatten" the string more often than needed
    // see https://github.com/babel/babel/pull/3283 for details.
    this.last = "";

    this.map = null;
    this._sourcePosition = {
      line: null,
      column: null,
      filename: null
    };
    this._endsWithWord = false;
  }

  /**
   * Description
   */

  Buffer.prototype.catchUp = function catchUp(node) {
    // catch up to this nodes newline if we're behind
    if (node.loc && this.format.retainLines && this.buf) {
      while (this.position.line < node.loc.start.line) {
        this.push("\n");
      }
    }
  };

  /**
   * Get the current trimmed buffer.
   */

  Buffer.prototype.get = function get() {
    return (0, _trimEnd2.default)(this.buf);
  };

  /**
   * Get the current indent.
   */

  Buffer.prototype.getIndent = function getIndent() {
    if (this.format.compact || this.format.concise) {
      return "";
    } else {
      return (0, _repeat2.default)(this.format.indent.style, this._indent);
    }
  };

  /**
   * Get the current indent size.
   */

  Buffer.prototype.indentSize = function indentSize() {
    return this.getIndent().length;
  };

  /**
   * Increment indent size.
   */

  Buffer.prototype.indent = function indent() {
    this._indent++;
  };

  /**
   * Decrement indent size.
   */

  Buffer.prototype.dedent = function dedent() {
    this._indent--;
  };

  /**
   * Add a semicolon to the buffer.
   */

  Buffer.prototype.semicolon = function semicolon() {
    this.token(";");
  };

  /**
   * Add a right brace to the buffer.
   */

  Buffer.prototype.rightBrace = function rightBrace() {
    this.newline(true);
    if (this.format.minified && !this._lastPrintedIsEmptyStatement) {
      this._removeLast(";");
    }
    this.token("}");
  };

  /**
   * Add a keyword to the buffer.
   */

  Buffer.prototype.keyword = function keyword(name) {
    this.word(name);
    this.space();
  };

  /**
   * Add a space to the buffer unless it is compact.
   */

  Buffer.prototype.space = function space() {
    if (this.format.compact) return;

    if (this.buf && !this.endsWith(" ") && !this.endsWith("\n")) {
      this.push(" ");
    }
  };

  /**
   * Writes a token that can't be safely parsed without taking whitespace into account.
   */

  Buffer.prototype.word = function word(str) {
    if (this._endsWithWord) this.push(" ");

    this.push(str);
    this._endsWithWord = true;
  };

  /**
   * Writes a simple token.
   */

  Buffer.prototype.token = function token(str) {
    // space is mandatory to avoid outputting <!--
    // http://javascript.spec.whatwg.org/#comment-syntax
    if (str === "--" && this.last === "!" ||

    // Need spaces for operators of the same kind to avoid: `a+++b`
    str[0] === "+" && this.last === "+" || str[0] === "-" && this.last === "-") {
      this.push(" ");
    }

    this.push(str);
  };

  /**
   * Remove the last character.
   */

  Buffer.prototype.removeLast = function removeLast(cha) {
    if (this.format.compact) return;
    return this._removeLast(cha);
  };

  Buffer.prototype._removeLast = function _removeLast(cha) {
    if (!this.endsWith(cha)) return;
    this.buf = this.buf.slice(0, -1);
    this.last = this.buf[this.buf.length - 1];
    this.position.unshift(cha);
  };

  /**
   * Set some state that will be modified if a newline has been inserted before any
   * non-space characters.
   *
   * This is to prevent breaking semantics for terminatorless separator nodes. eg:
   *
   *    return foo;
   *
   * returns `foo`. But if we do:
   *
   *   return
   *   foo;
   *
   *  `undefined` will be returned and not `foo` due to the terminator.
   */

  Buffer.prototype.startTerminatorless = function startTerminatorless() {
    return this.parenPushNewlineState = {
      printed: false
    };
  };

  /**
   * Print an ending parentheses if a starting one has been printed.
   */

  Buffer.prototype.endTerminatorless = function endTerminatorless(state) {
    if (state.printed) {
      this.dedent();
      this.newline();
      this.token(")");
    }
  };

  /**
   * Add a newline (or many newlines), maintaining formatting.
   * Strips multiple newlines if removeLast is true.
   */

  Buffer.prototype.newline = function newline(i, removeLast) {
    if (this.format.retainLines || this.format.compact) return;

    if (this.format.concise) {
      this.space();
      return;
    }

    // never allow more than two lines
    if (this.endsWith("\n\n")) return;

    if (typeof i === "boolean") removeLast = i;
    if (typeof i !== "number") i = 1;

    i = Math.min(2, i);
    if (this.endsWith("{\n") || this.endsWith(":\n")) i--;
    if (i <= 0) return;

    // remove the last newline
    if (removeLast) {
      this.removeLast("\n");
    }

    this.removeLast(" ");
    this._removeSpacesAfterLastNewline();
    for (var j = 0; j < i; j++) {
      this.push("\n");
    }
  };

  /**
   * If buffer ends with a newline and some spaces after it, trim those spaces.
   */

  Buffer.prototype._removeSpacesAfterLastNewline = function _removeSpacesAfterLastNewline() {
    var lastNewlineIndex = this.buf.lastIndexOf("\n");
    if (lastNewlineIndex >= 0 && this.get().length <= lastNewlineIndex) {
      var toRemove = this.buf.slice(lastNewlineIndex + 1);
      this.buf = this.buf.substring(0, lastNewlineIndex + 1);
      this.last = "\n";
      this.position.unshift(toRemove);
    }
  };

  /**
   * Sets a given position as the current source location so generated code after this call
   * will be given this position in the sourcemap.
   */

  Buffer.prototype.source = function source(prop, loc) {
    if (prop && !loc) return;

    var pos = loc ? loc[prop] : null;

    this._sourcePosition.line = pos ? pos.line : null;
    this._sourcePosition.column = pos ? pos.column : null;
    this._sourcePosition.filename = loc && loc.filename || null;
  };

  /**
   * Call a callback with a specific source location and restore on completion.
   */

  Buffer.prototype.withSource = function withSource(prop, loc, cb) {
    if (!this.opts.sourceMaps) return cb();

    // Use the call stack to manage a stack of "source location" data.
    var originalLine = this._sourcePosition.line;
    var originalColumn = this._sourcePosition.column;
    var originalFilename = this._sourcePosition.filename;

    this.source(prop, loc);

    cb();

    this._sourcePosition.line = originalLine;
    this._sourcePosition.column = originalColumn;
    this._sourcePosition.filename = originalFilename;
  };

  /**
   * Push a string to the buffer, maintaining indentation and newlines.
   */

  Buffer.prototype.push = function push(str) {
    if (!this.format.compact && this._indent && str[0] !== "\n") {
      // we've got a newline before us so prepend on the indentation
      if (this.endsWith("\n")) str = this.getIndent() + str;
    }

    // see startTerminatorless() instance method
    var parenPushNewlineState = this.parenPushNewlineState;
    if (parenPushNewlineState) {
      for (var i = 0; i < str.length; i++) {
        var cha = str[i];

        // we can ignore spaces since they wont interupt a terminatorless separator
        if (cha === " ") continue;

        this.parenPushNewlineState = null;

        if (cha === "\n" || cha === "/") {
          // we're going to break this terminator expression so we need to add a parentheses
          str = "(" + str;
          this.indent();
          parenPushNewlineState.printed = true;
        }

        break;
      }
    }

    // If there the line is ending, adding a new mapping marker is redundant
    if (this.opts.sourceMaps && str[0] !== "\n") this.map.mark(this._sourcePosition);

    //
    this.position.push(str);
    this.buf += str;
    this.last = str[str.length - 1];

    // Clear any state-tracking flags that may have been set.
    this._endsWithWord = false;
  };

  /**
   * Test if the buffer ends with a string.
   */

  Buffer.prototype.endsWith = function endsWith(str) {
    var _this = this;

    if (Array.isArray(str)) return str.some(function (s) {
      return _this.endsWith(s);
    });

    if (str.length === 1) {
      return this.last === str;
    } else {
      return this.buf.slice(-str.length) === str;
    }
  };

  return Buffer;
}();

exports.default = Buffer;
module.exports = exports["default"];