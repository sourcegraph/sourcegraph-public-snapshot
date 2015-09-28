'use strict';

var colourNames = require('./colourNames');
var toLonghand = require('./toLonghand');

module.exports.isHex = function (colour) {
    if (colour[0] === '#') {
        var c = toLonghand(colour).substring(1);
        return c.length === 6 && ! isNaN(parseInt(c, 16));
    }
    return false;
};

module.exports.isRGBorHSL = function (colour) {
    return /^(rgb|hsl)a?\(.*?\)/.test(colour);
};

module.exports.isKeyword = function (colour) {
    return ~Object.keys(colourNames).indexOf(colour.toLowerCase());
};
