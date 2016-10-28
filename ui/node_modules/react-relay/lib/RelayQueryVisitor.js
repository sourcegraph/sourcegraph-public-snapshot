/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayQueryVisitor
 * 
 */

'use strict';

var _classCallCheck3 = _interopRequireDefault(require('babel-runtime/helpers/classCallCheck'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

/**
 * @internal
 *
 * Base class for traversing a Relay Query.
 *
 * Subclasses can optionally implement methods to customize the traversal:
 *
 * - `visitField(field, state)`: Called for each field.
 * - `visitFragment(fragment, state)`: Called for each fragment.
 * - `visitQuery(fragment, state)`: Called for the top level query.
 *
 * A `state` variable is passed along to all callbacks and can be used to
 * accumulate data while traversing (effectively passing data back up the tree),
 * or modify the behavior of later callbacks (effectively passing data down the
 * tree).
 *
 * There are two additional methods for controlling the traversal:
 *
 * - `traverse(parent, state)`: Visits all children of `parent`. Subclasses
 *   may override in order to short-circuit traversal. Note that
 *   `visit{Field,Fragment,Query}` are //not// called on `parent`, as it will
 *   already have been visited by the time this method is called.
 * - `visit(child, state)`: Processes the `child` node, calling the appropriate
 *   `visit{Field,Fragment,Query}` method based on the node type.
 *
 * By convention, each of the callback methods returns the visited node. This is
 * used by the `RelayQueryTransform` subclass to implement mapping and filtering
 * behavior, but purely-visitor subclases do not need to follow this convention.
 *
 * @see RelayQueryTransform
 */

var RelayQueryVisitor = function () {
  function RelayQueryVisitor() {
    (0, _classCallCheck3['default'])(this, RelayQueryVisitor);
  }

  RelayQueryVisitor.prototype.visit = function visit(node, nextState) {
    if (node instanceof require('./RelayQuery').Field) {
      return this.visitField(node, nextState);
    } else if (node instanceof require('./RelayQuery').Fragment) {
      return this.visitFragment(node, nextState);
    } else if (node instanceof require('./RelayQuery').Root) {
      return this.visitRoot(node, nextState);
    }
  };

  RelayQueryVisitor.prototype.traverse = function traverse(node, nextState) {
    if (node.canHaveSubselections()) {
      this.traverseChildren(node, nextState, function (child) {
        this.visit(child, nextState);
      }, this);
    }
    return node;
  };

  RelayQueryVisitor.prototype.traverseChildren = function traverseChildren(node, nextState, callback, context) {
    var children = node.getChildren();
    for (var _index = 0; _index < children.length; _index++) {
      callback.call(context, children[_index], _index, children);
    }
  };

  RelayQueryVisitor.prototype.visitField = function visitField(node, nextState) {
    return this.traverse(node, nextState);
  };

  RelayQueryVisitor.prototype.visitFragment = function visitFragment(node, nextState) {
    return this.traverse(node, nextState);
  };

  RelayQueryVisitor.prototype.visitRoot = function visitRoot(node, nextState) {
    return this.traverse(node, nextState);
  };

  return RelayQueryVisitor;
}();

module.exports = RelayQueryVisitor;