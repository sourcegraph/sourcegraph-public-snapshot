/* @flow */

"use strict";

var _getIterator = require("babel-runtime/core-js/get-iterator")["default"];

exports.__esModule = true;
exports.JSXAttribute = JSXAttribute;
exports.JSXIdentifier = JSXIdentifier;
exports.JSXNamespacedName = JSXNamespacedName;
exports.JSXMemberExpression = JSXMemberExpression;
exports.JSXSpreadAttribute = JSXSpreadAttribute;
exports.JSXExpressionContainer = JSXExpressionContainer;
exports.JSXText = JSXText;
exports.JSXElement = JSXElement;
exports.JSXOpeningElement = JSXOpeningElement;
exports.JSXClosingElement = JSXClosingElement;
exports.JSXEmptyExpression = JSXEmptyExpression;

function JSXAttribute(node /*: Object*/) {
  this.print(node.name, node);
  if (node.value) {
    this.push("=");
    this.print(node.value, node);
  }
}

function JSXIdentifier(node /*: Object*/) {
  this.push(node.name);
}

function JSXNamespacedName(node /*: Object*/) {
  this.print(node.namespace, node);
  this.push(":");
  this.print(node.name, node);
}

function JSXMemberExpression(node /*: Object*/) {
  this.print(node.object, node);
  this.push(".");
  this.print(node.property, node);
}

function JSXSpreadAttribute(node /*: Object*/) {
  this.push("{...");
  this.print(node.argument, node);
  this.push("}");
}

function JSXExpressionContainer(node /*: Object*/) {
  this.push("{");
  this.print(node.expression, node);
  this.push("}");
}

function JSXText(node /*: Object*/) {
  this.push(node.value, true);
}

function JSXElement(node /*: Object*/) {
  var open = node.openingElement;
  this.print(open, node);
  if (open.selfClosing) return;

  this.indent();
  for (var _iterator = (node.children /*: Array<Object>*/), _isArray = Array.isArray(_iterator), _i = 0, _iterator = _isArray ? _iterator : _getIterator(_iterator);;) {
    var _ref;

    if (_isArray) {
      if (_i >= _iterator.length) break;
      _ref = _iterator[_i++];
    } else {
      _i = _iterator.next();
      if (_i.done) break;
      _ref = _i.value;
    }

    var child = _ref;

    this.print(child, node);
  }
  this.dedent();

  this.print(node.closingElement, node);
}

function JSXOpeningElement(node /*: Object*/) {
  this.push("<");
  this.print(node.name, node);
  if (node.attributes.length > 0) {
    this.push(" ");
    this.printJoin(node.attributes, node, { separator: " " });
  }
  this.push(node.selfClosing ? " />" : ">");
}

function JSXClosingElement(node /*: Object*/) {
  this.push("</");
  this.print(node.name, node);
  this.push(">");
}

function JSXEmptyExpression() {}