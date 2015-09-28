'use strict';

var removeSelf = require('./util/removeSelf');
var identical = require('./util/identicalValues');
var canMergeProperties = require('./util/canMergeProperties');
var postcss = require('postcss');

var trbl = ['top', 'right', 'bottom', 'left'];

function getLastNode (rule, prop) {
    return rule.nodes.filter(function (node) {
        return ~node.prop.indexOf(prop);
    }).pop();
}

function hasAllProps (rule, props) {
    return props.every(function (prop) {
        return rule.nodes.some(function (node) {
            return node.prop && ~node.prop.indexOf(prop);
        });
    });
}

function mergeLonghand (type) {
    function dashed (direction) {
        return type + '-' + direction;
    }

    return function (rule) {
        if (hasAllProps(rule, trbl.map(dashed))) {
            var rules = [
                getLastNode(rule, dashed('top')),
                getLastNode(rule, dashed('right')),
                getLastNode(rule, dashed('bottom')),
                getLastNode(rule, dashed('left'))
            ];

            if (canMergeProperties.apply(this, rules)) {
                var value = rules.map(function (rule) {
                    return rule.value;
                }).join(' ');

                rules.slice(0, 3).forEach(removeSelf);

                var left = rules.slice(3)[0];
                var optimised = left.clone({
                    prop: type,
                    value: value
                });

                left.replaceWith(optimised);
            }
        }
    };
}

function mergeBorders (rule) {
    function border (str) {
        return 'border-' + str;
    }

    if (hasAllProps(rule, trbl.map(border))) {
        var rules = [
            getLastNode(rule, border('top')),
            getLastNode(rule, border('right')),
            getLastNode(rule, border('bottom')),
            getLastNode(rule, border('left'))
        ];
        if (canMergeProperties(rules) && identical(rules)) {
            rules.slice(0, 3).forEach(removeSelf);

            var left = rules.slice(3)[0];
            var optimised = left.clone({
                prop: 'border',
            });

            left.replaceWith(optimised);
        }
    }
}

module.exports = postcss.plugin('cssnano-longhand-optimiser', function () {
    return function (css) {
        css.eachRule(mergeLonghand('margin'));
        css.eachRule(mergeLonghand('padding'));
        css.eachRule(mergeBorders);
    };
});
