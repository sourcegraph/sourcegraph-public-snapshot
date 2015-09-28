'use strict';

var colourNames = require('./lib/colourNames');
var toShorthand = require('./lib/toShorthand');
var trim = require('./lib/stripWhitespace');
var zero = require('./lib/trimLeadingZero');
var ctype = require('./lib/colourType');
var color = require('color');

function filterColours (callback) {
    return Object.keys(colourNames).filter(callback);
}

function shorter (a, b) {
    return (a && a.length < b.length ? a : b).toLowerCase();
}

function colormin (colour) {
    if (ctype.isRGBorHSL(colour)) {
        var c = color(colour);
        if (c.alpha() === 1) {
            // At full alpha, just use hex
            colour = c.hexString();
        } else {
            var rgb = c.rgb();
            if (!rgb.r && !rgb.g && !rgb.b && !rgb.a) {
                return 'transparent';
            }
            var hsla = c.hslaString();
            var rgba = c.rgbString();
            return zero(trim(hsla.length < rgba.length ? hsla : rgba));
        }
    }
    if (ctype.isHex(colour)) {
        colour = toShorthand(colour.toLowerCase());
        var keyword = filterColours(function (key) {
            return colourNames[key] === colour;
        })[0];
        return shorter(keyword, colour);
    } else if (ctype.isKeyword(colour)) {
        var hex = colourNames[filterColours(function (key) {
            return key === colour.toLowerCase();
        })[0]];
        return shorter(hex, colour);
    }
    // Possibly malformed, just pass through
    return colour;
}

module.exports = colormin;
