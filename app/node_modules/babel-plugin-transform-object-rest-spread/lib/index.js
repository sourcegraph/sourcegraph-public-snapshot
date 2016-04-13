"use strict";

exports.__esModule = true;

exports["default"] = function (_ref) {
  var t = _ref.types;

  function hasSpread(node) {
    var _arr = node.properties;

    for (var _i = 0; _i < _arr.length; _i++) {
      var prop = _arr[_i];
      if (t.isSpreadProperty(prop)) {
        return true;
      }
    }
    return false;
  }

  return {
    inherits: require("babel-plugin-syntax-object-rest-spread"),

    visitor: {
      ObjectExpression: function ObjectExpression(path, file) {
        if (!hasSpread(path.node)) return;

        var args = [];
        var props = [];

        function push() {
          if (!props.length) return;
          args.push(t.objectExpression(props));
          props = [];
        }

        var _arr2 = path.node.properties;
        for (var _i2 = 0; _i2 < _arr2.length; _i2++) {
          var prop = _arr2[_i2];
          if (t.isSpreadProperty(prop)) {
            push();
            args.push(prop.argument);
          } else {
            props.push(prop);
          }
        }

        push();

        if (!t.isObjectExpression(args[0])) {
          args.unshift(t.objectExpression([]));
        }

        path.replaceWith(t.callExpression(file.addHelper("extends"), args));
      }
    }
  };
};

module.exports = exports["default"];