'use strict';

var balanced = require('node-balanced');
var CommentRemover = require('./lib/commentRemover');
var postcss = require('postcss');
var space = postcss.list.space;

module.exports = postcss.plugin('postcss-discard-comments', function (options) {
    return function (css) {
        var remover = new CommentRemover(options || {});

        function replaceComments (source) {
            if (!source) {
                return;
            }
            var b = balanced.replacements({
                source: source,
                open: '/*',
                close: '*/',
                replace: function (comment, head, tail) {
                    if (remover.canRemove(comment)) {
                        return ' ';
                    }
                    return head + comment + tail;
                }
            });
            return space(b).join(' ');
        }

        css.eachComment(function (comment) {
            if (remover.canRemove(comment.text)) {
                comment.removeSelf();
            }
        });

        css.eachDecl(function (decl) {
            decl.between = replaceComments(decl.between);
            if (decl._value && decl._value.raw) {
                var replaced = replaceComments(decl._value.raw);
                decl._value.raw = decl._value.value = decl.value = replaced;
            }
            if (decl._important) {
                decl._important = replaceComments(decl._important);
                var b = balanced.matches({
                    source: decl._important,
                    open: '/*',
                    close: '*/'
                });
                decl._important = b.length ? decl._important : '!important';
            }
        });

        css.eachRule(function (rule) {
            if (rule.between) {
                rule.between = replaceComments(rule.between);
            }
            if (rule._selector && rule._selector.raw) {
                rule._selector.raw = replaceComments(rule._selector.raw);
            }
        });

        css.eachAtRule(function (rule) {
            if (rule.afterName) {
                var commentsReplaced = replaceComments(rule.afterName);
                if (!commentsReplaced.length) {
                    rule.afterName = commentsReplaced + ' ';
                } else {
                    rule.afterName = ' ' + commentsReplaced + ' ';
                }
            }
            if (rule._params && rule._params.raw) {
                rule._params.raw = replaceComments(rule._params.raw);
            }
            if (rule.between) {
                rule.between = replaceComments(rule.between);
            }
        });
    };
});
