'use strict';

module.exports = function toLonghand (hex) {
    var h = hex.substring(1);
    return h.length === 3 && '#' + h[0] + h[0] + h[1] + h[1] + h[2] + h[2] || hex;
};
