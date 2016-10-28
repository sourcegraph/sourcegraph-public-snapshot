/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule intersectRelayQuery
 * 
 */

'use strict';

var _classCallCheck3 = _interopRequireDefault(require('babel-runtime/helpers/classCallCheck'));

var _possibleConstructorReturn3 = _interopRequireDefault(require('babel-runtime/helpers/possibleConstructorReturn'));

var _inherits3 = _interopRequireDefault(require('babel-runtime/helpers/inherits'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

/**
 * @internal
 *
 * `intersectRelayQuery(subjectNode, patternNode)` returns a node with fields in
 * `subjectNode` that also exist in `patternNode`. `patternNode` is expected to
 * be flattened (and not contain fragments).
 *
 * If any field in `patternNode` is unterminated (i.e. has no sub-fields), we
 * treat the field as though it contains every descendant sub-field.
 *
 * If `filterUnterminatedRange` is supplied, it will be invoked with any fields
 * from `subjectNode` that are connections and unterminated in `patternNode`. If
 * it returns true, the `edges` and `page_info` fields will be filtered out.
 */
function intersectRelayQuery(subjectNode, patternNode, filterUnterminatedRange) {
  filterUnterminatedRange = filterUnterminatedRange || returnsFalse;
  var visitor = new RelayQueryIntersector(filterUnterminatedRange);
  return visitor.traverse(subjectNode, patternNode);
}

var RelayQueryIntersector = function (_RelayQueryTransform) {
  (0, _inherits3['default'])(RelayQueryIntersector, _RelayQueryTransform);

  function RelayQueryIntersector(filterUnterminatedRange) {
    (0, _classCallCheck3['default'])(this, RelayQueryIntersector);

    var _this = (0, _possibleConstructorReturn3['default'])(this, _RelayQueryTransform.call(this));

    _this._filterUnterminatedRange = filterUnterminatedRange;
    return _this;
  }

  RelayQueryIntersector.prototype.traverse = function traverse(subjectNode, patternNode) {
    var _this2 = this;

    if (!subjectNode.canHaveSubselections()) {
      // Since `patternNode` exists, `subjectNode` must be in the intersection.
      return subjectNode;
    }
    if (!hasChildren(patternNode)) {
      if (subjectNode instanceof require('./RelayQuery').Field && subjectNode.isConnection() && this._filterUnterminatedRange(subjectNode)) {
        return filterRangeFields(subjectNode);
      }
      // Unterminated `patternNode` is the same as containing every descendant
      // sub-field, so `subjectNode` must be in the intersection.
      return subjectNode;
    }
    return subjectNode.clone(subjectNode.getChildren().map(function (subjectChild) {
      if (subjectChild instanceof require('./RelayQuery').Fragment) {
        return _this2.visit(subjectChild, patternNode);
      }
      if (subjectChild instanceof require('./RelayQuery').Field) {
        var schemaName = subjectChild.getSchemaName();
        var patternChild = void 0;
        var patternChildren = patternNode.getChildren();
        for (var ii = 0; ii < patternChildren.length; ii++) {
          var child = patternChildren[ii];
          require('fbjs/lib/invariant')(child instanceof require('./RelayQuery').Field, 'intersectRelayQuery(): Nodes in `patternNode` must be fields.');
          if (child.getSchemaName() === schemaName) {
            patternChild = child;
            break;
          }
        }
        if (patternChild) {
          return _this2.visit(subjectChild, patternChild);
        }
      }
      return null;
    }));
  };

  return RelayQueryIntersector;
}(require('./RelayQueryTransform'));

/**
 * @private
 */


var RelayQueryRangeFilter = function (_RelayQueryTransform2) {
  (0, _inherits3['default'])(RelayQueryRangeFilter, _RelayQueryTransform2);

  function RelayQueryRangeFilter() {
    (0, _classCallCheck3['default'])(this, RelayQueryRangeFilter);
    return (0, _possibleConstructorReturn3['default'])(this, _RelayQueryTransform2.apply(this, arguments));
  }

  RelayQueryRangeFilter.prototype.visitField = function visitField(node) {
    var schemaName = node.getSchemaName();
    if (schemaName === require('./RelayConnectionInterface').EDGES || schemaName === require('./RelayConnectionInterface').PAGE_INFO) {
      return null;
    } else {
      return node;
    }
  };

  return RelayQueryRangeFilter;
}(require('./RelayQueryTransform'));

var rangeFilter = new RelayQueryRangeFilter();
function filterRangeFields(node) {
  return rangeFilter.traverse(node, undefined);
}

function returnsFalse() {
  return false;
}

function hasChildren(node) {
  return !node.getChildren().every(isGenerated);
}

function isGenerated(node) {
  return node.isGenerated();
}

module.exports = intersectRelayQuery;