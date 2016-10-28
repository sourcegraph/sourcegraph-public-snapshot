/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayMetricsRecorder
 * 
 */

'use strict';

var _extends3 = _interopRequireDefault(require('babel-runtime/helpers/extends'));

var _classCallCheck3 = _interopRequireDefault(require('babel-runtime/helpers/classCallCheck'));

var _keys2 = _interopRequireDefault(require('babel-runtime/core-js/object/keys'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

var measurementDefaults = {
  aggregateTime: 0,
  callCount: 0
};

/**
 * Collects timing information from key Relay subsystems.
 *
 * Example:
 *
 * ```
 * var recorder = new RelayMetricsRecorder();
 * recorder.start();
 * // ... do work ...
 * recorder.stop();
 * var metrics = recorder.getMetrics();
 * ```
 *
 * Metrics:
 * - `recordingTime`: the total time spent recording (between calls to `start()`
 *   and `stop()`).
 * - `totalTime`: the total time spent inside profiled Relay functions.
 * - `measurements`: an object mapping names of profiled functions to profiling
 *   data including:
 *   - `aggregateTime`: total time spent in the method.
 *   - `callCount`: number of times the method was called.
 */

var RelayMetricsRecorder = function () {
  function RelayMetricsRecorder() {
    (0, _classCallCheck3['default'])(this, RelayMetricsRecorder);

    this._isEnabled = false;
    this._measurements = {};
    this._profiles = [];
    this._profileStack = [];
    this._recordingStartTime = 0;
    this._recordingTotalTime = 0;
    this._startTimesStack = [];

    this._measure = this._measure.bind(this);
    this._instrumentProfile = this._instrumentProfile.bind(this);
    this._startMeasurement = this._startMeasurement.bind(this);
    this._stopMeasurement = this._stopMeasurement.bind(this);
  }

  RelayMetricsRecorder.prototype.start = function start() {
    if (this._isEnabled) {
      return;
    }
    this._recordingStartTime = require('fbjs/lib/performanceNow')();
    this._isEnabled = true;
    this._profileStack = [0];
    this._startTimesStack = [0];

    require('./RelayProfiler').attachAggregateHandler('*', this._measure);
    require('./RelayProfiler').attachProfileHandler('*', this._instrumentProfile);
  };

  RelayMetricsRecorder.prototype.stop = function stop() {
    if (!this._isEnabled) {
      return;
    }
    this._recordingTotalTime += require('fbjs/lib/performanceNow')() - this._recordingStartTime;
    this._isEnabled = false;

    require('./RelayProfiler').detachAggregateHandler('*', this._measure);
    require('./RelayProfiler').detachProfileHandler('*', this._instrumentProfile);
  };

  RelayMetricsRecorder.prototype.getMetrics = function getMetrics() {
    var _measurements = this._measurements;

    var totalTime = 0;
    var sortedMeasurements = {};
    (0, _keys2['default'])(_measurements).sort(function (a, b) {
      return _measurements[b].aggregateTime - _measurements[a].aggregateTime;
    }).forEach(function (name) {
      totalTime += _measurements[name].aggregateTime;
      sortedMeasurements[name] = _measurements[name];
    });
    var sortedProfiles = this._profiles.sort(function (a, b) {
      if (a.startTime < b.startTime) {
        return -1;
      } else if (a.startTime > b.startTime) {
        return 1;
      } else {
        // lower duration first
        return a.endTime - a.startTime - (b.endTime - b.startTime);
      }
    });

    return {
      measurements: sortedMeasurements,
      profiles: sortedProfiles,
      recordingTime: this._recordingTotalTime,
      totalTime: totalTime
    };
  };

  RelayMetricsRecorder.prototype._measure = function _measure(name, callback) {
    this._startMeasurement(name);
    callback();
    this._stopMeasurement(name);
  };

  RelayMetricsRecorder.prototype._instrumentProfile = function _instrumentProfile(name) {
    var _this = this;

    var startTime = require('fbjs/lib/performanceNow')();
    return function () {
      _this._profiles.push({
        endTime: require('fbjs/lib/performanceNow')(),
        name: name,
        startTime: startTime
      });
    };
  };

  RelayMetricsRecorder.prototype._startMeasurement = function _startMeasurement(name) {
    this._measurements[name] = this._measurements[name] || (0, _extends3['default'])({}, measurementDefaults);
    this._profileStack.unshift(0);
    this._startTimesStack.unshift(require('fbjs/lib/performanceNow')());
  };

  RelayMetricsRecorder.prototype._stopMeasurement = function _stopMeasurement(name) {
    var innerTime = this._profileStack.shift();
    var start = this._startTimesStack.shift();
    var totalTime = require('fbjs/lib/performanceNow')() - start;

    this._measurements[name].aggregateTime += totalTime - innerTime;
    this._measurements[name].callCount++;

    this._profileStack[0] += totalTime;
  };

  return RelayMetricsRecorder;
}();

module.exports = RelayMetricsRecorder;