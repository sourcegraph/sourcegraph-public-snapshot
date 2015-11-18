"use strict";

var _getIterator = require("babel-runtime/core-js/get-iterator")["default"];

exports.__esModule = true;

exports["default"] = function (_ref2) {
  var messages = _ref2.messages;

  return {
    visitor: {
      Scope: function Scope(_ref3) {
        var scope = _ref3.scope;

        for (var _name in scope.bindings) {
          var binding = scope.bindings[_name];
          if (binding.kind !== "const" && binding.kind !== "module") continue;

          for (var _iterator = (binding.constantViolations /*: Array*/), _isArray = Array.isArray(_iterator), _i = 0, _iterator = _isArray ? _iterator : _getIterator(_iterator);;) {
            var _ref;

            if (_isArray) {
              if (_i >= _iterator.length) break;
              _ref = _iterator[_i++];
            } else {
              _i = _iterator.next();
              if (_i.done) break;
              _ref = _i.value;
            }

            var violation = _ref;

            throw violation.buildCodeFrameError(messages.get("readOnly", _name));
          }
        }
      }
    }
  };
};

module.exports = exports["default"];