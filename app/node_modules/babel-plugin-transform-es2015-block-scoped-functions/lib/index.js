"use strict";

var _getIterator = require("babel-runtime/core-js/get-iterator")["default"];

exports.__esModule = true;

exports["default"] = function (_ref2) {
  var t = _ref2.types;

  function statementList(key, path) {
    var paths /*: Array*/ = path.get(key);

    for (var _iterator = paths, _isArray = Array.isArray(_iterator), _i = 0, _iterator = _isArray ? _iterator : _getIterator(_iterator);;) {
      var _ref;

      if (_isArray) {
        if (_i >= _iterator.length) break;
        _ref = _iterator[_i++];
      } else {
        _i = _iterator.next();
        if (_i.done) break;
        _ref = _i.value;
      }

      var _path = _ref;

      var func = _path.node;
      if (!_path.isFunctionDeclaration()) continue;

      var declar = t.variableDeclaration("let", [t.variableDeclarator(func.id, t.toExpression(func))]);

      // hoist it up above everything else
      declar._blockHoist = 2;

      // todo: name this
      func.id = null;

      _path.replaceWith(declar);
    }
  }

  return {
    visitor: {
      BlockStatement: function BlockStatement(path) {
        var node = path.node;
        var parent = path.parent;

        if (t.isFunction(parent, { body: node }) || t.isExportDeclaration(parent)) {
          return;
        }

        statementList("body", path);
      },

      SwitchCase: function SwitchCase(path) {
        statementList("consequent", path);
      }
    }
  };
};

module.exports = exports["default"];