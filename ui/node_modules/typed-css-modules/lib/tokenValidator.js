'use strict';

Object.defineProperty(exports, "__esModule", {
  value: true
});

var _createClass = function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; }();

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

var RESERVED_WORDS = ['break', 'case', 'catch', 'class', 'const', 'continue', 'debugger', 'default', 'delete', 'do', 'else', 'enum', 'export', 'extends', 'false', 'finally', 'for', 'function', 'if', 'import', 'in', 'instanceof', 'new', 'null', 'return', 'super', 'switch', 'this', 'throw', 'true', 'try', 'typeof', 'var', 'void', 'while', 'with', 'as', 'implements', 'interface', 'let', 'package', 'private', 'protected', 'public', 'static', 'yield'];

var TokenValidator = exports.TokenValidator = function () {
  function TokenValidator() {
    _classCallCheck(this, TokenValidator);
  }

  _createClass(TokenValidator, [{
    key: 'validate',
    value: function validate(key) {
      if (!key) {
        return {
          isValid: false,
          message: 'empty token'
        };
      }
      if (!/^[$_a-zA-Z][0-9a-zA-Z$_]*$/.test(key)) {
        return {
          isValid: false,
          message: key + ' is not valid TypeScript variable name.'
        };
      }
      if (RESERVED_WORDS.some(function (w) {
        return w === key;
      })) {
        return {
          isValid: false,
          message: key + ' is TypeScript reserved word.'
        };
      }
      return {
        isValid: true
      };
    }
  }]);

  return TokenValidator;
}();