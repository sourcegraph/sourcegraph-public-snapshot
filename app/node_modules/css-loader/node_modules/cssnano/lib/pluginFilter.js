'use strict';

var postcss = require('postcss');

module.exports = postcss.plugin('_pluginFilter', function () {
    return function (css, result) {
        var previousPlugins = [];
        var filter = false;
        var position = 0;

        while (position < result.processor.plugins.length) {
            var plugin = result.processor.plugins[position];
            if (plugin.postcssPlugin === '_pluginFilter') {
                position ++;
                filter = true;
                continue;
            }
            if (!filter) {
                previousPlugins.push(plugin.postcssPlugin);
                position ++;
                continue;
            }
            if (~previousPlugins.indexOf(plugin.postcssPlugin)) {
                result.processor.plugins.splice(position, 1);
                continue;
            }
            position ++;
        }
    }
});
