"use strict";

var _getIterator = require("babel-runtime/core-js/get-iterator")["default"];

exports.__esModule = true;

exports["default"] = function (_ref2) {
  var t = _ref2.types;

  var FLOW_DIRECTIVE = "@flow";

  return {
    inherits: require("babel-plugin-syntax-flow"),

    visitor: {
      Program: function Program(path, _ref3) {
        var comments = _ref3.file.ast.comments;

        for (var _iterator = (comments /*: Array<Object>*/), _isArray = Array.isArray(_iterator), _i = 0, _iterator = _isArray ? _iterator : _getIterator(_iterator);;) {
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

          if (comment.value.indexOf(FLOW_DIRECTIVE) >= 0) {
            // remove flow directive
            comment.value = comment.value.replace(FLOW_DIRECTIVE, "");

            // remove the comment completely if it only consists of whitespace and/or stars
            if (!comment.value.replace(/\*/g, "").trim()) comment.ignore = true;
          }
        }
      },

      Flow: function Flow(path) {
        path.remove();
      },

      ClassProperty: function ClassProperty(path) {
        path.node.typeAnnotation = null;
        if (!path.node.value) path.remove();
      },

      Class: function Class(_ref4) {
        var node = _ref4.node;

        node["implements"] = null;
      },

      Function: function Function(_ref5) {
        var node = _ref5.node;

        for (var i = 0; i < node.params.length; i++) {
          var param = node.params[i];
          param.optional = false;
        }
      },

      TypeCastExpression: function TypeCastExpression(path) {
        var node = path.node;

        do {
          node = node.expression;
        } while (t.isTypeCastExpression(node));
        path.replaceWith(node);
      }
    }
  };
};

module.exports = exports["default"];