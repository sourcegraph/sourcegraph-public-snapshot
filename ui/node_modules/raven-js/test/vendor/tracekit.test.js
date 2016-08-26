/*global Mocha, assert*/
'use strict';

var TraceKit = require('../../vendor/TraceKit/tracekit');

describe('TraceKit', function(){
    describe('stacktrace info', function() {
        it('should not remove anonymous functions from the stack', function() {
            // mock up an error object with a stack trace that includes both
            // named functions and anonymous functions
            var stack_str = "" +
                "  Error: \n" +
                "    at new <anonymous> (http://example.com/js/test.js:63:1)\n" +   // stack[0]
                "    at namedFunc0 (http://example.com/js/script.js:10:2)\n" +      // stack[1]
                "    at http://example.com/js/test.js:65:10\n" +                    // stack[2]
                "    at namedFunc2 (http://example.com/js/script.js:20:5)\n" +      // stack[3]
                "    at http://example.com/js/test.js:67:5\n" +                     // stack[4]
                "    at namedFunc4 (http://example.com/js/script.js:100001:10002)"; // stack[5]
            var mock_err = { stack: stack_str };
            var trace = TraceKit.computeStackTrace.computeStackTraceFromStackProp(mock_err);

            // Make sure TraceKit didn't remove the anonymous functions
            // from the stack like it used to :)
            assert.equal(trace.stack[0].func, 'new <anonymous>');
            assert.equal(trace.stack[0].url, 'http://example.com/js/test.js');
            assert.equal(trace.stack[0].line, 63);
            assert.equal(trace.stack[0].column, 1);

            assert.equal(trace.stack[1].func, 'namedFunc0');
            assert.equal(trace.stack[1].url, 'http://example.com/js/script.js');
            assert.equal(trace.stack[1].line, 10);
            assert.equal(trace.stack[1].column, 2);

            assert.equal(trace.stack[2].func, '?');
            assert.equal(trace.stack[2].url, 'http://example.com/js/test.js');
            assert.equal(trace.stack[2].line, 65);
            assert.equal(trace.stack[2].column, 10);

            assert.equal(trace.stack[3].func, 'namedFunc2');
            assert.equal(trace.stack[3].url, 'http://example.com/js/script.js');
            assert.equal(trace.stack[3].line, 20);
            assert.equal(trace.stack[3].column, 5);

            assert.equal(trace.stack[4].func, '?');
            assert.equal(trace.stack[4].url, 'http://example.com/js/test.js');
            assert.equal(trace.stack[4].line, 67);
            assert.equal(trace.stack[4].column, 5);

            assert.equal(trace.stack[5].func, 'namedFunc4');
            assert.equal(trace.stack[5].url, 'http://example.com/js/script.js');
            assert.equal(trace.stack[5].line, 100001);
            assert.equal(trace.stack[5].column, 10002);
        });

        it('should handle eval/anonymous strings in Chrome 46', function () {
            var stack_str = "" +
                "ReferenceError: baz is not defined\n" +
                "   at bar (http://example.com/js/test.js:19:7)\n" +
                "   at foo (http://example.com/js/test.js:23:7)\n" +
                "   at eval (eval at <anonymous> (http://example.com/js/test.js:26:5), <anonymous>:1:26)\n";

            var mock_err = { stack: stack_str };
            var trace = TraceKit.computeStackTrace.computeStackTraceFromStackProp(mock_err);
            assert.equal(trace.stack[0].func, 'bar');
            assert.equal(trace.stack[0].url, 'http://example.com/js/test.js');
            assert.equal(trace.stack[0].line, 19);
            assert.equal(trace.stack[0].column, 7);

            assert.equal(trace.stack[1].func, 'foo');
            assert.equal(trace.stack[1].url, 'http://example.com/js/test.js');
            assert.equal(trace.stack[1].line, 23);
            assert.equal(trace.stack[1].column, 7);

            assert.equal(trace.stack[2].func, 'eval');
            // TODO: fix nested evals
            assert.equal(trace.stack[2].url, 'eval at <anonymous> (http://example.com/js/test.js:26:5), <anonymous>');
            assert.equal(trace.stack[2].line, 1); // second set of line/column numbers used
            assert.equal(trace.stack[2].column, 26);
        });
    });

    describe('.computeStackTrace', function() {
        it('should handle a native error object', function() {
            var ex = new Error('test');
            var stack = TraceKit.computeStackTrace(ex);
            assert.deepEqual(stack.name, 'Error');
            assert.deepEqual(stack.message, 'test');
        });

        it('should handle a native error object stack from Chrome', function() {
            var stackStr = "" +
            "Error: foo\n" +
            "    at <anonymous>:2:11\n" +
            "    at Object.InjectedScript._evaluateOn (<anonymous>:904:140)\n" +
            "    at Object.InjectedScript._evaluateAndWrap (<anonymous>:837:34)\n" +
            "    at Object.InjectedScript.evaluate (<anonymous>:693:21)";
            var mockErr = {
                name: 'Error',
                message: 'foo',
                stack: stackStr
            };
            var trace = TraceKit.computeStackTrace(mockErr);
            assert.deepEqual(trace.stack[0].url, '<anonymous>');
        });
    });

    describe('error notifications', function(){
        var testMessage = "__mocha_ignore__";
        var testLineNo = 1337;

        var subscriptionHandler;
        // TraceKit waits 2000ms for window.onerror to fire, so give the tests
        // some extra time.
        this.timeout(3000);

        before(function() {
            // Prevent the onerror call that's part of our tests from getting to
            // mocha's handler, which would treat it as a test failure.
            //
            // We set this up here and don't ever restore the old handler, because
            // we can't do that without clobbering TraceKit's handler, which can only
            // be installed once.
            var oldOnError = window.onerror;
            window.onerror = function(message, url, lineNo) {
                if (message == testMessage || lineNo === testLineNo) {
                    return true;
                }
                return oldOnError.apply(this, arguments);
            };
        });

        afterEach(function() {
            if (subscriptionHandler) {
                TraceKit.report.unsubscribe(subscriptionHandler);
                subscriptionHandler = null;
            }
        });

        describe('with undefined arguments', function () {
            it('should pass undefined:undefined', function () {
                // this is probably not good behavior;  just writing this test to verify
                // that it doesn't change unintentionally
                subscriptionHandler = function (stackInfo, extra) {
                    assert.equal(stackInfo.name, undefined);
                    assert.equal(stackInfo.message, undefined);
                };
                TraceKit.report.subscribe(subscriptionHandler);
                window.onerror(undefined, undefined, testLineNo);
            });
        });
        describe('when no 5th argument (error object)', function () {
            it('should seperate name, message for default error types (e.g. ReferenceError)', function (done) {
                subscriptionHandler = function (stackInfo, extra) {
                    assert.equal(stackInfo.name, 'ReferenceError');
                    assert.equal(stackInfo.message, 'foo is undefined');
                };
                TraceKit.report.subscribe(subscriptionHandler);
                // should work with/without "Uncaught"
                window.onerror('Uncaught ReferenceError: foo is undefined', 'http://example.com', testLineNo);
                window.onerror('ReferenceError: foo is undefined', 'http://example.com', testLineNo);
                done();
            });

            it('should separate name, message for default error types on Opera Mini (see #546)', function (done) {
                subscriptionHandler = function (stackInfo, extra) {
                    assert.equal(stackInfo.name, 'ReferenceError');
                    assert.equal(stackInfo.message, 'Undefined variable: foo');
                };
                TraceKit.report.subscribe(subscriptionHandler);
                window.onerror('Uncaught exception: ReferenceError: Undefined variable: foo', 'http://example.com', testLineNo);
                done();
            });

            it('should ignore unknown error types', function (done) {
                // TODO: should we attempt to parse this?
                subscriptionHandler = function (stackInfo, extra) {
                    assert.equal(stackInfo.name, undefined);
                    assert.equal(stackInfo.message, 'CustomError: woo scary');
                    done();
                };
                TraceKit.report.subscribe(subscriptionHandler);
                window.onerror('CustomError: woo scary', 'http://example.com', testLineNo);
            });

            it('should ignore arbitrary messages passed through onerror', function (done) {
                subscriptionHandler = function (stackInfo, extra) {
                    assert.equal(stackInfo.name, undefined);
                    assert.equal(stackInfo.message, 'all work and no play makes homer: something something');
                    done();
                };
                TraceKit.report.subscribe(subscriptionHandler);
                window.onerror('all work and no play makes homer: something something', 'http://example.com', testLineNo);
            });
        });

        function testErrorNotification(collectWindowErrors, callOnError, numReports, done) {
            var extraVal = "foo";
            var numDone = 0;
            // TraceKit's collectWindowErrors flag shouldn't affect direct calls
            // to TraceKit.report, so we parameterize it for the tests.
            TraceKit.collectWindowErrors = collectWindowErrors;

            subscriptionHandler = function(stackInfo, extra) {
                assert.equal(extra, extraVal);
                numDone++;
                if (numDone == numReports) {
                    done();
                }
            };
            TraceKit.report.subscribe(subscriptionHandler);

            // TraceKit.report always throws an exception in order to trigger
            // window.onerror so it can gather more stack data. Mocha treats
            // uncaught exceptions as errors, so we catch it via assert.throws
            // here (and manually call window.onerror later if appropriate).
            //
            // We test multiple reports because TraceKit has special logic for when
            // report() is called a second time before either a timeout elapses or
            // window.onerror is called (which is why we always call window.onerror
            // only once below, after all calls to report()).
            for (var i=0; i < numReports; i++) {
                var e = new Error('testing');
                assert.throws(function() {
                    TraceKit.report(e, extraVal);
                }, e);
            }
            // The call to report should work whether or not window.onerror is
            // triggered, so we parameterize it for the tests. We only call it
            // once, regardless of numReports, because the case we want to test for
            // multiple reports is when window.onerror is *not* called between them.
            if (callOnError) {
                window.onerror(testMessage);
            }
        }

        Mocha.utils.forEach([false, true], function(collectWindowErrors) {
            Mocha.utils.forEach([false, true], function(callOnError) {
                Mocha.utils.forEach([1, 2], function(numReports) {
                    it('it should receive arguments from report() when' +
                       ' collectWindowErrors is ' + collectWindowErrors +
                       ' and callOnError is ' + callOnError +
                       ' and numReports is ' + numReports, function(done) {
                        testErrorNotification(collectWindowErrors, callOnError, numReports, done);
                    });
                });
            });
        });
    });
});
