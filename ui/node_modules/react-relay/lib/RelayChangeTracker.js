/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayChangeTracker
 * 
 */

'use strict';

var _classCallCheck3 = _interopRequireDefault(require('babel-runtime/helpers/classCallCheck'));

var _freeze2 = _interopRequireDefault(require('babel-runtime/core-js/object/freeze'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

/**
 * @internal
 *
 * Keeps track of records that have been created or updated; used primarily to
 * record changes during the course of a `write` operation.
 */

var RelayChangeTracker = function () {
  function RelayChangeTracker() {
    (0, _classCallCheck3['default'])(this, RelayChangeTracker);

    this._created = {};
    this._updated = {};
  }

  /**
   * Record the creation of a record.
   */


  RelayChangeTracker.prototype.createID = function createID(recordID) {
    this._created[recordID] = true;
  };

  /**
   * Record an update to a record.
   */


  RelayChangeTracker.prototype.updateID = function updateID(recordID) {
    if (!this._created.hasOwnProperty(recordID)) {
      this._updated[recordID] = true;
    }
  };

  /**
   * Determine if the record has any changes (was created or updated).
   */


  RelayChangeTracker.prototype.hasChange = function hasChange(recordID) {
    return !!(this._updated[recordID] || this._created[recordID]);
  };

  /**
   * Determine if the record was created.
   */


  RelayChangeTracker.prototype.isNewRecord = function isNewRecord(recordID) {
    return !!this._created[recordID];
  };

  /**
   * Get the ids of records that were created/updated.
   */


  RelayChangeTracker.prototype.getChangeSet = function getChangeSet() {
    if (process.env.NODE_ENV !== 'production') {
      return {
        created: (0, _freeze2['default'])(this._created),
        updated: (0, _freeze2['default'])(this._updated)
      };
    }
    return {
      created: this._created,
      updated: this._updated
    };
  };

  return RelayChangeTracker;
}();

module.exports = RelayChangeTracker;