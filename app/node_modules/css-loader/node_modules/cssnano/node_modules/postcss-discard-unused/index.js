'use strict';

var uniqs = require('uniqs');
var postcss = require('postcss');
var flatten = require('flatten');
var comma = postcss.list.comma;
var space = postcss.list.space;

function filterAtRule (css, properties, atrule) {
    var cache = [];
    css.eachDecl(properties, function (decl) {
        cache.push(space(decl.value));
    });
    cache = uniqs(flatten(cache));
    css.eachAtRule(atrule, function (rule) {
        var hasAtRule = cache.some(function (c) {
            return c === rule.params;
        });
        if (!hasAtRule) {
            rule.removeSelf();
        }
    });
}

module.exports = postcss.plugin('postcss-discard-unused', function () {
    return function (css) {
        // fonts have slightly different logic
        var cache = [];
        css.eachRule(function (rule) {
            rule.eachDecl(/font(|-family)/, function (decl) {
                cache.push(comma(decl.value));
            });
        });
        cache = uniqs(flatten(cache));
        css.eachAtRule('font-face', function (rule) {
            var fontFamilies = rule.nodes.filter(function (node) {
                return node.prop === 'font-family';
            });
            // Discard the @font-face if it has no font-family
            if (!fontFamilies.length) {
                return rule.removeSelf();
            }
            fontFamilies.forEach(function (family) {
                var hasFont = comma(family.value).some(function (font) {
                    return cache.some(function (c) {
                        return ~c.indexOf(font);
                    });
                });
                if (!hasFont) {
                    rule.removeSelf();
                }
            });
        });

        // keyframes & counter styles
        filterAtRule(css, /list-style|system/, 'counter-style');
        filterAtRule(css, /animation/, /keyframes/);
    };
});
