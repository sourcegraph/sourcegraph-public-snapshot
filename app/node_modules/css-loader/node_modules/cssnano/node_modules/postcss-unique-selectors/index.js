'use strict';

var uniqs = require('uniqs');
var natural = require('javascript-natural-sort');
var postcss = require('postcss');
var split = require('css-list').split;

module.exports = postcss.plugin('postcss-unique-selectors', function () {
    return function (css) {
        css.eachRule(function (rule) {
            var unique = uniqs(split(rule.selector, [','])).sort(natural);
            rule.selector = unique.join(',');
        });
    }
});
