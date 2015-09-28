'use strict';

var postcss = require('postcss');
var shorter = require('./lib/shorter');
var normalize = require('normalize-url');
var isAbsolute = require('is-absolute-url');
var path = require('path');
var assign = require('object-assign');

var cssList = require('css-list');

var multiline = /\\[\r\n]/;
var unquote = /^("|')(.*)\1$/;
var escapeChars = /([\s\(\)"'])/g;

function convert (url, options) {
    if (isAbsolute(url) || !url.indexOf('//')) {
        return normalize(url, options);
    }
    return path.normalize(url).replace(new RegExp('\\' + path.sep, 'g'), '/');
}

function namespaceOptimiser (options) {
    return function (rule) {
        rule.params = cssList.map(rule.params, function (param) {
            if (/^url/.test(param)) {
                param = param.replace(/^url\((.*)\)$/, '$1');
            }
            return param.replace(/^("|')(.*)\1$/, function (_, quo, body) {
                return quo + convert(body.trim(), options) + quo;
            });
        });
    }
}

function eachValue (val, options) {
    return cssList.map(val, function (value, type) {
        if (
            type !== 'func' ||
            value.indexOf('url') !== 0 ||
            ~value.indexOf('data:image/') ||
            ~value.indexOf('data:application/')
        ) {
            return value;
        }
        var url = value.substring(4, value.length - 1).trim();
        url = url.replace(unquote, function (_, quote, body) {
            return quote + convert(body.trim(), options) + quote;
        });
        var trimmed = url.replace(unquote, '$2').trim();
        var optimised = convert(trimmed, options);
        if (escapeChars.test(trimmed)) {
            var isEscaped = trimmed.replace(escapeChars, '\\$1');
            optimised = shorter(isEscaped, url);
        }
        return 'url(' + optimised + ')';
    });
}

module.exports = postcss.plugin('postcss-normalize-url', function (options) {
    options = assign({
        normalizeProtocol: false,
        stripFragment: false,
        stripWWW: true
    }, options);

    return function (css) {
        css.eachDecl(function (decl) {
            decl.value = eachValue(decl.value.replace(multiline, ''), options);
        });
        css.eachAtRule('namespace', namespaceOptimiser(options));
    };
});
