"use strict";

exports.__esModule = true;

var _assign = require("babel-runtime/core-js/object/assign");

var _assign2 = _interopRequireDefault(_assign);

var _getIterator2 = require("babel-runtime/core-js/get-iterator");

var _getIterator3 = _interopRequireDefault(_getIterator2);

var _stringify = require("babel-runtime/core-js/json/stringify");

var _stringify2 = _interopRequireDefault(_stringify);

var _classCallCheck2 = require("babel-runtime/helpers/classCallCheck");

var _classCallCheck3 = _interopRequireDefault(_classCallCheck2);

var _possibleConstructorReturn2 = require("babel-runtime/helpers/possibleConstructorReturn");

var _possibleConstructorReturn3 = _interopRequireDefault(_possibleConstructorReturn2);

var _inherits2 = require("babel-runtime/helpers/inherits");

var _inherits3 = _interopRequireDefault(_inherits2);

var _repeat = require("lodash/repeat");

var _repeat2 = _interopRequireDefault(_repeat);

var _buffer = require("./buffer");

var _buffer2 = _interopRequireDefault(_buffer);

var _node = require("./node");

var n = _interopRequireWildcard(_node);

var _babelTypes = require("babel-types");

var t = _interopRequireWildcard(_babelTypes);

function _interopRequireWildcard(obj) { if (obj && obj.__esModule) { return obj; } else { var newObj = {}; if (obj != null) { for (var key in obj) { if (Object.prototype.hasOwnProperty.call(obj, key)) newObj[key] = obj[key]; } } newObj.default = obj; return newObj; } }

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

/* eslint max-len: 0 */

var Printer = function (_Buffer) {
  (0, _inherits3.default)(Printer, _Buffer);

  function Printer() {
    (0, _classCallCheck3.default)(this, Printer);

    for (var _len = arguments.length, args = Array(_len), _key = 0; _key < _len; _key++) {
      args[_key] = arguments[_key];
    }

    var _this = (0, _possibleConstructorReturn3.default)(this, _Buffer.call.apply(_Buffer, [this].concat(args)));

    _this.insideAux = false;
    _this.printAuxAfterOnNextUserNode = false;
    _this._printStack = [];
    return _this;
  }

  Printer.prototype.print = function print(node, parent) {
    var _this2 = this;

    var opts = arguments.length <= 2 || arguments[2] === undefined ? {} : arguments[2];

    if (!node) return;

    this._lastPrintedIsEmptyStatement = false;

    if (parent && parent._compact) {
      node._compact = true;
    }

    var oldInAux = this.insideAux;
    this.insideAux = !node.loc;

    var oldConcise = this.format.concise;
    if (node._compact) {
      this.format.concise = true;
    }

    var printMethod = this[node.type];
    if (!printMethod) {
      throw new ReferenceError("unknown node of type " + (0, _stringify2.default)(node.type) + " with constructor " + (0, _stringify2.default)(node && node.constructor.name));
    }

    this._printStack.push(node);

    if (node.loc) this.printAuxAfterComment();
    this.printAuxBeforeComment(oldInAux);

    var needsParens = n.needsParens(node, parent, this._printStack);
    if (needsParens) this.token("(");

    this.printLeadingComments(node, parent);

    this.catchUp(node);

    this._printNewline(true, node, parent, opts);

    if (opts.before) opts.before();

    var loc = t.isProgram(node) || t.isFile(node) ? null : node.loc;
    this.withSource("start", loc, function () {
      _this2[node.type](node, parent);
    });

    // Check again if any of our children may have left an aux comment on the stack
    if (node.loc) this.printAuxAfterComment();

    this.printTrailingComments(node, parent);

    if (needsParens) this.token(")");

    // end
    this._printStack.pop();
    if (opts.after) opts.after();

    this.format.concise = oldConcise;
    this.insideAux = oldInAux;

    this._printNewline(false, node, parent, opts);
  };

  Printer.prototype.printAuxBeforeComment = function printAuxBeforeComment(wasInAux) {
    var comment = this.format.auxiliaryCommentBefore;
    if (!wasInAux && this.insideAux && !this.printAuxAfterOnNextUserNode) {
      this.printAuxAfterOnNextUserNode = true;
      if (comment) this.printComment({
        type: "CommentBlock",
        value: comment
      });
    }
  };

  Printer.prototype.printAuxAfterComment = function printAuxAfterComment() {
    if (this.printAuxAfterOnNextUserNode) {
      this.printAuxAfterOnNextUserNode = false;
      var comment = this.format.auxiliaryCommentAfter;
      if (comment) this.printComment({
        type: "CommentBlock",
        value: comment
      });
    }
  };

  Printer.prototype.getPossibleRaw = function getPossibleRaw(node) {
    if (this.format.minified) return;

    var extra = node.extra;
    if (extra && extra.raw != null && extra.rawValue != null && node.value === extra.rawValue) {
      return extra.raw;
    }
  };

  Printer.prototype.printJoin = function printJoin(nodes, parent) {
    var _this3 = this;

    var opts = arguments.length <= 2 || arguments[2] === undefined ? {} : arguments[2];

    if (!nodes || !nodes.length) return;

    var len = nodes.length;
    var node = void 0,
        i = void 0;

    if (opts.indent) this.indent();

    var printOpts = {
      statement: opts.statement,
      addNewlines: opts.addNewlines,
      after: function after() {
        if (opts.iterator) {
          opts.iterator(node, i);
        }

        if (opts.separator && parent.loc) {
          _this3.printAuxAfterComment();
        }

        if (opts.separator && i < len - 1) {
          opts.separator.call(_this3);
        }
      }
    };

    for (i = 0; i < nodes.length; i++) {
      node = nodes[i];
      this.print(node, parent, printOpts);
    }

    if (opts.indent) this.dedent();
  };

  Printer.prototype.printAndIndentOnComments = function printAndIndentOnComments(node, parent) {
    var indent = !!node.leadingComments;
    if (indent) this.indent();
    this.print(node, parent);
    if (indent) this.dedent();
  };

  Printer.prototype.printBlock = function printBlock(parent) {
    var node = parent.body;

    if (!t.isEmptyStatement(node)) {
      this.space();
    }

    this.print(node, parent);
  };

  Printer.prototype.generateComment = function generateComment(comment) {
    var val = comment.value;
    if (comment.type === "CommentLine") {
      val = "//" + val;
    } else {
      val = "/*" + val + "*/";
    }
    return val;
  };

  Printer.prototype.printTrailingComments = function printTrailingComments(node, parent) {
    this.printComments(this.getComments(false, node, parent));
  };

  Printer.prototype.printLeadingComments = function printLeadingComments(node, parent) {
    this.printComments(this.getComments(true, node, parent));
  };

  Printer.prototype.printInnerComments = function printInnerComments(node) {
    var indent = arguments.length <= 1 || arguments[1] === undefined ? true : arguments[1];

    if (!node.innerComments) return;
    if (indent) this.indent();
    this.printComments(node.innerComments);
    if (indent) this.dedent();
  };

  Printer.prototype.printSequence = function printSequence(nodes, parent) {
    var opts = arguments.length <= 2 || arguments[2] === undefined ? {} : arguments[2];

    opts.statement = true;
    return this.printJoin(nodes, parent, opts);
  };

  Printer.prototype.printList = function printList(items, parent) {
    var opts = arguments.length <= 2 || arguments[2] === undefined ? {} : arguments[2];

    if (opts.separator == null) {
      opts.separator = commaSeparator;
    }

    return this.printJoin(items, parent, opts);
  };

  Printer.prototype._printNewline = function _printNewline(leading, node, parent, opts) {
    // Fast path since 'this.newline' does nothing when not tracking lines.
    if (this.format.retainLines || this.format.compact) return;

    if (!opts.statement && !n.isUserWhitespacable(node, parent)) {
      return;
    }

    // Fast path for concise since 'this.newline' just inserts a space when
    // concise formatting is in use.
    if (this.format.concise) {
      this.space();
      return;
    }

    var lines = 0;

    if (node.start != null && !node._ignoreUserWhitespace && this.tokens.length) {
      // user node
      if (leading) {
        lines = this.whitespace.getNewlinesBefore(node);
      } else {
        lines = this.whitespace.getNewlinesAfter(node);
      }
    } else {
      // generated node
      if (!leading) lines++; // always include at least a single line after
      if (opts.addNewlines) lines += opts.addNewlines(leading, node) || 0;

      var needs = n.needsWhitespaceAfter;
      if (leading) needs = n.needsWhitespaceBefore;
      if (needs(node, parent)) lines++;

      // generated nodes can't add starting file whitespace
      if (!this.buf) lines = 0;
    }

    this.newline(lines);
  };

  Printer.prototype.getComments = function getComments(leading, node) {
    // Note, we use a boolean flag here instead of passing in the attribute name as it is faster
    // because this is called extremely frequently.
    return node && (leading ? node.leadingComments : node.trailingComments) || [];
  };

  Printer.prototype.shouldPrintComment = function shouldPrintComment(comment) {
    if (this.format.shouldPrintComment) {
      return this.format.shouldPrintComment(comment.value);
    } else {
      if (!this.format.minified && (comment.value.indexOf("@license") >= 0 || comment.value.indexOf("@preserve") >= 0)) {
        return true;
      } else {
        return this.format.comments;
      }
    }
  };

  Printer.prototype.printComment = function printComment(comment) {
    var _this4 = this;

    if (!this.shouldPrintComment(comment)) return;

    if (comment.ignore) return;
    comment.ignore = true;

    if (comment.start != null) {
      if (this.printedCommentStarts[comment.start]) return;
      this.printedCommentStarts[comment.start] = true;
    }

    // Exclude comments from source mappings since they will only clutter things.
    this.withSource(null, null, function () {
      _this4.catchUp(comment);

      // whitespace before
      _this4.newline(_this4.whitespace.getNewlinesBefore(comment));

      if (!_this4.endsWith(["[", "{"])) _this4.space();

      var column = _this4.position.column;
      var val = _this4.generateComment(comment);

      //
      if (comment.type === "CommentBlock" && _this4.format.indent.adjustMultilineComment) {
        var offset = comment.loc && comment.loc.start.column;
        if (offset) {
          var newlineRegex = new RegExp("\\n\\s{1," + offset + "}", "g");
          val = val.replace(newlineRegex, "\n");
        }

        var indent = Math.max(_this4.indentSize(), column);
        val = val.replace(/\n/g, "\n" + (0, _repeat2.default)(" ", indent));
      }

      // force a newline for line comments when retainLines is set in case the next printed node
      // doesn't catch up
      if ((_this4.format.compact || _this4.format.concise || _this4.format.retainLines) && comment.type === "CommentLine") {
        val += "\n";
      }

      //
      _this4.push(val);

      // whitespace after
      _this4.newline(_this4.whitespace.getNewlinesAfter(comment));
    });
  };

  Printer.prototype.printComments = function printComments(comments) {
    if (!comments || !comments.length) return;

    for (var _iterator = comments, _isArray = Array.isArray(_iterator), _i = 0, _iterator = _isArray ? _iterator : (0, _getIterator3.default)(_iterator);;) {
      var _ref;

      if (_isArray) {
        if (_i >= _iterator.length) break;
        _ref = _iterator[_i++];
      } else {
        _i = _iterator.next();
        if (_i.done) break;
        _ref = _i.value;
      }

      var comment = _ref;

      this.printComment(comment);
    }
  };

  return Printer;
}(_buffer2.default);

exports.default = Printer;


function commaSeparator() {
  this.token(",");
  this.space();
}

var _arr = [require("./generators/template-literals"), require("./generators/expressions"), require("./generators/statements"), require("./generators/classes"), require("./generators/methods"), require("./generators/modules"), require("./generators/types"), require("./generators/flow"), require("./generators/base"), require("./generators/jsx")];
for (var _i2 = 0; _i2 < _arr.length; _i2++) {
  var generator = _arr[_i2];
  (0, _assign2.default)(Printer.prototype, generator);
}
module.exports = exports["default"];