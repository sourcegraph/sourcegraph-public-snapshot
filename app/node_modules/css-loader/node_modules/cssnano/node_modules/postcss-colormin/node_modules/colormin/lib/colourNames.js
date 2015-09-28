'use strict';

var colours = require('css-color-names');
var toShorthand = require('./toShorthand');

Object.keys(colours).forEach(function (colour) {
    colours[colour] = toShorthand(colours[colour]);
});

module.exports = colours;
