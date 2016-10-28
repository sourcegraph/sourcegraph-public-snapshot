/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule checkRelayQueryData
 * 
 */

'use strict';

var _classCallCheck3 = _interopRequireDefault(require('babel-runtime/helpers/classCallCheck'));

var _possibleConstructorReturn3 = _interopRequireDefault(require('babel-runtime/helpers/possibleConstructorReturn'));

var _inherits3 = _interopRequireDefault(require('babel-runtime/helpers/inherits'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

var EDGES = require('./RelayConnectionInterface').EDGES;

var PAGE_INFO = require('./RelayConnectionInterface').PAGE_INFO;

/**
 * @internal
 *
 * Traverses a query and data in the record store to determine whether we have
 * enough data to satisfy the query.
 */


function checkRelayQueryData(store, query) {

  var checker = new RelayQueryChecker(store);

  var state = {
    dataID: undefined,
    rangeInfo: undefined,
    result: true
  };

  checker.visit(query, state);
  return state.result;
}

var RelayQueryChecker = function (_RelayQueryVisitor) {
  (0, _inherits3['default'])(RelayQueryChecker, _RelayQueryVisitor);

  function RelayQueryChecker(store) {
    (0, _classCallCheck3['default'])(this, RelayQueryChecker);

    var _this = (0, _possibleConstructorReturn3['default'])(this, _RelayQueryVisitor.call(this));

    _this._store = store;
    return _this;
  }

  /**
   * Skip visiting children if result is already false.
   */


  RelayQueryChecker.prototype.traverse = function traverse(node, state) {
    var children = node.getChildren();
    for (var ii = 0; ii < children.length; ii++) {
      if (!state.result) {
        return;
      }
      this.visit(children[ii], state);
    }
  };

  RelayQueryChecker.prototype.visitRoot = function visitRoot(root, state) {
    var _this2 = this;

    var storageKey = root.getStorageKey();
    require('./forEachRootCallArg')(root, function (_ref) {
      var identifyingArgKey = _ref.identifyingArgKey;

      var dataID = _this2._store.getDataID(storageKey, identifyingArgKey);
      if (dataID == null) {
        state.result = false;
      } else {
        var nextState = {
          dataID: dataID,
          rangeInfo: undefined,
          result: true
        };
        _this2.traverse(root, nextState);
        state.result = state.result && nextState.result;
      }
    });
  };

  RelayQueryChecker.prototype.visitFragment = function visitFragment(fragment, state) {
    var dataID = state.dataID;
    // The dataID check is for Flow; it must be non-null to have gotten here.
    if (dataID && require('./isCompatibleRelayFragmentType')(fragment, this._store.getType(dataID))) {
      this.traverse(fragment, state);
    }
  };

  RelayQueryChecker.prototype.visitField = function visitField(field, state) {
    var dataID = state.dataID;
    var recordState = dataID && this._store.getRecordState(dataID);
    if (recordState === require('./RelayRecordState').UNKNOWN) {
      state.result = false;
      return;
    } else if (recordState === require('./RelayRecordState').NONEXISTENT) {
      return;
    }
    var rangeInfo = state.rangeInfo;
    if (rangeInfo && field.getSchemaName() === EDGES) {
      this._checkEdges(field, state);
    } else if (rangeInfo && field.getSchemaName() === PAGE_INFO) {
      this._checkPageInfo(field, state);
    } else if (!field.canHaveSubselections()) {
      this._checkScalar(field, state);
    } else if (field.isPlural()) {
      this._checkPlural(field, state);
    } else if (field.isConnection()) {
      this._checkConnection(field, state);
    } else {
      this._checkLinkedField(field, state);
    }
  };

  RelayQueryChecker.prototype._checkScalar = function _checkScalar(field, state) {
    var fieldData = state.dataID && this._store.getField(state.dataID, field.getStorageKey());
    if (fieldData === undefined) {
      state.result = false;
    }
  };

  RelayQueryChecker.prototype._checkPlural = function _checkPlural(field, state) {
    var dataIDs = state.dataID && this._store.getLinkedRecordIDs(state.dataID, field.getStorageKey());
    if (dataIDs === undefined) {
      state.result = false;
      return;
    }
    if (dataIDs) {
      for (var ii = 0; ii < dataIDs.length; ii++) {
        if (!state.result) {
          break;
        }
        var nextState = {
          dataID: dataIDs[ii],
          rangeInfo: undefined,
          result: true
        };
        this.traverse(field, nextState);
        state.result = nextState.result;
      }
    }
  };

  RelayQueryChecker.prototype._checkConnection = function _checkConnection(field, state) {
    var calls = field.getCallsWithValues();
    var dataID = state.dataID && this._store.getLinkedRecordID(state.dataID, field.getStorageKey());
    if (dataID === undefined) {
      state.result = false;
      return;
    }
    var nextState = {
      dataID: dataID,
      rangeInfo: null, // Flow rejects `undefined` here
      result: true
    };
    var metadata = this._store.getRangeMetadata(dataID, calls);
    if (metadata) {
      nextState.rangeInfo = metadata;
    }
    this.traverse(field, nextState);
    state.result = state.result && nextState.result;
  };

  RelayQueryChecker.prototype._checkEdges = function _checkEdges(field, state) {
    var rangeInfo = state.rangeInfo;
    if (!rangeInfo) {
      state.result = false;
      return;
    }
    if (rangeInfo.diffCalls.length) {
      state.result = false;
      return;
    }
    var edges = rangeInfo.filteredEdges;
    for (var ii = 0; ii < edges.length; ii++) {
      if (!state.result) {
        break;
      }
      var nextState = {
        dataID: edges[ii].edgeID,
        rangeInfo: undefined,
        result: true
      };
      this.traverse(field, nextState);
      state.result = nextState.result;
    }
  };

  RelayQueryChecker.prototype._checkPageInfo = function _checkPageInfo(field, state) {
    var rangeInfo = state.rangeInfo;
    if (!rangeInfo || !rangeInfo.pageInfo) {
      state.result = false;
      return;
    }
  };

  RelayQueryChecker.prototype._checkLinkedField = function _checkLinkedField(field, state) {
    var dataID = state.dataID && this._store.getLinkedRecordID(state.dataID, field.getStorageKey());
    if (dataID === undefined) {
      state.result = false;
      return;
    }
    if (dataID) {
      var nextState = {
        dataID: dataID,
        rangeInfo: undefined,
        result: true
      };
      this.traverse(field, nextState);
      state.result = state.result && nextState.result;
    }
  };

  return RelayQueryChecker;
}(require('./RelayQueryVisitor'));

module.exports = require('./RelayProfiler').instrument('checkRelayQueryData', checkRelayQueryData);