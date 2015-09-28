'use strict';

var postcss = require('postcss');

function remove (callback) {
    return function (node) {
        callback.call(this, node) && node.removeSelf();
    };
}

module.exports = postcss.plugin('postcss-discard-empty', function () {
    return function (css) {
        css.eachDecl(remove(function (decl) {
            return !decl.value;
        }));
        css.eachRule(remove(function (rule) {
            return !rule.selector.length || !rule.nodes.length;
        }));
        css.eachAtRule(remove(function (rule) {
            if (rule.nodes) {
                return !rule.nodes.length;
            }
            return !rule.params;
        }));
    };
});
