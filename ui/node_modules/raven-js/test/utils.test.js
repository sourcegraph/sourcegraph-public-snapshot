/*jshint mocha:true*/
/*global assert:false, console:true*/
'use strict';

var utils = require('../src/utils');
var Raven = require('../src/raven');
var RavenConfigError = require('../src/configError');

var isUndefined = utils.isUndefined;
var isFunction = utils.isFunction;
var isString = utils.isString;
var isObject = utils.isObject;
var isEmptyObject = utils.isEmptyObject;
var isError = utils.isError;
var joinRegExp = utils.joinRegExp;
var objectMerge = utils.objectMerge;
var truncate = utils.truncate;
var urlencode = utils.urlencode;
var htmlTreeAsString = utils.htmlTreeAsString;
var htmlElementAsString = utils.htmlElementAsString;
var parseUrl = utils.parseUrl;

describe('utils', function () {
    describe('isUndefined', function() {
        it('should do as advertised', function() {
            assert.isTrue(isUndefined());
            assert.isFalse(isUndefined({}));
            assert.isFalse(isUndefined(''));
            assert.isTrue(isUndefined(undefined));
        });
    });

    describe('isFunction', function() {
        it('should do as advertised', function() {
            assert.isTrue(isFunction(function(){}));
            assert.isFalse(isFunction({}));
            assert.isFalse(isFunction(''));
            assert.isFalse(isFunction(undefined));
        });
    });

    describe('isString', function() {
        it('should do as advertised', function() {
            assert.isTrue(isString(''));
            assert.isTrue(isString(String('')));
            assert.isTrue(isString(new String('')));
            assert.isFalse(isString({}));
            assert.isFalse(isString(undefined));
            assert.isFalse(isString(function(){}));
        });
    });

    describe('isObject', function() {
        it('should do as advertised', function() {
            assert.isTrue(isObject({}));
            assert.isTrue(isObject(new Error()))
            assert.isFalse(isObject(''));
        });
    });

    describe('isEmptyObject', function() {
        it('should work as advertised', function() {
            assert.isTrue(isEmptyObject({}));
            assert.isFalse(isEmptyObject({foo: 1}));
        });
    });

    describe('isError', function() {
        it('should work as advertised', function() {
            assert.isTrue(isError(new Error()));
            assert.isTrue(isError(new ReferenceError()));
            assert.isTrue(isError(new RavenConfigError()));
            assert.isFalse(isError({}));
            assert.isFalse(isError(''));
            assert.isFalse(isError(true));
        });
    });

    describe('objectMerge', function() {
        it('should work as advertised', function() {
            assert.deepEqual(objectMerge({}, {}), {});
            assert.deepEqual(objectMerge({a:1}, {b:2}), {a:1, b:2});
            assert.deepEqual(objectMerge({a:1}), {a:1});
        });
    });

    describe('truncate', function() {
        it('should work as advertised', function() {
            assert.equal(truncate('lolol', 3), 'lol\u2026');
            assert.equal(truncate('lolol', 10), 'lolol');
            assert.equal(truncate('lol', 3), 'lol');
            assert.equal(truncate(new Array(1000).join('f'), 0), new Array(1000).join('f'));
        });
    });


    describe('joinRegExp', function() {
        it('should work as advertised', function() {
            assert.equal(joinRegExp([
                'a', 'b', 'a.b', /d/, /[0-9]/
            ]).source, 'a|b|a\\.b|d|[0-9]');
        });

        it('should not process empty or undefined variables', function() {
            assert.equal(joinRegExp([
                'a', 'b', null, undefined
            ]).source, 'a|b');
        });

        it('should skip entries that are not strings or regular expressions in the passed array of patterns', function() {
            assert.equal(joinRegExp([
                'a', 'b', null, 'a.b', undefined, true, /d/, 123, {}, /[0-9]/, []
            ]).source, 'a|b|a\\.b|d|[0-9]');
        });
    });


    describe('urlencode', function() {
        it('should work', function() {
            assert.equal(urlencode({}), '');
            assert.equal(urlencode({'foo': 'bar', 'baz': '1 2'}), 'foo=bar&baz=1%202');
        });
    });

    describe('htmlTreeAsString', function () {
        it('should work', function () {
            var tree = {
                tagName: 'INPUT',
                id: 'the-username',
                className: 'form-control',
                getAttribute: function (key){
                    return {
                        name: 'username'
                    }[key];
                },
                parentNode: {
                    tagName: 'span',
                    getAttribute: function () {},
                    parentNode: {
                        tagName: 'div',
                        getAttribute: function () {},
                    }
                }
            };

            assert.equal(htmlTreeAsString(tree), 'div > span > input#the-username.form-control[name="username"]');
        });


        it('should not create strings that are too big', function () {
            var tree = {
                tagName: 'INPUT',
                id: 'the-username',
                className: 'form-control',
                getAttribute: function (key){
                    return {
                        name: 'username'
                    }[key];
                },
                parentNode: {
                    tagName: 'span',
                    getAttribute: function () {},
                    parentNode: {
                        tagName: 'div',
                        getAttribute: function (key) {
                            return {
                                name: 'super long input name that nobody would really ever have i mean come on look at this'
                            }[key];
                        }
                    }
                }
            };

            // NOTE: <div/> omitted because crazy long name
            assert.equal(htmlTreeAsString(tree), 'span > input#the-username.form-control[name="username"]');
        });
    });

    describe('htmlElementAsString', function () {
        it('should work', function () {
            assert.equal(htmlElementAsString({
                tagName: 'INPUT',
                id: 'the-username',
                className: 'form-control',
                getAttribute: function (key){
                    return {
                        name: 'username',
                        placeholder: 'Enter your username'
                    }[key];
                }
            }), 'input#the-username.form-control[name="username"]');

            assert.equal(htmlElementAsString({
                tagName: 'IMG',
                id: 'image-3',
                getAttribute: function (key){
                    return {
                        title: 'A picture of an apple',
                        'data-something': 'This should be ignored' // skipping data-* attributes in first implementation
                    }[key];
                }
            }), 'img#image-3[title="A picture of an apple"]');
        });

        it('should return an empty string if the input element is falsy', function () {
            assert.equal(htmlElementAsString(null), '');
            assert.equal(htmlElementAsString(0), '');
            assert.equal(htmlElementAsString(undefined), '');
        });

        it('should return an empty string if the input element has no tagName property', function () {
            assert.equal(htmlElementAsString({
                id: 'the-username',
                className: 'form-control'
            }), '');
        });

        it('should gracefully handle when className is not a string (e.g. SVGAnimatedString', function () {
            assert.equal(htmlElementAsString({
                tagName: 'INPUT',
                id: 'the-username',
                className: {}, // not a string
                getAttribute: function (key){
                    return {
                        name: 'username'
                    }[key];
                }
            }), 'input#the-username[name="username"]');
        });
    });

    describe('parseUrl', function () {
        it('should parse fully qualified URLs', function () {
            assert.deepEqual(parseUrl('http://example.com/foo'), {
                protocol: 'http',
                host: 'example.com',
                path: '/foo',
                relative: '/foo'
            });
            assert.deepEqual(parseUrl('//example.com/foo?query'), {
                protocol: undefined,
                host: 'example.com',
                path: '/foo',
                relative: '/foo?query'
            });
        });

        it('should parse partial URLs, e.g. path only', function () {
            assert.deepEqual(parseUrl('/foo'), {
                protocol: undefined,
                host: undefined,
                path: '/foo',
                relative: '/foo'
            });
            assert.deepEqual(parseUrl('example.com/foo#derp'), {
                protocol: undefined,
                host: undefined,
                path: 'example.com/foo',
                relative: 'example.com/foo#derp'
                // this is correct! pushState({}, '', 'example.com/foo') would take you
                // from example.com => example.com/example.com/foo (valid url).
            });
        });
    });
});
