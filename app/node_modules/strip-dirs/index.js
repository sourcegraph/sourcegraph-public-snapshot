'use strict';

var path = require('path');

var isAbsolutePath = require('is-absolute');
var isInt = require('is-integer');

module.exports = function stripDirs(pathStr, count, option) {
  option = option || {narrow: false};

  if (arguments.length < 2) {
    throw new Error('Expecting two arguments and more. (path, count[, option])');
  }

  if (typeof pathStr !== 'string') {
    throw new TypeError(pathStr + ' is not a string.');
  }
  if (isAbsolutePath(pathStr)) {
    throw new TypeError(pathStr + ' is an absolute path. A relative path required.');
  }

  if (!isInt(count)) {
    throw new TypeError(count + ' is not an integer.');
  }
  if (count < 0) {
    throw new RangeError('Expecting a natural number or 0.');
  }

  var pathComponents = path.normalize(pathStr).split(path.sep);
  if (pathComponents.length > 1 && pathComponents[0] === '.') {
    pathComponents.shift();
  }

  if (count > pathComponents.length - 1) {
    if (option.narrow) {
      throw new RangeError('Cannot strip more directories than there are.');
    }
    count = pathComponents.length - 1;
  }

  return path.join.apply(null, pathComponents.slice(count));
};
