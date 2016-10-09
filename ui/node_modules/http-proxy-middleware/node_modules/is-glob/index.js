/*!
 * is-glob <https://github.com/jonschlinkert/is-glob>
 *
 * Copyright (c) 2014-2016, Jon Schlinkert.
 * Licensed under the MIT License.
 */

var isExtglob = require('is-extglob');

module.exports = function isGlob(str) {
  if (!str || typeof str !== 'string') {
    return false;
  }

  if (isExtglob(str)) return true;
  var m, matches = [];

  while ((m = /(\\).|([*?]|\[.*\]|\{.*\}|\(.*\|.*\)|^!)/g.exec(str))) {
    if (m[2]) matches.push(m[2]);
    str = str.slice(m.index + m[0].length);
  }
  return matches.length > 0;
};
