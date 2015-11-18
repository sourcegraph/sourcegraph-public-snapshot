"use strict";

var _getIterator = require("babel-runtime/core-js/get-iterator")["default"];

exports.__esModule = true;

exports["default"] = function (_ref4) {
  var t = _ref4.types;

  function isString(node) {
    return t.isLiteral(node) && typeof node.value === "string";
  }

  function buildBinaryExpression(left, right) {
    return t.binaryExpression("+", left, right);
  }

  return {
    visitor: {
      TaggedTemplateExpression: function TaggedTemplateExpression(path, state) {
        var node = path.node;

        var quasi = node.quasi;
        var args = [];

        var strings = [];
        var raw = [];

        for (var _iterator = (quasi.quasis /*: Array*/), _isArray = Array.isArray(_iterator), _i = 0, _iterator = _isArray ? _iterator : _getIterator(_iterator);;) {
          var _ref;

          if (_isArray) {
            if (_i >= _iterator.length) break;
            _ref = _iterator[_i++];
          } else {
            _i = _iterator.next();
            if (_i.done) break;
            _ref = _i.value;
          }

          var elem = _ref;

          strings.push(t.stringLiteral(elem.value.cooked));
          raw.push(t.stringLiteral(elem.value.raw));
        }

        strings = t.arrayExpression(strings);
        raw = t.arrayExpression(raw);

        var templateName = "taggedTemplateLiteral";
        if (state.opts.loose) templateName += "Loose";

        var templateObject = state.file.addTemplateObject(templateName, strings, raw);
        args.push(templateObject);

        args = args.concat(quasi.expressions);

        path.replaceWith(t.callExpression(node.tag, args));
      },

      TemplateLiteral: function TemplateLiteral(path, state) {
        var nodes /*: Array<Object>*/ = [];

        var expressions = path.get("expressions");

        for (var _iterator2 = (path.node.quasis /*: Array*/), _isArray2 = Array.isArray(_iterator2), _i2 = 0, _iterator2 = _isArray2 ? _iterator2 : _getIterator(_iterator2);;) {
          var _ref2;

          if (_isArray2) {
            if (_i2 >= _iterator2.length) break;
            _ref2 = _iterator2[_i2++];
          } else {
            _i2 = _iterator2.next();
            if (_i2.done) break;
            _ref2 = _i2.value;
          }

          var elem = _ref2;

          nodes.push(t.stringLiteral(elem.value.cooked));

          var expr = expressions.shift();
          if (expr) {
            if (state.opts.spec && !expr.isBaseType("string") && !expr.isBaseType("number")) {
              nodes.push(t.callExpression(t.identifier("String"), [expr.node]));
            } else {
              nodes.push(expr.node);
            }
          }
        }

        // filter out empty string literals
        nodes = nodes.filter(function (n) {
          return !t.isLiteral(n, { value: "" });
        });

        // since `+` is left-to-right associative
        // ensure the first node is a string if first/second isn't
        if (!isString(nodes[0]) && !isString(nodes[1])) {
          nodes.unshift(t.stringLiteral(""));
        }

        if (nodes.length > 1) {
          var root = buildBinaryExpression(nodes.shift(), nodes.shift());

          for (var _iterator3 = nodes, _isArray3 = Array.isArray(_iterator3), _i3 = 0, _iterator3 = _isArray3 ? _iterator3 : _getIterator(_iterator3);;) {
            var _ref3;

            if (_isArray3) {
              if (_i3 >= _iterator3.length) break;
              _ref3 = _iterator3[_i3++];
            } else {
              _i3 = _iterator3.next();
              if (_i3.done) break;
              _ref3 = _i3.value;
            }

            var node = _ref3;

            root = buildBinaryExpression(root, node);
          }

          path.replaceWith(root);
        } else {
          path.replaceWith(nodes[0]);
        }
      }
    }
  };
};

module.exports = exports["default"];