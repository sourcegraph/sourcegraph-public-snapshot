/*!
 * object.omit <https://github.com/jonschlinkert/object.omit>
 *
 * Copyright (c) 2014-2015, Jon Schlinkert.
 * Licensed under the MIT License.
 */

'use strict';

var isObject = require('isobject');
var forOwn = require('for-own');

module.exports = function omit(obj, keys) {
  if (!isObject(obj)) return {};
  if (!keys) return obj;

  keys = Array.isArray(keys) ? keys : [keys];
  var res = {};

  forOwn(obj, function (value, key) {
    if (keys.indexOf(key) === -1) {
      res[key] = value;
    }
  });
  return res;
};
