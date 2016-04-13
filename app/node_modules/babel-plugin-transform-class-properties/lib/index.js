/* eslint max-len: 0 */
"use strict";

var _getIterator = require("babel-runtime/core-js/get-iterator")["default"];

exports.__esModule = true;

exports["default"] = function (_ref2) {
  var t = _ref2.types;

  var findBareSupers = {
    Super: function Super(path) {
      if (path.parentPath.isCallExpression({ callee: path.node })) {
        this.push(path.parentPath);
      }
    }
  };

  var referenceVisitor = {
    ReferencedIdentifier: function ReferencedIdentifier(path) {
      if (this.scope.hasOwnBinding(path.node.name)) {
        this.collision = true;
        path.skip();
      }
    }
  };

  return {
    inherits: require("babel-plugin-syntax-class-properties"),

    visitor: {
      Class: function Class(path) {
        var isDerived = !!path.node.superClass;
        var constructor = undefined;
        var props = [];
        var body = path.get("body");

        for (var _iterator = body.get("body"), _isArray = Array.isArray(_iterator), _i = 0, _iterator = _isArray ? _iterator : _getIterator(_iterator);;) {
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

          if (_path.isClassProperty()) {
            props.push(_path);
          } else if (_path.isClassMethod({ kind: "constructor" })) {
            constructor = _path;
          }
        }

        if (!props.length) return;

        var nodes = [];
        var ref = undefined;

        if (path.isClassExpression() || !path.node.id) {
          ref = path.scope.generateUidIdentifier("class");
        } else {
          // path.isClassDeclaration() && path.node.id
          ref = path.node.id;
        }

        var instanceBody = [];

        for (var _i2 = 0; _i2 < props.length; _i2++) {
          var prop = props[_i2];
          var propNode = prop.node;
          if (propNode.decorators && propNode.decorators.length > 0) continue;
          if (!propNode.value) continue;

          var isStatic = propNode["static"];

          if (isStatic) {
            nodes.push(t.expressionStatement(t.assignmentExpression("=", t.memberExpression(ref, propNode.key), propNode.value)));
          } else {
            instanceBody.push(t.expressionStatement(t.assignmentExpression("=", t.memberExpression(t.thisExpression(), propNode.key), propNode.value)));
          }
        }

        if (instanceBody.length) {
          if (!constructor) {
            var newConstructor = t.classMethod("constructor", t.identifier("constructor"), [], t.blockStatement([]));
            if (isDerived) {
              newConstructor.params = [t.restElement(t.identifier("args"))];
              newConstructor.body.body.push(t.returnStatement(t.callExpression(t["super"](), [t.spreadElement(t.identifier("args"))])));
            }

            var _body$unshiftContainer = body.unshiftContainer("body", newConstructor);

            constructor = _body$unshiftContainer[0];
          }

          var collisionState = {
            collision: false,
            scope: constructor.scope
          };

          for (var _i3 = 0; _i3 < props.length; _i3++) {
            var prop = props[_i3];
            prop.traverse(referenceVisitor, collisionState);
            if (collisionState.collision) break;
          }

          if (collisionState.collision) {
            var initialisePropsRef = path.scope.generateUidIdentifier("initialiseProps");

            nodes.push(t.variableDeclaration("var", [t.variableDeclarator(initialisePropsRef, t.functionExpression(null, [], t.blockStatement(instanceBody)))]));

            instanceBody = [t.expressionStatement(t.callExpression(t.memberExpression(initialisePropsRef, t.identifier("call")), [t.thisExpression()]))];
          }

          //

          if (isDerived) {
            var bareSupers = [];
            constructor.traverse(findBareSupers, bareSupers);
            for (var _i4 = 0; _i4 < bareSupers.length; _i4++) {
              var bareSuper = bareSupers[_i4];
              bareSuper.insertAfter(instanceBody);
            }
          } else {
            constructor.get("body").unshiftContainer("body", instanceBody);
          }
        }

        for (var _i5 = 0; _i5 < props.length; _i5++) {
          var prop = props[_i5];
          prop.remove();
        }

        if (!nodes.length) return;

        if (path.isClassExpression()) {
          path.scope.push({ id: ref });
          path.replaceWith(t.assignmentExpression("=", ref, path.node));
        } else {
          // path.isClassDeclaration()
          if (!path.node.id) {
            path.node.id = ref;
          }

          if (path.parentPath.isExportDeclaration()) {
            path = path.parentPath;
          }
        }

        path.insertAfter(nodes);
      }
    }
  };
};

module.exports = exports["default"];
// todo: define instead of assign