/*!
 * is-extglob <https://github.com/jonschlinkert/is-extglob>
 *
 * Copyright (c) 2014-2016, Jon Schlinkert.
 * Licensed under the MIT License.
 */

module.exports = function isExtglob(str) {
  if (!str || typeof str !== 'string') {
    return false;
  }

  var m, matches = [];
  while ((m = /(\\).|([@?!+*]\(.*\))/g.exec(str))) {
    if (m[2]) matches.push(m[2]);
    str = str.slice(m.index + m[0].length);
  }
  return matches.length;
};
