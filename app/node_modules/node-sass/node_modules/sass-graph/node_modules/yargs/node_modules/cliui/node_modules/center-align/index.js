/*!
 * center-align <https://github.com/jonschlinkert/center-align>
 *
 * Copycenter (c) 2015, Jon Schlinkert.
 * Licensed under the MIT License.
 */

'use strict';

var align = require('align-text');

module.exports = function centerAlign(val) {
  return align(val, function (len, longest) {
    return Math.floor((longest - len) / 2);
  });
};
