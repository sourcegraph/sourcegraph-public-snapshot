/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayQueryTracker
 * 
 */

'use strict';

var _classCallCheck3 = _interopRequireDefault(require('babel-runtime/helpers/classCallCheck'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

var TYPE = '__type__';

var RelayQueryTracker = function () {
  function RelayQueryTracker() {
    (0, _classCallCheck3['default'])(this, RelayQueryTracker);

    this._trackedNodesByID = {};
  }

  RelayQueryTracker.prototype.trackNodeForID = function trackNodeForID(node, dataID) {
    // Don't track `__type__` fields
    if (node instanceof require('./RelayQuery').Field && node.getSchemaName() === TYPE) {
      return;
    }

    this._trackedNodesByID[dataID] = this._trackedNodesByID[dataID] || {
      trackedNodes: [],
      isMerged: false
    };
    this._trackedNodesByID[dataID].trackedNodes.push(node);
    this._trackedNodesByID[dataID].isMerged = false;
  };

  /**
   * Get the children that are tracked for the given `dataID`, if any.
   */


  RelayQueryTracker.prototype.getTrackedChildrenForID = function getTrackedChildrenForID(dataID) {
    var trackedNodesByID = this._trackedNodesByID[dataID];
    if (!trackedNodesByID) {
      return [];
    }
    var isMerged = trackedNodesByID.isMerged;
    var trackedNodes = trackedNodesByID.trackedNodes;

    if (!isMerged) {
      (function () {
        var trackedChildren = [];
        trackedNodes.forEach(function (trackedQuery) {
          trackedChildren.push.apply(trackedChildren, trackedQuery.getChildren());
        });
        trackedNodes.length = 0;
        trackedNodesByID.isMerged = true;
        var containerNode = require('./RelayQuery').Fragment.build('RelayQueryTracker', require('./RelayNodeInterface').NODE_TYPE, trackedChildren);
        containerNode = require('./flattenRelayQuery')(containerNode);
        if (containerNode) {
          trackedNodes.push(containerNode);
        }
      })();
    }
    var trackedNode = trackedNodes[0];
    if (trackedNode) {
      return trackedNode.getChildren();
    }
    return [];
  };

  /**
   * Removes all nodes that are tracking the given DataID from the
   * query-tracker.
   */


  RelayQueryTracker.prototype.untrackNodesForID = function untrackNodesForID(dataID) {
    delete this._trackedNodesByID[dataID];
  };

  return RelayQueryTracker;
}();

module.exports = RelayQueryTracker;