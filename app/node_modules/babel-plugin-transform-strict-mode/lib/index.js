"use strict";

var _getIterator = require("babel-runtime/core-js/get-iterator")["default"];

var _interopRequireWildcard = require("babel-runtime/helpers/interop-require-wildcard")["default"];

exports.__esModule = true;

var _babelTypes = require("babel-types");

var t = _interopRequireWildcard(_babelTypes);

exports["default"] = function () {
  return {
    visitor: {
      Program: function Program(path, state) {
        if (state.opts.strict === false) return;

        var node = path.node;

        for (var _iterator = (node.directives /*: Array<Object>*/), _isArray = Array.isArray(_iterator), _i = 0, _iterator = _isArray ? _iterator : _getIterator(_iterator);;) {
          var _ref;

          if (_isArray) {
            if (_i >= _iterator.length) break;
            _ref = _iterator[_i++];
          } else {
            _i = _iterator.next();
            if (_i.done) break;
            _ref = _i.value;
          }

          var directive = _ref;

          if (directive.value.value === "use strict") return;
        }

        path.unshiftContainer("directives", t.directive(t.directiveLiteral("use strict")));
      }
    }
  };
};

module.exports = exports["default"];