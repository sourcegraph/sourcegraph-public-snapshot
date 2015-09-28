'use strict';

var postcss = require('postcss');
var list = postcss.list;
var flatten = require('flatten');
var clone = require('./lib/clone');

var prefixes = ['-webkit-', '-moz-', '-ms-', '-o-'];

function intersect (a, b, not) {
    return a.filter(function (c) {
        var index = ~b.indexOf(c);
        return not ? !index : index;
    });
}

function different (a, b) {
    return intersect(a, b, true).concat(intersect(b, a, true));
}

function filterPrefixes (selector) {
    return intersect(prefixes, selector);
}

function sameVendor (selectorsA, selectorsB) {
    var same = function (selectors) {
        return flatten(selectors.map(filterPrefixes)).join();
    };
    return same(selectorsA) === same(selectorsB);
}

function noVendor (selector) {
    return !filterPrefixes(selector).length;
}

function sameParent (ruleA, ruleB) {
    var hasParent = ruleA.parent && ruleB.parent;
    var sameType = hasParent && ruleA.parent.type === ruleB.parent.type;
    // If an at rule, ensure that the parameters are the same
    if (hasParent && ruleA.parent.type !== 'root' && ruleB.parent.type !== 'root') {
        sameType = sameType && ruleA.parent.params === ruleB.parent.params;
    }
    return hasParent ? sameType : true;
}

function canMerge (ruleA, ruleB) {
    var a = list.comma(ruleA.selector);
    var b = list.comma(ruleB.selector);

    var parent = sameParent(ruleA, ruleB);
    return parent && (a.concat(b).every(noVendor) || sameVendor(a, b));
}

function getDeclarations (rule) {
    return rule.nodes.map(String);
}

function joinSelectors (/* rules... */) {
    var args = Array.prototype.slice.call(arguments);
    return flatten(args.map(function (s) { return s.selector; })).join(',');
}

function ruleLength (/* rules... */) {
    var args = Array.prototype.slice.call(arguments);
    return args.map(function (selector) {
        return selector.nodes.length ? String(selector) : '';
    }).join('').length;
}

function partialMerge (first, second) {
    var intersection = intersect(getDeclarations(first), getDeclarations(second));
    if (!intersection.length) {
        return second;
    }
    var nextRule = second.next();
    if (nextRule && nextRule.type !== 'comment') {
        var nextIntersection = intersect(getDeclarations(second), getDeclarations(nextRule));
        if (nextIntersection.length > intersection.length) {
            first = second; second = nextRule; intersection = nextIntersection;
        }
    }
    var recievingBlock = second.cloneBefore({
        selector: joinSelectors(first, second),
        nodes: [],
        before: ''
    });
    var difference = different(getDeclarations(first), getDeclarations(second));
    var firstClone = clone(first);
    var secondClone = clone(second);
    var moveDecl = function (callback) {
        return function (decl) {
            var intersects = ~intersection.indexOf(String(decl));
            var baseProperty = decl.prop.split('-')[0];
            var canMove = difference.every(function (d) {
                return d.split(':')[0] !== baseProperty;
            });
            if (intersects && canMove) {
                callback.call(this, decl);
            }
        };
    };
    firstClone.eachInside(moveDecl(function (decl) {
        decl.moveTo(recievingBlock);
    }));
    secondClone.eachInside(moveDecl(function (decl) {
        decl.removeSelf();
    }));
    var merged = ruleLength(firstClone, recievingBlock, secondClone);
    var original = ruleLength(first, second);
    if (merged < original) {
        first.replaceWith(firstClone);
        second.replaceWith(secondClone);
        [firstClone, recievingBlock, secondClone].forEach(function (r) {
            if (!r.nodes.length) {
                r.removeSelf();
            }
        });
        return secondClone;
    } else {
        recievingBlock.removeSelf();
        return first;
    }
}

function selectorMerger () {
    var cache = null;
    return function (rule) {
        // Prime the cache with the first rule, or alternately ensure that it is
        // safe to merge both declarations before continuing
        if (!cache || !canMerge(rule, cache)) {
            cache = rule;
            return;
        }
        var cacheDecls = getDeclarations(cache);
        var ruleDecls = getDeclarations(rule);
        // Merge when declarations are exactly equal
        // e.g. h1 { color: red } h2 { color: red }
        if (ruleDecls.join(';') === cacheDecls.join(';')) {
            rule.selector = joinSelectors(cache, rule);
            cache.removeSelf();
            cache = rule;
            return;
        }
        // Merge when both selectors are exactly equal
        // e.g. a { color: blue } a { font-weight: bold }
        if (cache.selector === rule.selector) {
            rule.eachInside(function (declaration) {
                declaration.moveTo(cache);
            });
            rule.removeSelf();
            return;
        }
        // Partial merge: check if the rule contains a subset of the last; if
        // so create a joined selector with the subset, if smaller.
        cache = partialMerge(cache, rule);
    };
}

module.exports = postcss.plugin('postcss-merge-rules', function () {
    return function (css) {
        css.eachRule(selectorMerger());
    };
});
