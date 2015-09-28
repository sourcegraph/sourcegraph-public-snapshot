'use strict';

module.exports = function toShorthand (hex) {
    if (hex[1] === hex[2] && hex[3] === hex[4] && hex[5] === hex[6]) {
        return '#' + hex[2] + hex[4] + hex[6];
    }
    return hex;
};
