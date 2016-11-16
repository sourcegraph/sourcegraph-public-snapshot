/*global Mocha, assert*/
'use strict';

var TraceKit = require('../../vendor/TraceKit/tracekit');
var CapturedExceptions = require('./fixtures/captured-errors');

describe('TraceKit', function () {
    describe('Parser', function() {
        it('should parse Safari 6 error', function () {
            var stackFrames = TraceKit.computeStackTrace(CapturedExceptions.SAFARI_6);
            assert.ok(stackFrames);
            assert.deepEqual(stackFrames.stack.length, 4);
            assert.deepEqual(stackFrames.stack[0], { url: 'http://path/to/file.js', func: '?', args: [], line: 48, column: null });
            assert.deepEqual(stackFrames.stack[1], { url: 'http://path/to/file.js', func: 'dumpException3', args: [], line: 52, column: null });
            assert.deepEqual(stackFrames.stack[2], { url: 'http://path/to/file.js', func: 'onclick', args: [], line: 82, column: null });
            assert.deepEqual(stackFrames.stack[3], { url: '[native code]', func: '?', args: [], line: null, column: null });
        });

        it('should parse Safari 7 error', function () {
            var stackFrames = TraceKit.computeStackTrace(CapturedExceptions.SAFARI_7);
            assert.ok(stackFrames);
            assert.deepEqual(stackFrames.stack.length, 3);
            assert.deepEqual(stackFrames.stack[0], { url: 'http://path/to/file.js', func: '?', args: [], line: 48, column: 22 });
            assert.deepEqual(stackFrames.stack[1], { url: 'http://path/to/file.js', func: 'foo', args: [], line: 52, column: 15 });
            assert.deepEqual(stackFrames.stack[2], { url: 'http://path/to/file.js', func: 'bar', args: [], line: 108, column: 107 });
        });

        it('should parse Safari 8 error', function () {
            var stackFrames = TraceKit.computeStackTrace(CapturedExceptions.SAFARI_8);
            assert.ok(stackFrames);
            assert.deepEqual(stackFrames.stack.length, 3);
            assert.deepEqual(stackFrames.stack[0], { url: 'http://path/to/file.js', func: '?', args: [], line: 47, column: 22 });
            assert.deepEqual(stackFrames.stack[1], { url: 'http://path/to/file.js', func: 'foo', args: [], line: 52, column: 15 });
            assert.deepEqual(stackFrames.stack[2], { url: 'http://path/to/file.js', func: 'bar', args: [], line: 108, column: 23 });
        });

        it('should parse Safari 8 eval error', function () {
            // TODO: Take into account the line and column properties on the error object and use them for the first stack trace.
            var stackFrames = TraceKit.computeStackTrace(CapturedExceptions.SAFARI_8_EVAL);
            assert.ok(stackFrames);
            assert.deepEqual(stackFrames.stack.length, 3);
            assert.deepEqual(stackFrames.stack[0], { url: '[native code]', func: 'eval', args: [], line: null, column: null });
            assert.deepEqual(stackFrames.stack[1], { url: 'http://path/to/file.js', func: 'foo', args: [], line: 58, column: 21 });
            assert.deepEqual(stackFrames.stack[2], { url: 'http://path/to/file.js', func: 'bar', args: [], line: 109, column: 91 });
        });

        it('should parse Firefox 3 error', function () {
            var stackFrames = TraceKit.computeStackTrace(CapturedExceptions.FIREFOX_3);
            assert.ok(stackFrames);
            assert.deepEqual(stackFrames.stack.length, 7);
            assert.deepEqual(stackFrames.stack[0], { url: 'http://127.0.0.1:8000/js/stacktrace.js', func: '?', args: [], line: 44, column: null });
            assert.deepEqual(stackFrames.stack[1], { url: 'http://127.0.0.1:8000/js/stacktrace.js', func: '?', args: ['null'], line: 31, column: null });
            assert.deepEqual(stackFrames.stack[2], { url: 'http://127.0.0.1:8000/js/stacktrace.js', func: 'printStackTrace', args: [], line: 18, column: null });
            assert.deepEqual(stackFrames.stack[3], { url: 'http://127.0.0.1:8000/js/file.js', func: 'bar', args: ['1'], line: 13, column: null });
            assert.deepEqual(stackFrames.stack[4], { url: 'http://127.0.0.1:8000/js/file.js', func: 'bar', args: ['2'], line: 16, column: null });
            assert.deepEqual(stackFrames.stack[5], { url: 'http://127.0.0.1:8000/js/file.js', func: 'foo', args: [], line: 20, column: null });
            assert.deepEqual(stackFrames.stack[6], { url: 'http://127.0.0.1:8000/js/file.js', func: '?', args: [], line: 24, column: null });
        });

        it('should parse Firefox 7 error', function () {
            var stackFrames = TraceKit.computeStackTrace(CapturedExceptions.FIREFOX_7);
            assert.ok(stackFrames);
            assert.deepEqual(stackFrames.stack.length, 7);
            assert.deepEqual(stackFrames.stack[0], { url: 'file:///G:/js/stacktrace.js', func: '?', args: [], line: 44, column: null });
            assert.deepEqual(stackFrames.stack[1], { url: 'file:///G:/js/stacktrace.js', func: '?', args: ['null'], line: 31, column: null });
            assert.deepEqual(stackFrames.stack[2], { url: 'file:///G:/js/stacktrace.js', func: 'printStackTrace', args: [], line: 18, column: null });
            assert.deepEqual(stackFrames.stack[3], { url: 'file:///G:/js/file.js', func: 'bar', args: ['1'], line: 13, column: null });
            assert.deepEqual(stackFrames.stack[4], { url: 'file:///G:/js/file.js', func: 'bar', args: ['2'], line: 16, column: null });
            assert.deepEqual(stackFrames.stack[5], { url: 'file:///G:/js/file.js', func: 'foo', args: [], line: 20, column: null });
            assert.deepEqual(stackFrames.stack[6], { url: 'file:///G:/js/file.js', func: '?', args: [], line: 24, column: null });
        });

        it('should parse Firefox 14 error', function () {
            var stackFrames = TraceKit.computeStackTrace(CapturedExceptions.FIREFOX_14);
            assert.ok(stackFrames);
            assert.deepEqual(stackFrames.stack.length, 3);
            assert.deepEqual(stackFrames.stack[0], { url: 'http://path/to/file.js', func: '?', args: [], line: 48, column: null });
            assert.deepEqual(stackFrames.stack[1], { url: 'http://path/to/file.js', func: 'dumpException3', args: [], line: 52, column: null });
            assert.deepEqual(stackFrames.stack[2], { url: 'http://path/to/file.js', func: 'onclick', args: [], line: 1, column: null });
        });

        it('should parse Firefox 31 error', function () {
            var stackFrames = TraceKit.computeStackTrace(CapturedExceptions.FIREFOX_31);
            assert.ok(stackFrames);
            assert.deepEqual(stackFrames.stack.length, 3);
            assert.deepEqual(stackFrames.stack[0], { url: 'http://path/to/file.js', func: 'foo', args: [], line: 41, column: 13 });
            assert.deepEqual(stackFrames.stack[1], { url: 'http://path/to/file.js', func: 'bar', args: [], line: 1, column: 1 });
            assert.deepEqual(stackFrames.stack[2], { url: 'http://path/to/file.js', func: '.plugin/e.fn[c]/<', args: [], line: 1, column: 1 });
        });

        it('should parse Firefox 44 ns exceptions', function () {
            var stackFrames = TraceKit.computeStackTrace(CapturedExceptions.FIREFOX_44_NS_EXCEPTION);
            assert.ok(stackFrames);
            assert.deepEqual(stackFrames.stack.length, 4);
            assert.deepEqual(stackFrames.stack[0], { url: 'http://path/to/file.js', func: '[2]</Bar.prototype._baz/</<', args: [], line: 703, column: 28 });
            assert.deepEqual(stackFrames.stack[1], { url: 'file:///path/to/file.js', func: 'App.prototype.foo', args: [], line: 15, column: 2 });
            assert.deepEqual(stackFrames.stack[2], { url: 'file:///path/to/file.js', func: 'bar', args: [], line: 20, column: 3 });
            assert.deepEqual(stackFrames.stack[3], { url: 'file:///path/to/index.html', func: '?', args: [], line: 23, column: 1 });
        });

        it('should parse Chrome error with no location', function () {
            var stackFrames = TraceKit.computeStackTrace({stack: "error\n at Array.forEach (native)"});
            assert.deepEqual(stackFrames.stack.length, 1);
            assert.deepEqual(stackFrames.stack[0], { url: null, func: 'Array.forEach', args: ['native'], line: null, column: null });
        });

        it('should parse Chrome 15 error', function () {
            var stackFrames = TraceKit.computeStackTrace(CapturedExceptions.CHROME_15);
            assert.ok(stackFrames);
            assert.deepEqual(stackFrames.stack.length, 4);
            assert.deepEqual(stackFrames.stack[0], { url: 'http://path/to/file.js', func: 'bar', args: [], line: 13, column: 17 });
            assert.deepEqual(stackFrames.stack[1], { url: 'http://path/to/file.js', func: 'bar', args: [], line: 16, column: 5 });
            assert.deepEqual(stackFrames.stack[2], { url: 'http://path/to/file.js', func: 'foo', args: [], line: 20, column: 5 });
            assert.deepEqual(stackFrames.stack[3], { url: 'http://path/to/file.js', func: '?', args: [], line: 24, column: 4 });
        });

        it('should parse Chrome 36 error with port numbers', function () {
            var stackFrames = TraceKit.computeStackTrace(CapturedExceptions.CHROME_36);
            assert.ok(stackFrames);
            assert.deepEqual(stackFrames.stack.length, 3);
            assert.deepEqual(stackFrames.stack[0], { url: 'http://localhost:8080/file.js', func: 'dumpExceptionError', args: [], line: 41, column: 27 });
            assert.deepEqual(stackFrames.stack[1], { url: 'http://localhost:8080/file.js', func: 'HTMLButtonElement.onclick', args: [], line: 107, column: 146 });
            assert.deepEqual(stackFrames.stack[2], { url: 'http://localhost:8080/file.js', func: 'I.e.fn.(anonymous function) [as index]', args: [], line: 10, column: 3651 });
        });

        it('should parse Chrome error with blob URLs', function () {
            var stackFrames = TraceKit.computeStackTrace(CapturedExceptions.CHROME_48_BLOB);
            assert.ok(stackFrames);
            assert.deepEqual(stackFrames.stack.length, 7);
            assert.deepEqual(stackFrames.stack[1], { url: 'blob:http%3A//localhost%3A8080/abfc40e9-4742-44ed-9dcd-af8f99a29379', func: 's', args: [], line: 31, column: 29146 });
            assert.deepEqual(stackFrames.stack[2], { url: 'blob:http%3A//localhost%3A8080/abfc40e9-4742-44ed-9dcd-af8f99a29379', func: 'Object.d [as add]', args: [  ], line: 31, column: 30039 });
            assert.deepEqual(stackFrames.stack[3], { url: 'blob:http%3A//localhost%3A8080/d4eefe0f-361a-4682-b217-76587d9f712a', func: '?', args: [], line: 15, column: 10978 });
            assert.deepEqual(stackFrames.stack[4], { url: 'blob:http%3A//localhost%3A8080/abfc40e9-4742-44ed-9dcd-af8f99a29379', func: '?', args: [], line: 1, column: 6911 });
            assert.deepEqual(stackFrames.stack[5], { url: 'blob:http%3A//localhost%3A8080/abfc40e9-4742-44ed-9dcd-af8f99a29379', func: 'n.fire', args: [], line: 7, column: 3019 });
            assert.deepEqual(stackFrames.stack[6], { url: 'blob:http%3A//localhost%3A8080/abfc40e9-4742-44ed-9dcd-af8f99a29379', func: 'n.handle', args: [], line: 7, column: 2863 });
        });

        it('should parse empty IE 9 error', function() {
            var stackFrames = TraceKit.computeStackTrace(CapturedExceptions.IE_9);
            assert.ok(stackFrames);
            stackFrames.stack && assert.deepEqual(stackFrames.stack.length, 0);
        });

        it('should parse IE 10 error', function () {
            var stackFrames = TraceKit.computeStackTrace(CapturedExceptions.IE_10);
            assert.ok(stackFrames);
            assert.deepEqual(stackFrames.stack.length, 3);
            // TODO: func should be normalized
            assert.deepEqual(stackFrames.stack[0], { url: 'http://path/to/file.js', func: 'Anonymous function', args: [], line: 48, column: 13 });
            assert.deepEqual(stackFrames.stack[1], { url: 'http://path/to/file.js', func: 'foo', args: [], line: 46, column: 9 });
            assert.deepEqual(stackFrames.stack[2], { url: 'http://path/to/file.js', func: 'bar', args: [], line: 82, column: 1 });
        });

        it('should parse IE 11 error', function () {
            var stackFrames = TraceKit.computeStackTrace(CapturedExceptions.IE_11);
            assert.ok(stackFrames);
            assert.deepEqual(stackFrames.stack.length, 3);
            // TODO: func should be normalized
            assert.deepEqual(stackFrames.stack[0], { url: 'http://path/to/file.js', func: 'Anonymous function', args: [], line: 47, column: 21 });
            assert.deepEqual(stackFrames.stack[1], { url: 'http://path/to/file.js', func: 'foo', args: [], line: 45, column: 13 });
            assert.deepEqual(stackFrames.stack[2], { url: 'http://path/to/file.js', func: 'bar', args: [], line: 108, column: 1 });
        });


        it('should parse IE 11 eval error', function () {
            var stackFrames = TraceKit.computeStackTrace(CapturedExceptions.IE_11_EVAL);
            assert.ok(stackFrames);
            assert.deepEqual(stackFrames.stack.length, 3);
            assert.deepEqual(stackFrames.stack[0], { url: 'eval code', func: 'eval code', args: [], line: 1, column: 1 });
            assert.deepEqual(stackFrames.stack[1], { url: 'http://path/to/file.js', func: 'foo', args: [], line: 58, column: 17 });
            assert.deepEqual(stackFrames.stack[2], { url: 'http://path/to/file.js', func: 'bar', args: [], line: 109, column: 1 });
        });

        it('should parse Opera 11 error', function () {
            var stackFrames = TraceKit.computeStackTrace(CapturedExceptions.OPERA_11);
            assert.ok(stackFrames);
            assert.deepEqual(stackFrames.stack.length, 5);
            assert.deepEqual(stackFrames.stack[0], { url: 'http://path/to/file.js', func: '<anonymous function: run>', args: ['[arguments not available]'], line: 27, column: null });
            assert.deepEqual(stackFrames.stack[1], { url: 'http://domain.com:1234/path/to/file.js', func: 'bar', args: ['[arguments not available]'], line: 18, column: null });
            assert.deepEqual(stackFrames.stack[2], { url: 'http://domain.com:1234/path/to/file.js', func: 'foo', args: ['[arguments not available]'], line: 11, column: null });
            assert.deepEqual(stackFrames.stack[3], { url: 'http://path/to/file.js', func: '<anonymous function>', args: [], line: 15, column: null });
            assert.deepEqual(stackFrames.stack[4], { url: 'http://path/to/file.js', func: 'Error created at <anonymous function>', args: [], line: 15, column: null });
        });

        it('should parse Opera 12 error', function () {
            // TODO: Improve anonymous function name.
            var stackFrames = TraceKit.computeStackTrace(CapturedExceptions.OPERA_12);
            assert.ok(stackFrames);
            assert.deepEqual(stackFrames.stack.length, 3);
            assert.deepEqual(stackFrames.stack[0], { url: 'http://localhost:8000/ExceptionLab.html', func: '<anonymous function>', args: ['[arguments not available]'], line: 48, column: null });
            assert.deepEqual(stackFrames.stack[1], { url: 'http://localhost:8000/ExceptionLab.html', func: 'dumpException3', args: ['[arguments not available]'], line: 46, column: null });
            assert.deepEqual(stackFrames.stack[2], { url: 'http://localhost:8000/ExceptionLab.html', func: '<anonymous function>', args: ['[arguments not available]'], line: 1, column: null });
        });

        it('should parse Opera 25 error', function () {
            var stackFrames = TraceKit.computeStackTrace(CapturedExceptions.OPERA_25);
            assert.ok(stackFrames);
            assert.deepEqual(stackFrames.stack.length, 3);
            assert.deepEqual(stackFrames.stack[0], { url: 'http://path/to/file.js', func: '?', args: [], line: 47, column: 22 });
            assert.deepEqual(stackFrames.stack[1], { url: 'http://path/to/file.js', func: 'foo', args: [], line: 52, column: 15 });
            assert.deepEqual(stackFrames.stack[2], { url: 'http://path/to/file.js', func: 'bar', args: [], line: 108, column: 168 });
        });

        it('should parse PhantomJS 1.19 error', function () {
            var stackFrames = TraceKit.computeStackTrace(CapturedExceptions.PHANTOMJS_1_19);
            assert.ok(stackFrames);
            assert.deepEqual(stackFrames.stack.length, 3);
            assert.deepEqual(stackFrames.stack[0], { url: 'file:///path/to/file.js', func: '?', args: [], line: 878, column: null });
            assert.deepEqual(stackFrames.stack[1], { url: 'http://path/to/file.js', func: 'foo', args: [], line: 4283, column: null });
            assert.deepEqual(stackFrames.stack[2], { url: 'http://path/to/file.js', func: '?', args: [], line: 4287, column: null });
        });
    });
});
