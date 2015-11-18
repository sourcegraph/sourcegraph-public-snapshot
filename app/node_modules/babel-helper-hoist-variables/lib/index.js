"use strict";

var _getIterator = require("babel-runtime/core-js/get-iterator")["default"];

var _interopRequireWildcard = require("babel-runtime/helpers/interop-require-wildcard")["default"];

exports.__esModule = true;

var _babelTypes = require("babel-types");

var t = _interopRequireWildcard(_babelTypes);

var visitor = {
  Scope: function Scope(path, state) {
    if (state.kind === "let") path.skip();
  },

  Function: function Function(path) {
    path.skip();
  },

  VariableDeclaration: function VariableDeclaration(path, state) {
    if (state.kind && path.node.kind !== state.kind) return;

    var nodes = [];

    var declarations /*: Array<Object>*/ = path.get("declarations");
    var firstId = undefined;

    for (var _iterator = declarations, _isArray = Array.isArray(_iterator), _i = 0, _iterator = _isArray ? _iterator : _getIterator(_iterator);;) {
      var _ref;

      if (_isArray) {
        if (_i >= _iterator.length) break;
        _ref = _iterator[_i++];
      } else {
        _i = _iterator.next();
        if (_i.done) break;
        _ref = _i.value;
      }

      var declar = _ref;

      firstId = declar.node.id;

      if (declar.node.init) {
        nodes.push(t.expressionStatement(t.assignmentExpression("=", declar.node.id, declar.node.init)));
      }

      for (var _name in declar.getBindingIdentifiers()) {
        state.emit(t.identifier(_name), _name);
      }
    }

    // for (var i in test)
    if (path.parentPath.isFor({ left: path.node })) {
      path.replaceWith(firstId);
    } else {
      path.replaceWithMultiple(nodes);
    }
  }
};

exports["default"] = function (path, emit /*: Function*/) {
  var kind /*: "var" | "let"*/ = arguments.length <= 2 || arguments[2] === undefined ? "var" : arguments[2];

  path.traverse(visitor, { kind: kind, emit: emit });
};

module.exports = exports["default"];