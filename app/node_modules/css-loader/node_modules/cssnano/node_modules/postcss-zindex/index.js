'use strict';

var postcss = require('postcss');

module.exports = postcss.plugin('postcss-zindex', function () {
    return function (css) {
        var cache = require('./lib/layerCache')();
        // First pass; cache all z indexes
        css.eachDecl('z-index', function (declaration) {
            cache.addValue(declaration.value);
        });
        // Second pass; optimise
        css.eachDecl('z-index', function (declaration) {
            // Need to coerce to string so that the
            // AST is updated correctly
            declaration.value = '' + cache.convert(declaration.value);
        });
    };
});
