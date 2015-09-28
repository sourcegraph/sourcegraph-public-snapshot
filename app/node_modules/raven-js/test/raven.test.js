function flushRavenState() {
    authQueryString = undefined;
    hasJSON = !isUndefined(window.JSON);
    lastCapturedException = undefined;
    lastEventId = undefined;
    globalServer = undefined;
    globalUser = undefined;
    globalProject = undefined;
    globalOptions = {
        logger: 'javascript',
        release: undefined,
        ignoreErrors: [],
        ignoreUrls: [],
        whitelistUrls: [],
        includePaths: [],
        collectWindowErrors: true,
        maxMessageLength: 100,
        tags: {},
        extra: {}
    },
    startTime = 0
    ;

    Raven.uninstall();
}

// window.console must be stubbed in for browsers that don't have it
if (typeof window.console === 'undefined') {
    console = {error: function(){}};
}

var SENTRY_DSN = 'http://abc@example.com:80/2';

function setupRaven() {
    Raven.config(SENTRY_DSN);
}

// patched to return a predictable result
function uuid4() {
    return 'abc123';
}

// patched to be predictable
function now() {
    return 100;
}

describe('TraceKit', function(){
    describe('stacktrace info', function() {
        it('should not remove anonymous functions from the stack', function() {
            // mock up an error object with a stack trace that includes both
            // named functions and anonymous functions
            var stack_str = "" +
                "  Error: \n" +
                "    at new <anonymous> (http://example.com/js/test.js:63)\n" + // stack[0]
                "    at namedFunc0 (http://example.com/js/script.js:10)\n" +    // stack[1]
                "    at http://example.com/js/test.js:65\n" +                   // stack[2]
                "    at namedFunc2 (http://example.com/js/script.js:20)\n" +    // stack[3]
                "    at http://example.com/js/test.js:67\n" +                   // stack[4]
                "    at namedFunc4 (http://example.com/js/script.js:100001)";   // stack[5]
            var mock_err = { stack: stack_str };
            var trace = TraceKit.computeStackTrace.computeStackTraceFromStackProp(mock_err);

            // Make sure TraceKit didn't remove the anonymous functions
            // from the stack like it used to :)
            assert.equal(trace.stack[0].func, 'new <anonymous>');
            assert.equal(trace.stack[1].func, 'namedFunc0');
            assert.equal(trace.stack[2].func, '?');
            assert.equal(trace.stack[3].func, 'namedFunc2');
            assert.equal(trace.stack[4].func, '?');
            assert.equal(trace.stack[5].func, 'namedFunc4');
        });
    });
    describe('error notifications', function(){
        var testMessage = "__mocha_ignore__";
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
            window.onerror = function(message) {
                if (message == testMessage) {
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

describe('globals', function() {
    beforeEach(function() {
        setupRaven();
        globalOptions.fetchContext = true;
    });

    afterEach(function() {
        flushRavenState();
    });

    describe('getHttpData', function() {
        var data = getHttpData();

        it('should have a url', function() {
                assert.equal(data.url, window.location.href);
        });

        it('should have the user-agent header', function() {
            assert.equal(data.headers['User-Agent'], navigator.userAgent);
        });

        it('should have referer header when available', function() {
            // lol this test is awful
            if (window.document.referrer) {
                assert.equal(data.headers.Referer, window.document.referrer);
            } else {
                assert.isUndefined(data.headers.Referer);
            }
        });

    });

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
        });
    });

    describe('isSetup', function() {
        it('should return false with no JSON support', function() {
            globalServer = 'http://localhost/';
            hasJSON = false;
            assert.isFalse(isSetup());
        });

        it('should return false when Raven is not configured', function() {
            hasJSON = true;    // be explicit
            globalServer = undefined;
            this.sinon.stub(window, 'logDebug');
            assert.isFalse(isSetup());
        });

        it('should return true when everything is all gravy', function() {
            hasJSON = true;
            assert.isTrue(isSetup());
        });
    });

    describe('logDebug', function() {
        var level = 'error',
            message = 'foobar';

        it('should not write to console when Raven.debug is false', function() {
            Raven.debug = false;
            this.sinon.stub(console, level);
            logDebug(level, message);
            assert.isFalse(console[level].called);
        });

        it('should write to console when Raven.debug is true', function() {
            Raven.debug = true;
            this.sinon.stub(console, level);
            logDebug(level, message);
            assert.isTrue(console[level].calledOnce);
        });
    });

    describe('setAuthQueryString', function() {
        it('should return a properly formatted string and cache it', function() {
            var expected = '?sentry_version=4&sentry_client=raven-js/<%= pkg.version %>&sentry_key=abc';
            setAuthQueryString();
            assert.strictEqual(authQueryString, expected);
        });
    });

    describe('parseDSN', function() {
        it('should do what it advertises', function() {
            var pieces = parseDSN('http://abc@example.com:80/2');
            assert.strictEqual(pieces.protocol, 'http');
            assert.strictEqual(pieces.user, 'abc');
            assert.strictEqual(pieces.port, '80');
            assert.strictEqual(pieces.path, '/2');
            assert.strictEqual(pieces.host, 'example.com');
        });

        it('should parse protocol relative', function() {
            var pieces = parseDSN('//user@mattrobenolt.com/');
            assert.strictEqual(pieces.protocol, '');
            assert.strictEqual(pieces.user, 'user');
            assert.strictEqual(pieces.port, '');
            assert.strictEqual(pieces.path, '/');
            assert.strictEqual(pieces.host, 'mattrobenolt.com');
        });

        it('should parse domain with hyphen', function() {
            var pieces = parseDSN('http://user@matt-robenolt.com/1');
            assert.strictEqual(pieces.protocol, 'http');
            assert.strictEqual(pieces.user, 'user');
            assert.strictEqual(pieces.port, '');
            assert.strictEqual(pieces.path, '/1');
            assert.strictEqual(pieces.host, 'matt-robenolt.com');
        });

        it('should raise a RavenConfigError when setting a password', function() {
            try {
                parseDSN('http://user:pass@example.com/2');
            } catch(e) {
                return assert.equal(e.name, 'RavenConfigError');
            }
            // shouldn't hit this
            assert.isTrue(false);
        });

        it('should raise a RavenConfigError with an invalid DSN', function() {
            try {
                parseDSN('lol');
            } catch(e) {
                return assert.equal(e.name, 'RavenConfigError');
            }
            // shouldn't hit this
            assert.isTrue(false);
        });
    });

    describe('normalizeFrame', function() {
        it('should handle a normal frame', function() {
            var context = [
                ['line1'],    // pre
                'line2',        // culprit
                ['line3']     // post
            ];
            this.sinon.stub(window, 'extractContextFromFrame').returns(context);
            var frame = {
                url: 'http://example.com/path/file.js',
                line: 10,
                column: 11,
                func: 'lol'
                // context: []    context is stubbed
            };

            globalOptions.fetchContext = true;

            assert.deepEqual(normalizeFrame(frame), {
                filename: 'http://example.com/path/file.js',
                lineno: 10,
                colno: 11,
                'function': 'lol',
                pre_context: ['line1'],
                context_line: 'line2',
                post_context: ['line3'],
                in_app: true
            });
        });

        it('should handle a frame without context', function() {
            this.sinon.stub(window, 'extractContextFromFrame').returns(undefined);
            var frame = {
                url: 'http://example.com/path/file.js',
                line: 10,
                column: 11,
                func: 'lol'
                // context: []    context is stubbed
            };

            globalOptions.fetchContext = true;

            assert.deepEqual(normalizeFrame(frame), {
                filename: 'http://example.com/path/file.js',
                lineno: 10,
                colno: 11,
                'function': 'lol',
                in_app: true
            });
        });

        it('should not mark `in_app` if rules match', function() {
            this.sinon.stub(window, 'extractContextFromFrame').returns(undefined);
            var frame = {
                url: 'http://example.com/path/file.js',
                line: 10,
                column: 11,
                func: 'lol'
                // context: []    context is stubbed
            };

            globalOptions.fetchContext = true;
            globalOptions.includePaths = /^http:\/\/example\.com/;

            assert.deepEqual(normalizeFrame(frame), {
                filename: 'http://example.com/path/file.js',
                lineno: 10,
                colno: 11,
                'function': 'lol',
                in_app: true
            });
        });

        it('should mark `in_app` if rules do not match', function() {
            this.sinon.stub(window, 'extractContextFromFrame').returns(undefined);
            var frame = {
                url: 'http://lol.com/path/file.js',
                line: 10,
                column: 11,
                func: 'lol'
                // context: []    context is stubbed
            };

            globalOptions.fetchContext = true;
            globalOptions.includePaths = /^http:\/\/example\.com/;

            assert.deepEqual(normalizeFrame(frame), {
                filename: 'http://lol.com/path/file.js',
                lineno: 10,
                colno: 11,
                'function': 'lol',
                in_app: false
            });
        });

        it('should mark `in_app` for raven.js', function() {
            this.sinon.stub(window, 'extractContextFromFrame').returns(undefined);
            var frame = {
                url: 'http://lol.com/path/raven.js',
                line: 10,
                column: 11,
                func: 'lol'
                // context: []    context is stubbed
            };

            assert.deepEqual(normalizeFrame(frame), {
                filename: 'http://lol.com/path/raven.js',
                lineno: 10,
                colno: 11,
                'function': 'lol',
                in_app: false
            });
        });

        it('should mark `in_app` for raven.min.js', function() {
            this.sinon.stub(window, 'extractContextFromFrame').returns(undefined);
            var frame = {
                url: 'http://lol.com/path/raven.min.js',
                line: 10,
                column: 11,
                func: 'lol'
                // context: []    context is stubbed
            };

            assert.deepEqual(normalizeFrame(frame), {
                filename: 'http://lol.com/path/raven.min.js',
                lineno: 10,
                colno: 11,
                'function': 'lol',
                in_app: false
            });
        });

        it('should mark `in_app` for Raven', function() {
            this.sinon.stub(window, 'extractContextFromFrame').returns(undefined);
            var frame = {
                url: 'http://lol.com/path/file.js',
                line: 10,
                column: 11,
                func: 'Raven.wrap'
                // context: []    context is stubbed
            };

            assert.deepEqual(normalizeFrame(frame), {
                filename: 'http://lol.com/path/file.js',
                lineno: 10,
                colno: 11,
                'function': 'Raven.wrap',
                in_app: false
            });
        });

        it('should mark `in_app` for TraceKit', function() {
            this.sinon.stub(window, 'extractContextFromFrame').returns(undefined);
            var frame = {
                url: 'http://lol.com/path/file.js',
                line: 10,
                column: 11,
                func: 'TraceKit.lol'
                // context: []    context is stubbed
            };

            assert.deepEqual(normalizeFrame(frame), {
                filename: 'http://lol.com/path/file.js',
                lineno: 10,
                colno: 11,
                'function': 'TraceKit.lol',
                in_app: false
            });
        });
    });

    describe('extractContextFromFrame', function() {
        it('should handle a normal frame', function() {
            var frame = {
                column: 2,
                context: [
                    'line1',
                    'line2',
                    'line3',
                    'line4',
                    'line5',
                    'culprit',
                    'line7',
                    'line8',
                    'line9',
                    'line10',
                    'line11'
                ]
            };
            var context = extractContextFromFrame(frame);
            assert.deepEqual(context, [
                ['line1', 'line2', 'line3', 'line4', 'line5'],
                'culprit',
                ['line7', 'line8', 'line9', 'line10', 'line11']
            ]);
        });

        it('should return nothing if there is no context', function() {
            var frame = {
                column: 2
            };
            assert.isUndefined(extractContextFromFrame(frame));
        });

        it('should reject a context if a line is too long without a column', function() {
            var frame = {
                context: [
                    new Array(1000).join('f')    // generate a line that is 1000 chars long
                ]
            };
            assert.isUndefined(extractContextFromFrame(frame));
        });

        it('should reject a minified context with fetchContext disabled', function() {
            var frame = {
                column: 2,
                context: [
                    'line1',
                    'line2',
                    'line3',
                    'line4',
                    'line5',
                    'culprit',
                    'line7',
                    'line8',
                    'line9',
                    'line10',
                    'line11'
                ]
            };
            globalOptions.fetchContext = false;
            assert.isUndefined(extractContextFromFrame(frame));
        });

        it('should truncate the minified line if there is a column number without sourcemaps enabled', function() {
            // Note to future self:
            // Array(51).join('f').length === 50
            var frame = {
                column: 2,
                context: [
                    'aa' + (new Array(51).join('f')) + (new Array(500).join('z'))
                ]
            };
            assert.deepEqual(extractContextFromFrame(frame), [[], new Array(51).join('f'), []]);
        });
    });

    describe('processException', function() {
        it('should respect `ignoreErrors`', function() {
            this.sinon.stub(window, 'send');

            globalOptions.ignoreErrors = joinRegExp(['e1', 'e2']);
            processException('Error', 'e1', 'http://example.com', []);
            assert.isFalse(window.send.called);
            processException('Error', 'e2', 'http://example.com', []);
            assert.isFalse(window.send.called);
            processException('Error', 'error', 'http://example.com', []);
            assert.isTrue(window.send.calledOnce);
        });

        it('should respect `ignoreUrls`', function() {
            this.sinon.stub(window, 'send');

            globalOptions.ignoreUrls = joinRegExp([/.+?host1.+/, /.+?host2.+/]);
            processException('Error', 'error', 'http://host1/', []);
            assert.isFalse(window.send.called);
            processException('Error', 'error', 'http://host2/', []);
            assert.isFalse(window.send.called);
            processException('Error', 'error', 'http://host3/', []);
            assert.isTrue(window.send.calledOnce);
        });

        it('should respect `whitelistUrls`', function() {
            this.sinon.stub(window, 'send');

            globalOptions.whitelistUrls = joinRegExp([/.+?host1.+/, /.+?host2.+/]);
            processException('Error', 'error', 'http://host1/', []);
            assert.equal(window.send.callCount, 1);
            processException('Error', 'error', 'http://host2/', []);
            assert.equal(window.send.callCount, 2);
            processException('Error', 'error', 'http://host3/', []);
            assert.equal(window.send.callCount, 2);
        });

        it('should send a proper payload with frames', function() {
            this.sinon.stub(window, 'send');

            var frames = [
                {
                    filename: 'http://example.com/file1.js'
                },
                {
                    filename: 'http://example.com/file2.js'
                }
            ], framesFlipped = frames.slice(0);

            framesFlipped.reverse();

            processException('Error', 'lol', 'http://example.com/override.js', 10, frames.slice(0), {});
            assert.deepEqual(window.send.lastCall.args, [{
                exception: {
                    type: 'Error',
                    value: 'lol'
                },
                stacktrace: {
                    frames: framesFlipped
                },
                culprit: 'http://example.com/file1.js',
                message: 'lol at 10'
            }]);

            processException('Error', 'lol', '', 10, frames.slice(0), {});
            assert.deepEqual(window.send.lastCall.args, [{
                exception: {
                    type: 'Error',
                    value: 'lol'
                },
                stacktrace: {
                    frames: framesFlipped
                },
                culprit: 'http://example.com/file1.js',
                message: 'lol at 10'
            }]);

            processException('Error', 'lol', '', 10, frames.slice(0), {extra: 'awesome'});
            assert.deepEqual(window.send.lastCall.args, [{
                exception: {
                    type: 'Error',
                    value: 'lol'
                },
                stacktrace: {
                    frames: framesFlipped
                },
                culprit: 'http://example.com/file1.js',
                message: 'lol at 10',
                extra: 'awesome'
            }]);
        });

        it('should send a proper payload without frames', function() {
            this.sinon.stub(window, 'send');

            processException('Error', 'lol', 'http://example.com/override.js', 10, [], {});
            assert.deepEqual(window.send.lastCall.args, [{
                exception: {
                    type: 'Error',
                    value: 'lol'
                },
                stacktrace: {
                    frames: [{
                        filename: 'http://example.com/override.js',
                        lineno: 10,
                        in_app: true
                    }]
                },
                culprit: 'http://example.com/override.js',
                message: 'lol at 10'
            }]);

            processException('Error', 'lol', 'http://example.com/override.js', 10, [], {});
            assert.deepEqual(window.send.lastCall.args, [{
                exception: {
                    type: 'Error',
                    value: 'lol'
                },
                stacktrace: {
                    frames: [{
                        filename: 'http://example.com/override.js',
                        lineno: 10,
                        in_app: true
                    }]
                },
                culprit: 'http://example.com/override.js',
                message: 'lol at 10'
            }]);

            processException('Error', 'lol', 'http://example.com/override.js', 10, [], {extra: 'awesome'});
            assert.deepEqual(window.send.lastCall.args, [{
                exception: {
                    type: 'Error',
                    value: 'lol'
                },
                stacktrace: {
                    frames: [{
                        filename: 'http://example.com/override.js',
                        lineno: 10,
                        in_app: true
                    }]
                },
                culprit: 'http://example.com/override.js',
                message: 'lol at 10',
                extra: 'awesome'
            }]);
        });

        it('should ignored falsey messages', function() {
            this.sinon.stub(window, 'send');

            processException('Error', '', 'http://example.com', []);
            assert.isFalse(window.send.called);

            processException('TypeError', '', 'http://example.com', []);
            assert.isTrue(window.send.called);
        });

        it('should not blow up with `undefined` message', function() {
            this.sinon.stub(window, 'send');

            processException('TypeError', undefined, 'http://example.com', []);
            assert.isTrue(window.send.called);
        });

        it('should truncate messages to the specified length', function() {
            this.sinon.stub(window, 'send');

            processException('TypeError', new Array(500).join('a'), 'http://example.com', []);
            assert.deepEqual(window.send.lastCall.args, [{
                message: new Array(101).join('a')+'\u2026 at ',
                exception: {
                    type: 'TypeError',
                    value: new Array(101).join('a')+'\u2026'
                },
                stacktrace: {
                    frames: [{
                        filename: 'http://example.com',
                        lineno: [],
                        in_app: true
                    }]
                },
                culprit: 'http://example.com',
            }]);

            globalOptions.maxMessageLength = 150;

            processException('TypeError', new Array(500).join('a'), 'http://example.com', []);
            assert.deepEqual(window.send.lastCall.args, [{
                message: new Array(151).join('a')+'\u2026 at ',
                exception: {
                    type: 'TypeError',
                    value: new Array(151).join('a')+'\u2026'
                },
                stacktrace: {
                    frames: [{
                        filename: 'http://example.com',
                        lineno: [],
                        in_app: true
                    }]
                },
                culprit: 'http://example.com',
            }]);
        });
    });

    describe('send', function() {
        it('should check `isSetup`', function() {
            this.sinon.stub(window, 'isSetup').returns(false);
            this.sinon.stub(window, 'makeRequest');

            send();
            assert.isTrue(window.isSetup.calledOnce);
            assert.isFalse(window.makeRequest.calledOnce);
        });

        it('should build a good data payload', function() {
            this.sinon.stub(window, 'isSetup').returns(true);
            this.sinon.stub(window, 'makeRequest');
            this.sinon.stub(window, 'getHttpData').returns({
                url: 'http://localhost/?a=b',
                headers: {'User-Agent': 'lolbrowser'}
            });

            globalProject = '2';
            globalOptions = {
                logger: 'javascript'
            };

            send({foo: 'bar'});
            assert.deepEqual(window.makeRequest.lastCall.args[0], {
                project: '2',
                logger: 'javascript',
                platform: 'javascript',
                request: {
                    url: 'http://localhost/?a=b',
                    headers: {
                        'User-Agent': 'lolbrowser'
                    }
                },
                event_id: 'abc123',
                foo: 'bar',
                extra: {'session:duration': 100}
            });
        });

        it('should build a good data payload with a User', function() {
            this.sinon.stub(window, 'isSetup').returns(true);
            this.sinon.stub(window, 'makeRequest');
            this.sinon.stub(window, 'getHttpData').returns({
                url: 'http://localhost/?a=b',
                headers: {'User-Agent': 'lolbrowser'}
            });

            globalProject = '2';
            globalOptions = {
                logger: 'javascript'
            };

            globalUser = {name: 'Matt'};

            send({foo: 'bar'});
            assert.deepEqual(window.makeRequest.lastCall.args, [{
                project: '2',
                logger: 'javascript',
                platform: 'javascript',
                request: {
                    url: 'http://localhost/?a=b',
                    headers: {
                        'User-Agent': 'lolbrowser'
                    }
                },
                event_id: 'abc123',
                user: {
                    name: 'Matt'
                },
                foo: 'bar',
                extra: {'session:duration': 100}
            }]);
        });

        it('should merge in global tags', function() {
            this.sinon.stub(window, 'isSetup').returns(true);
            this.sinon.stub(window, 'makeRequest');
            this.sinon.stub(window, 'getHttpData').returns({
                url: 'http://localhost/?a=b',
                headers: {'User-Agent': 'lolbrowser'}
            });

            globalProject = '2';
            globalOptions = {
                logger: 'javascript',
                tags: {tag1: 'value1'}
            };


            send({tags: {tag2: 'value2'}});
            assert.deepEqual(window.makeRequest.lastCall.args, [{
                project: '2',
                logger: 'javascript',
                platform: 'javascript',
                request: {
                    url: 'http://localhost/?a=b',
                    headers: {
                        'User-Agent': 'lolbrowser'
                    }
                },
                event_id: 'abc123',
                tags: {tag1: 'value1', tag2: 'value2'},
                extra: {'session:duration': 100}
            }]);
            assert.deepEqual(globalOptions, {
                logger: 'javascript',
                tags: {tag1: 'value1'}
            });
        });

        it('should merge in global extra', function() {
            this.sinon.stub(window, 'isSetup').returns(true);
            this.sinon.stub(window, 'makeRequest');
            this.sinon.stub(window, 'getHttpData').returns({
                url: 'http://localhost/?a=b',
                headers: {'User-Agent': 'lolbrowser'}
            });

            globalProject = '2';
            globalOptions = {
                logger: 'javascript',
                extra: {key1: 'value1'}
            };


            send({extra: {key2: 'value2'}});
            assert.deepEqual(window.makeRequest.lastCall.args, [{
                project: '2',
                logger: 'javascript',
                platform: 'javascript',
                request: {
                    url: 'http://localhost/?a=b',
                    headers: {
                        'User-Agent': 'lolbrowser'
                    }
                },
                event_id: 'abc123',
                extra: {key1: 'value1', key2: 'value2', 'session:duration': 100}
            }]);
            assert.deepEqual(globalOptions, {
                logger: 'javascript',
                extra: {key1: 'value1'}
            });
        });

        it('should let dataCallback override everything', function() {
            this.sinon.stub(window, 'isSetup').returns(true);
            this.sinon.stub(window, 'makeRequest');

            globalOptions = {
                projectId: 2,
                logger: 'javascript',
                dataCallback: function() {
                    return {lol: 'ibrokeit'};
                }
            };

            globalUser = {name: 'Matt'};

            send({foo: 'bar'});
            assert.deepEqual(window.makeRequest.lastCall.args, [{
                lol: 'ibrokeit',
                event_id: 'abc123',
            }]);
        });

        it('should ignore dataCallback if it does not return anything', function() {
            this.sinon.stub(window, 'isSetup').returns(true);
            this.sinon.stub(window, 'makeRequest');
            this.sinon.stub(window, 'getHttpData').returns({
                url: 'http://localhost/?a=b',
                headers: {'User-Agent': 'lolbrowser'}
            });

            globalProject = '2';
            globalOptions = {
                logger: 'javascript',
                dataCallback: function() {
                    return;
                }
            };

            send({foo: 'bar'});
            assert.deepEqual(window.makeRequest.lastCall.args[0], {
                project: '2',
                logger: 'javascript',
                platform: 'javascript',
                request: {
                    url: 'http://localhost/?a=b',
                    headers: {
                        'User-Agent': 'lolbrowser'
                    }
                },
                event_id: 'abc123',
                foo: 'bar',
                extra: {'session:duration': 100}
            });
        });

        it('should strip empty tags', function() {
            this.sinon.stub(window, 'isSetup').returns(true);
            this.sinon.stub(window, 'makeRequest');
            this.sinon.stub(window, 'getHttpData').returns({
                url: 'http://localhost/?a=b',
                headers: {'User-Agent': 'lolbrowser'}
            });

            globalOptions = {
                projectId: 2,
                logger: 'javascript',
                tags: {}
            };

            send({foo: 'bar', tags: {}, extra: {}});
            assert.deepEqual(window.makeRequest.lastCall.args[0], {
                project: '2',
                logger: 'javascript',
                platform: 'javascript',
                request: {
                    url: 'http://localhost/?a=b',
                    headers: {
                        'User-Agent': 'lolbrowser'
                    }
                },
                event_id: 'abc123',
                foo: 'bar',
                extra: {'session:duration': 100}
            });
        });

        it('should attach release if available', function() {
            this.sinon.stub(window, 'isSetup').returns(true);
            this.sinon.stub(window, 'makeRequest');
            this.sinon.stub(window, 'getHttpData').returns({
                url: 'http://localhost/?a=b',
                headers: {'User-Agent': 'lolbrowser'}
            });

            globalOptions = {
                projectId: 2,
                logger: 'javascript',
                release: 'abc123',
            };

            send({foo: 'bar'});
            assert.deepEqual(window.makeRequest.lastCall.args[0], {
                project: '2',
                release: 'abc123',
                logger: 'javascript',
                platform: 'javascript',
                request: {
                    url: 'http://localhost/?a=b',
                    headers: {
                        'User-Agent': 'lolbrowser'
                    }
                },
                event_id: 'abc123',
                foo: 'bar',
                extra: {'session:duration': 100}
            });
        });
    });

    describe('makeRequest', function() {
        it('should load an Image', function() {
            authQueryString = '?lol';
            globalServer = 'http://localhost/';
            var imageCache = [];
            this.sinon.stub(window, 'newImage', function(){ var img = {}; imageCache.push(img); return img; });

            makeRequest({foo: 'bar'});
            assert.equal(imageCache.length, 1);
            assert.equal(imageCache[0].src, 'http://localhost/?lol&sentry_data=%7B%22foo%22%3A%22bar%22%7D');
        });
    });

    describe('handleStackInfo', function() {
        it('should work as advertised', function() {
            var frame = {url: 'http://example.com'};
            this.sinon.stub(window, 'normalizeFrame').returns(frame);
            this.sinon.stub(window, 'processException');

            var stackInfo = {
                name: 'Matt',
                message: 'hey',
                url: 'http://example.com',
                lineno: 10,
                stack: [
                    frame, frame
                ]
            };

            handleStackInfo(stackInfo, {foo: 'bar'});
            assert.deepEqual(window.processException.lastCall.args, [
                'Matt', 'hey', 'http://example.com', 10, [frame, frame], {foo: 'bar'}
            ]);
        });

        it('should work as advertised #integration', function() {
            this.sinon.stub(window, 'makeRequest');
            var stackInfo = {
                name: 'Error',
                message: 'crap',
                url: 'http://example.com',
                lineno: 10,
                stack: [
                    {
                        url: 'http://example.com/file1.js',
                        line: 10,
                        column: 11,
                        func: 'broken',
                        context: [
                            'line1',
                            'line2',
                            'line3'
                        ]
                    },
                    {
                        url: 'http://example.com/file2.js',
                        line: 12,
                        column: 13,
                        func: 'lol',
                        context: [
                            'line4',
                            'line5',
                            'line6'
                        ]
                    }
                ]
            };

            handleStackInfo(stackInfo, {foo: 'bar'});
            assert.isTrue(window.makeRequest.calledOnce);
            /* This is commented out because chai is broken.

            assert.deepEqual(window.makeRequest.lastCall.args, [{
                project: '2',
                logger: 'javascript',
                platform: 'javascript',
                request: {
                    url: window.location.protocol + '//' + window.location.host + window.location.pathname,
                    querystring: window.location.search.slice(1)
                },
                exception: {
                    type: 'Error',
                    value: 'crap'
                },
                stacktrace: {
                    frames: [{
                        filename: 'http://example.com/file1.js',
                        filename: 'file1.js',
                        lineno: 10,
                        colno: 11,
                        'function': 'broken',
                        post_context: ['line3'],
                        context_line: 'line2',
                        pre_context: ['line1']
                    }, {
                        filename: 'http://example.com/file2.js',
                        filename: 'file2.js',
                        lineno: 12,
                        colno: 13,
                        'function': 'lol',
                        post_context: ['line6'],
                        context_line: 'line5',
                        pre_context: ['line4']
                    }]
                },
                culprit: 'http://example.com',
                message: 'crap at 10',
                foo: 'bar'
            }]);
            */
        });

        it('should ignore frames that dont have a url', function() {
            this.sinon.stub(window, 'normalizeFrame').returns(undefined);
            this.sinon.stub(window, 'processException');

            var stackInfo = {
                name: 'Matt',
                message: 'hey',
                url: 'http://example.com',
                lineno: 10,
                stack: new Array(2)
            };

            handleStackInfo(stackInfo, {foo: 'bar'});
            assert.deepEqual(window.processException.lastCall.args, [
                'Matt', 'hey', 'http://example.com', 10, [], {foo: 'bar'}
            ]);
        });

        it('should not shit when there is no stack object from TK', function() {
            this.sinon.stub(window, 'normalizeFrame').returns(undefined);
            this.sinon.stub(window, 'processException');

            var stackInfo = {
                name: 'Matt',
                message: 'hey',
                url: 'http://example.com',
                lineno: 10
                // stack: new Array(2)
            };

            handleStackInfo(stackInfo);
            assert.isFalse(window.normalizeFrame.called);
            assert.deepEqual(window.processException.lastCall.args, [
                'Matt', 'hey', 'http://example.com', 10, [], undefined
            ]);
        });

        it('should detect 2-words patterns (angularjs frequent case)', function() {
            this.sinon.stub(window, 'normalizeFrame').returns(undefined);
            this.sinon.stub(window, 'processException');

            var stackInfo = {
                name: 'new <anonymous>',
                message: 'hey',
                url: 'http://example.com',
                lineno: 10
                // stack: new Array(2)
            };

            handleStackInfo(stackInfo);
            assert.isFalse(window.normalizeFrame.called);
            assert.deepEqual(window.processException.lastCall.args, [
                'new <anonymous>', 'hey', 'http://example.com', 10, [], undefined
            ]);
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
});

describe('Raven (public API)', function() {
    afterEach(function() {
        flushRavenState();
    });

    describe('.VERSION', function() {
        it('should have a version', function() {
            assert.isString(Raven.VERSION);
        });
    });

    describe('ignore errors', function() {
        it('should install default ignore errors', function() {
            Raven.config('//abc@example.com/2');

            assert.isTrue(globalOptions.ignoreErrors.test('Script error'), 'it should install "Script error" by default');
            assert.isTrue(globalOptions.ignoreErrors.test('Script error.'), 'it should install "Script error." by default');
            assert.isTrue(globalOptions.ignoreErrors.test('Javascript error: Script error on line 0'), 'it should install "Javascript error: Script error on line 0" by default');
            assert.isTrue(globalOptions.ignoreErrors.test('Javascript error: Script error. on line 0'), 'it should install "Javascript error: Script error. on line 0" by default');
        });
    });

    describe('callback function', function() {
        it('should callback a function if it is global', function() {
            window.RavenConfig = {
                dsn: "http://random@some.other.server:80/2",
                config: {some: 'config'}
            };

            this.sinon.stub(window, 'isSetup').returns(false);
            this.sinon.stub(TraceKit.report, 'subscribe');

            afterLoad();

            assert.equal(globalKey, 'random');
            assert.equal(globalServer, 'http://some.other.server:80/api/2/store/');

            assert.equal(globalOptions.some, 'config');
            assert.equal(globalProject, '2');

            assert.isTrue(window.isSetup.calledOnce);
            assert.isFalse(TraceKit.report.subscribe.calledOnce);

            delete window.RavenConfig;
        });
    });

    describe('.config', function() {
        it('should work with a DSN', function() {
            assert.equal(Raven, Raven.config(SENTRY_DSN, {foo: 'bar'}), 'it should return Raven');
            assert.equal(globalKey, 'abc');
            assert.equal(globalServer, 'http://example.com:80/api/2/store/');
            assert.equal(globalOptions.foo, 'bar');
            assert.equal(globalProject, '2');
            assert.isTrue(isSetup());
        });

        it('should work with a protocol relative DSN', function() {
            Raven.config('//abc@example.com/2');
            assert.equal(globalKey, 'abc');
            assert.equal(globalServer, '//example.com/api/2/store/');
            assert.equal(globalProject, '2');
            assert.isTrue(isSetup());
        });

        it('should work should work at a non root path', function() {
            Raven.config('//abc@example.com/sentry/2');
            assert.equal(globalKey, 'abc');
            assert.equal(globalServer, '//example.com/sentry/api/2/store/');
            assert.equal(globalProject, '2');
            assert.isTrue(isSetup());
        });

        it('should noop a falsey dsn', function() {
            Raven.config('');
            assert.isFalse(isSetup());
        });

        it('should return Raven for a falsey dsn', function() {
            assert.equal(Raven.config(''), Raven);
        });

        it('should not set global options more than once', function() {
            this.sinon.spy(window, 'parseDSN');
            this.sinon.stub(window, 'logDebug');
            setupRaven();
            setupRaven();
            assert.isTrue(parseDSN.calledOnce);
            assert.isTrue(logDebug.called);
        });

        describe('whitelistUrls', function() {
            it('should be false if none are passed', function() {
                Raven.config('//abc@example.com/2');
                assert.equal(globalOptions.whitelistUrls, false);
            });

            it('should join into a single RegExp', function() {
                Raven.config('//abc@example.com/2', {
                    whitelistUrls: [
                        /my.app/i,
                        /other.app/i
                    ]
                });

                assert.match(globalOptions.whitelistUrls, /my.app|other.app/i);
            });

            it('should handle strings as well', function() {
                Raven.config('//abc@example.com/2', {
                    whitelistUrls: [
                        /my.app/i,
                        "stringy.app"
                    ]
                });

                assert.match(globalOptions.whitelistUrls, /my.app|stringy.app/i);
            });
        });

        describe('collectWindowErrors', function() {
            it('should be true by default', function() {
                Raven.config(SENTRY_DSN);
                assert.isTrue(TraceKit.collectWindowErrors);
            });

            it('should be true if set to true', function() {
                Raven.config(SENTRY_DSN, {
                    collectWindowErrors: true
                });

                assert.isTrue(TraceKit.collectWindowErrors);
            });

            it('should be false if set to false', function() {
                Raven.config(SENTRY_DSN, {
                    collectWindowErrors: false
                });

                assert.isFalse(TraceKit.collectWindowErrors);
            });
        });
    });

    describe('.install', function() {
        it('should check `isSetup`', function() {
            this.sinon.stub(window, 'isSetup').returns(false);
            this.sinon.stub(TraceKit.report, 'subscribe');
            Raven.install();
            assert.isTrue(window.isSetup.calledOnce);
            assert.isFalse(TraceKit.report.subscribe.calledOnce);
        });

        it('should register itself with TraceKit', function() {
            this.sinon.stub(window, 'isSetup').returns(true);
            this.sinon.stub(TraceKit.report, 'subscribe');
            assert.equal(Raven, Raven.install());
            assert.isTrue(TraceKit.report.subscribe.calledOnce);
            assert.equal(TraceKit.report.subscribe.lastCall.args[0], handleStackInfo);
        });

        it('should not register itself more than once', function() {
            this.sinon.stub(window, 'isSetup').returns(true);
            this.sinon.stub(TraceKit.report, 'subscribe');
            Raven.install();
            Raven.install();
            assert.isTrue(TraceKit.report.subscribe.calledOnce);
        });
    });

    describe('.wrap', function() {
        it('should return a wrapped callback', function() {
            var spy = this.sinon.spy();
            var wrapped = Raven.wrap(spy);
            assert.isFunction(wrapped);
            assert.isTrue(wrapped.__raven__);
            wrapped();
            assert.isTrue(spy.calledOnce);
        });

        it('should copy property when wrapping function', function() {
            var func = function() {};
            func.test = true;
            var wrapped = Raven.wrap(func);
            assert.isTrue(wrapped.test);
        });

        it('should not copy prototype property when wrapping function', function() {
            var func = function() {};
            func.prototype.test = true;
            var wrapped = Raven.wrap(func);
            assert.isUndefined(new wrapped().test);
        });

        it('should return the result of a wrapped function', function() {
            var func = function() { return 'foo'; };
            var wrapped = Raven.wrap(func);
            assert.equal(wrapped(), 'foo');
        });

        it('should not wrap a non-function', function() {
            assert.equal(Raven.wrap('lol'), 'lol');
            assert.equal(Raven.wrap({}, 'lol'), 'lol');
            assert.equal(Raven.wrap(undefined, 'lol'), 'lol');
            var a = [1, 2];
            assert.equal(Raven.wrap(a), a);
        });

        it('should wrap function arguments', function() {
            var spy = this.sinon.spy();
            var wrapped = Raven.wrap(function(f) {
                assert.isTrue(f.__raven__);
                f();
            });
            wrapped(spy);
            assert.isTrue(spy.calledOnce);
        });

        it('should not wrap function arguments', function() {
            var spy = this.sinon.spy();
            var wrapped = Raven.wrap({ deep: false }, function(f) {
                assert.isUndefined(f.__raven__);
                f();
            });
            wrapped(spy);
            assert.isTrue(spy.calledOnce);
        });

        it('should maintain the correct scope', function() {
            var foo = {};
            var bar = function() {
                assert.equal(this, foo);
            };
            bar.apply(foo, []);
            Raven.wrap(bar).apply(foo, []);
        });

        it('should re-raise a thrown exception', function() {
            var error = new Error('lol');
            assert.throws(function() {
                Raven.wrap(function() { throw error; })();
            }, error);
        });

    });

    describe('.context', function() {
        it('should execute the callback with options', function() {
            var spy = this.sinon.spy();
            this.sinon.stub(Raven, 'captureException');
            Raven.context({'foo': 'bar'}, spy);
            assert.isTrue(spy.calledOnce);
            assert.isFalse(Raven.captureException.called);
        });

        it('should execute the callback with arguments', function() {
            var spy = this.sinon.spy();
            var args = [1, 2];
            Raven.context(spy, args);
            assert.deepEqual(spy.lastCall.args, args);
        });

        it('should execute the callback without options', function() {
            var spy = this.sinon.spy();
            this.sinon.stub(Raven, 'captureException');
            Raven.context(spy);
            assert.isTrue(spy.calledOnce);
            assert.isFalse(Raven.captureException.called);
        });

        it('should capture the exception with options', function() {
            var error = new Error('crap');
            var broken = function() { throw error; };
            this.sinon.stub(Raven, 'captureException');
            assert.throws(function() {
                Raven.context({foo: 'bar'}, broken);
            }, error);
            assert.isTrue(Raven.captureException.called);
            assert.deepEqual(Raven.captureException.lastCall.args, [error, {'foo': 'bar'}]);
        });

        it('should capture the exception without options', function() {
            var error = new Error('crap');
            var broken = function() { throw error; };
            this.sinon.stub(Raven, 'captureException');
            assert.throws(function() {
                Raven.context(broken);
            }, error);
            assert.isTrue(Raven.captureException.called);
            assert.deepEqual(Raven.captureException.lastCall.args, [error, undefined]);
        });

        it('should execute the callback without arguments', function() {
            // This is only reproducable in a browser that complains about passing
            // undefined to Function.apply
            var spy = this.sinon.spy();
            Raven.context(spy);
            assert.deepEqual(spy.lastCall.args, []);
        });

        it('should return the result of the wrapped function', function() {
            var val = {};
            var func = function() { return val; };
            assert.equal(Raven.context(func), val);
        });
    });

    describe('.uninstall', function() {
        it('should uninstall from TraceKit', function() {
            this.sinon.stub(TraceKit.report, 'uninstall');
            Raven.uninstall();
            assert.isTrue(TraceKit.report.uninstall.calledOnce);
        });

        it('should set isRavenInstalled flag to false', function() {
            isRavenInstalled = true;
            this.sinon.stub(TraceKit.report, 'uninstall');
            Raven.uninstall();
            assert.isFalse(isRavenInstalled);
        });
    });

    describe('.setUserContext', function() {
        it('should set the globalUser object', function() {
            Raven.setUserContext({name: 'Matt'});
            assert.deepEqual(globalUser, {name: 'Matt'});
        });

        it('should clear the globalUser with no arguments', function() {
            globalUser = {name: 'Matt'};
            Raven.setUserContext();
            assert.isUndefined(globalUser);
        });
    });

    describe('.setExtraContext', function() {
        it('should set the globalOptions.extra object', function() {
            Raven.setExtraContext({name: 'Matt'});
            assert.deepEqual(globalOptions.extra, {name: 'Matt'});
        });

        it('should clear globalOptions.extra with no arguments', function() {
            globalOptions = {name: 'Matt'};
            Raven.setExtraContext();
            assert.deepEqual(globalOptions.extra, {});
        });
    });

    describe('.setTagsContext', function() {
        it('should set the globalOptions.tags object', function() {
            Raven.setTagsContext({name: 'Matt'});
            assert.deepEqual(globalOptions.tags, {name: 'Matt'});
        });

        it('should clear globalOptions.tags with no arguments', function() {
            globalOptions = {name: 'Matt'};
            Raven.setTagsContext();
            assert.deepEqual(globalOptions.tags, {});
        });
    });

    describe('.setReleaseContext', function() {
        it('should set the globalOptions.release attribute', function() {
            Raven.setReleaseContext('abc123');
            assert.equal(globalOptions.release, 'abc123');
        });

        it('should clear globalOptions.release with no arguments', function() {
            globalOptions.release = 'abc123';
            Raven.setReleaseContext();
            assert.isUndefined(globalOptions.release);
        });
    });

    describe('.setDataCallback', function() {
        it('should set the globalOptions.dataCallback attribute', function() {
            var foo = function(){};
            Raven.setDataCallback(foo);
            assert.equal(globalOptions.dataCallback, foo);
        });

        it('should clear globalOptions.dataCallback with no arguments', function() {
            var foo = function(){};
            globalOptions.dataCallback = foo;
            Raven.setDataCallback();
            assert.isUndefined(globalOptions.dataCallback);
        });
    });

    describe('.setShouldSendCallback', function() {
        it('should set the globalOptions.shouldSendCallback attribute', function() {
            var foo = function(){};
            Raven.setShouldSendCallback(foo);
            assert.equal(globalOptions.shouldSendCallback, foo);
        });

        it('should clear globalOptions.shouldSendCallback with no arguments', function() {
            var foo = function(){};
            globalOptions.shouldSendCallback = foo;
            Raven.setShouldSendCallback();
            assert.isUndefined(globalOptions.shouldSendCallback);
        });
    });

    describe('.captureMessage', function() {
        it('should work as advertised', function() {
            this.sinon.stub(window, 'send');
            Raven.captureMessage('lol', {foo: 'bar'});
            assert.deepEqual(window.send.lastCall.args, [{
                message: 'lol',
                foo: 'bar'
            }]);
        });

        it('should coerce message to a string', function() {
            this.sinon.stub(window, 'send');
            Raven.captureMessage({});
            assert.deepEqual(window.send.lastCall.args, [{
                message: '[object Object]'
            }]);
        });

        it('should work as advertised #integration', function() {
            var imageCache = [];
            this.sinon.stub(window, 'newImage', function(){ var img = {}; imageCache.push(img); return img; });

            setupRaven();
            Raven.captureMessage('lol', {foo: 'bar'});
            assert.equal(imageCache.length, 1);
            // It'd be hard to assert the actual payload being sent
            // since it includes the generated url, which is going to
            // vary between users running the tests
            // Unit tests should cover that the payload was constructed properly
        });

        it('should tag lastEventId #integration', function() {
            setupRaven();
            Raven.captureMessage('lol');
            assert.equal(Raven.lastEventId(), 'abc123');
        });

        it('should respect `ignoreErrors`', function() {
            this.sinon.stub(window, 'send');

            globalOptions.ignoreErrors = joinRegExp(['e1', 'e2']);
            Raven.captureMessage('e1');
            assert.isFalse(window.send.called);
            Raven.captureMessage('e2');
            assert.isFalse(window.send.called);
            Raven.captureMessage('Non-ignored error');
            assert.isTrue(window.send.calledOnce);
        });
    });

    describe('.captureException', function() {
        it('should call TraceKit.report', function() {
            var error = new Error('crap');
            this.sinon.stub(TraceKit, 'report');
            Raven.captureException(error, {foo: 'bar'});
            assert.isTrue(TraceKit.report.calledOnce);
            assert.deepEqual(TraceKit.report.lastCall.args, [error, {foo: 'bar'}]);
        });

        it('should store the last exception', function() {
            var error = new Error('crap');
            this.sinon.stub(TraceKit, 'report');
            Raven.captureException(error);
            assert.equal(Raven.lastException(), error);
        });

        it('shouldn\'t reraise the if the error is the same error', function() {
            var error = new Error('crap');
            this.sinon.stub(TraceKit, 'report').throws(error);
            // this would raise if the errors didn't match
            Raven.captureException(error, {foo: 'bar'});
            assert.isTrue(TraceKit.report.calledOnce);
        });

        it('should reraise a different error', function() {
            var error = new Error('crap1');
            this.sinon.stub(TraceKit, 'report').throws(error);
            assert.throws(function() {
                Raven.captureException(new Error('crap2'));
            }, error);
        });

        it('should capture as a normal message if a non-Error is passed', function() {
            this.sinon.stub(Raven, 'captureMessage');
            this.sinon.stub(TraceKit, 'report');
            Raven.captureException('derp');
            assert.equal(Raven.captureMessage.lastCall.args[0], 'derp');
            assert.isFalse(TraceKit.report.called);
            Raven.captureException(true);
            assert.equal(Raven.captureMessage.lastCall.args[0], true);
            assert.isFalse(TraceKit.report.called);
        });
    });

    describe('.isSetup', function() {
        it('should work as advertised', function() {
            var isSetup = this.sinon.stub(window, 'isSetup');
            isSetup.returns(true);
            assert.isTrue(Raven.isSetup());
            isSetup.returns(false);
            assert.isFalse(Raven.isSetup());
        });
    });
});
