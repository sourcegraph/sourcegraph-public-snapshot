/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @typechecks
 */

/**
 * Unicode-enabled extra utility functions not always needed.
 */

'use strict';

var UnicodeUtils = require('./UnicodeUtils');

/**
 * @param {number} codePoint  Valid Unicode code-point
 * @return {string}           A formatted Unicode code-point string
 *                            of the format U+XXXX, U+XXXXX, or U+XXXXXX
 */
function formatCodePoint(codePoint) {
  codePoint = codePoint || 0; // NaN --> 0
  var codePointHex = codePoint.toString(16).toUpperCase();
  return 'U+' + '0'.repeat(Math.max(0, 4 - codePointHex.length)) + codePointHex;
}

/**
 * Get a list of formatted (string) Unicode code-points from a String
 *
 * @param {string} str        Valid Unicode string
 * @return {array<string>}    A list of formatted code-point strings
 */
function getCodePointsFormatted(str) {
  var codePoints = UnicodeUtils.getCodePoints(str);
  return codePoints.map(formatCodePoint);
}

var UnicodeUtilsExtra = {
  formatCodePoint: formatCodePoint,
  getCodePointsFormatted: getCodePointsFormatted
};

module.exports = UnicodeUtilsExtra;