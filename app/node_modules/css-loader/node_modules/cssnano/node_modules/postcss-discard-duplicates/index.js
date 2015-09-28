'use strict';

var postcss = require('postcss');

function filterIdent (cache) {
    return function (node) {
        // Ensure we don't dedupe in different contexts
        var sameContext = node.parent.type === cache.parent.type;
        // Ensure that at rules have exactly the same name; this accounts for
        // vendor prefixes
        if (node.parent.type !== 'root') {
            sameContext = sameContext && node.parent.name === cache.parent.name;
        }
        return sameContext && String(node) === String(cache);
    };
}

function dedupe (callback) {
    var cache = [];
    return function (rule) {
        if (!callback || callback.call(this, rule)) {
            var cached = cache.filter(filterIdent(rule));
            if (cached.length) {
                cached[0].removeSelf();
                cache.splice(cache.indexOf(cached[0]), 1);
            }
            cache.push(rule);
        }
    };
}

module.exports = postcss.plugin('postcss-discard-duplicates', function () {
    return function (css) {
        css.eachAtRule(dedupe());
        css.eachAtRule(function (rule) {
            rule.eachRule(dedupe());
        });
        css.eachRule(function (rule) {
            rule.eachDecl(dedupe());
        });
        css.eachRule(dedupe(function (rule) {
            return rule.parent.type === 'root';
        }));
    };
});
