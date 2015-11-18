"use strict";

var _getIterator = require("babel-runtime/core-js/get-iterator")["default"];

exports.__esModule = true;

exports["default"] = function (_ref5) {
  var t = _ref5.types;
  var template = _ref5.template;

  var buildMutatorMapAssign = template("\n    MUTATOR_MAP_REF[KEY] = MUTATOR_MAP_REF[KEY] || {};\n    MUTATOR_MAP_REF[KEY].KIND = VALUE;\n  ");

  function getValue(prop) {
    if (t.isObjectProperty(prop)) {
      return prop.value;
    } else if (t.isObjectMethod(prop)) {
      return t.functionExpression(null, prop.params, prop.body, prop.generator, prop.async);
    }
  }

  function pushAssign(objId, prop, body) {
    if (prop.kind === "get" && prop.kind === "set") {
      pushMutatorDefine(objId, prop, body);
    } else {
      body.push(t.expressionStatement(t.assignmentExpression("=", t.memberExpression(objId, prop.key, prop.computed || t.isLiteral(prop.key)), getValue(prop))));
    }
  }

  function pushMutatorDefine(_ref6, prop) {
    var objId = _ref6.objId;
    var body = _ref6.body;
    var getMutatorId = _ref6.getMutatorId;
    var scope = _ref6.scope;

    var key = prop.key;

    var maybeMemoise = scope.maybeGenerateMemoised(prop.key);
    if (maybeMemoise) {
      body.push(t.expressionStatement(t.assignmentExpression("=", maybeMemoise, key)));
      key = maybeMemoise;
    }

    body.push.apply(body, buildMutatorMapAssign({
      MUTATOR_MAP_REF: getMutatorId(),
      KEY: key,
      VALUE: getValue(prop),
      KIND: t.identifier(prop.kind)
    }));
  }

  function loose(info) {
    for (var _iterator = info.computedProps, _isArray = Array.isArray(_iterator), _i = 0, _iterator = _isArray ? _iterator : _getIterator(_iterator);;) {
      var _ref;

      if (_isArray) {
        if (_i >= _iterator.length) break;
        _ref = _iterator[_i++];
      } else {
        _i = _iterator.next();
        if (_i.done) break;
        _ref = _i.value;
      }

      var prop = _ref;

      if (prop.kind === "get" || prop.kind === "set") {
        pushMutatorDefine(info, prop);
      } else {
        pushAssign(info.objId, prop, info.body);
      }
    }
  }

  function spec(info) {
    var objId = info.objId;
    var body = info.body;
    var computedProps = info.computedProps;
    var state = info.state;

    for (var _iterator2 = computedProps, _isArray2 = Array.isArray(_iterator2), _i2 = 0, _iterator2 = _isArray2 ? _iterator2 : _getIterator(_iterator2);;) {
      var _ref2;

      if (_isArray2) {
        if (_i2 >= _iterator2.length) break;
        _ref2 = _iterator2[_i2++];
      } else {
        _i2 = _iterator2.next();
        if (_i2.done) break;
        _ref2 = _i2.value;
      }

      var prop = _ref2;

      var key = t.toComputedKey(prop);

      if (prop.kind === "get" || prop.kind === "set") {
        pushMutatorDefine(info, prop);
      } else if (t.isStringLiteral(key, { value: "__proto__" })) {
        pushAssign(objId, prop, body);
      } else {
        if (computedProps.length === 1) {
          return t.callExpression(state.addHelper("defineProperty"), [info.initPropExpression, key, getValue(prop)]);
        } else {
          body.push(t.expressionStatement(t.callExpression(state.addHelper("defineProperty"), [objId, key, getValue(prop)])));
        }
      }
    }
  }

  return {
    visitor: {
      ObjectExpression: {
        exit: function exit(path, state) {
          var node = path.node;
          var parent = path.parent;
          var scope = path.scope;

          var hasComputed = false;
          for (var _iterator3 = (node.properties /*: Array<Object>*/), _isArray3 = Array.isArray(_iterator3), _i3 = 0, _iterator3 = _isArray3 ? _iterator3 : _getIterator(_iterator3);;) {
            var _ref3;

            if (_isArray3) {
              if (_i3 >= _iterator3.length) break;
              _ref3 = _iterator3[_i3++];
            } else {
              _i3 = _iterator3.next();
              if (_i3.done) break;
              _ref3 = _i3.value;
            }

            var prop = _ref3;

            hasComputed = prop.computed === true;
            if (hasComputed) break;
          }
          if (!hasComputed) return;

          // put all getters/setters into the first object expression as well as all initialisers up
          // to the first computed property

          var initProps = [];
          var computedProps = [];
          var foundComputed = false;

          for (var _iterator4 = node.properties, _isArray4 = Array.isArray(_iterator4), _i4 = 0, _iterator4 = _isArray4 ? _iterator4 : _getIterator(_iterator4);;) {
            var _ref4;

            if (_isArray4) {
              if (_i4 >= _iterator4.length) break;
              _ref4 = _iterator4[_i4++];
            } else {
              _i4 = _iterator4.next();
              if (_i4.done) break;
              _ref4 = _i4.value;
            }

            var prop = _ref4;

            if (prop.computed) {
              foundComputed = true;
            }

            if (foundComputed) {
              computedProps.push(prop);
            } else {
              initProps.push(prop);
            }
          }

          var objId = scope.generateUidIdentifierBasedOnNode(parent);
          var initPropExpression = t.objectExpression(initProps);
          var body = [];

          body.push(t.variableDeclaration("var", [t.variableDeclarator(objId, initPropExpression)]));

          var callback = spec;
          if (state.opts.loose) callback = loose;

          var mutatorRef = undefined;

          var getMutatorId = function getMutatorId() {
            if (!mutatorRef) {
              mutatorRef = scope.generateUidIdentifier("mutatorMap");

              body.push(t.variableDeclaration("var", [t.variableDeclarator(mutatorRef, t.objectExpression([]))]));
            }

            return mutatorRef;
          };

          var single = callback({
            scope: scope,
            objId: objId,
            body: body,
            computedProps: computedProps,
            initPropExpression: initPropExpression,
            getMutatorId: getMutatorId,
            state: state
          });

          if (mutatorRef) {
            body.push(t.expressionStatement(t.callExpression(state.addHelper("defineEnumerableProperties"), [objId, mutatorRef])));
          }

          if (single) {
            path.replaceWith(single);
          } else {
            body.push(t.expressionStatement(objId));
            path.replaceWithMultiple(body);
          }
        }
      }
    }
  };
};

module.exports = exports["default"];