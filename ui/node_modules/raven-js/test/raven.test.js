/*jshint mocha:true*/
/*global assert:false, console:true*/
'use strict';

var proxyquire = require('proxyquireify')(require);

var TraceKit = require('../vendor/TraceKit/tracekit');

var _Raven = proxyquire('../src/raven', {
    // Ensure same TraceKit obj is shared (without specifying this, proxyquire
    // seems to clone dependencies or something weird)
    '../vendor/TraceKit/tracekit': TraceKit
});

_Raven.prototype._getUuid = function () {
    return 'abc123';
};

var joinRegExp = _Raven.utils.joinRegExp;

// window.console must be stubbed in for browsers that don't have it
if (typeof window.console === 'undefined') {
    console = {error: function(){}};
}

var SENTRY_DSN = 'http://abc@example.com:80/2';

function setupRaven() {
    Raven.config(SENTRY_DSN);
}

var Raven;

describe('globals', function() {
    beforeEach(function() {
        this.clock = sinon.useFakeTimers();
        this.clock.tick(0); // Raven initialized at time "0"
        Raven = new _Raven();
        setupRaven();

        this.clock.tick(100); // tick 100 ms
    });

    afterEach(function () {
        this.clock.restore();
    });

    describe('getHttpData', function() {
        var data;

        beforeEach(function () {
            data = Raven._getHttpData();
        });

        describe('with document', function() {
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

        // describe('without document', function () {
        //     it('should return undefined if no document', function () {
        //         hasDocument = false;
        //         var data = getHttpData();
        //         assert.isUndefined(data);
        //     });
        // });
    });

    describe('trimPacket', function() {
        it('should work as advertised', function() {
            Raven._globalOptions.maxMessageLength = 3;
            assert.deepEqual(
                Raven._trimPacket({message: 'lol'}),
                {message: 'lol'}
            );
            assert.deepEqual(
                Raven._trimPacket({message: 'lolol'}),
                {message: 'lol\u2026'}
            );
            assert.deepEqual(
                Raven._trimPacket({message: 'lol', exception: {values: [{value: 'lol'}]}}),
                {message: 'lol', exception: {values: [{value: 'lol'}]}}
            );
            assert.deepEqual(
                Raven._trimPacket({message: 'lolol', exception: {values: [{value: 'lolol'}]}}),
                {message: 'lol\u2026', exception: {values: [{value: 'lol\u2026'}]}}
            );
        });
    });

    describe('isSetup', function() {
        beforeEach(function () {
          this.sinon.stub(Raven, '_logDebug');
        });

        it('should return false with no JSON support', function() {
            Raven._globalServer = 'http://localhost/';
            Raven._hasJSON = false;

            assert.isFalse(Raven.isSetup());
        });

        describe('when Raven is not configured', function () {
            it('should return false when Raven is not configured', function() {
                Raven._hasJSON = true;    // be explicit
                Raven._globalServer = undefined;

                assert.isFalse(Raven.isSetup());
            });

            it('should log an error message, the first time it is called', function () {
                Raven._hasJSON = true;
                Raven._globalServer = undefined;

                Raven.isSetup();
                Raven.isSetup();
                assert.isTrue(Raven._logDebug.calledWith('error', 'Error: Raven has not been configured.'));
                assert.isTrue(Raven._logDebug.calledOnce);
            });
        });

        it('should return true when everything is all gravy', function() {
            Raven._hasJSON = true;

            assert.isTrue(Raven.isSetup());
        });
    });

    describe('logDebug', function() {
        var level = 'error',
            message = 'foobar',
            originalConsoleMethods;

        beforeEach(function () {
            originalConsoleMethods = Raven._originalConsoleMethods;
        });

        it('should not write to console when Raven.debug is false', function() {
            Raven.debug = false;
            this.sinon.stub(originalConsoleMethods, level);
            Raven._logDebug(level, message);
            assert.isFalse(originalConsoleMethods[level].called);
        });

        it('should write to console when Raven.debug is true', function() {
            Raven.debug = true;
            this.sinon.stub(originalConsoleMethods, level);
            Raven._logDebug(level, message);
            assert.isTrue(originalConsoleMethods[level].calledOnce);
        });

        it('should handle variadic arguments', function() {
            Raven.debug = true;
            this.sinon.stub(originalConsoleMethods, level);
            Raven._logDebug(level, message, {}, 'foo');
        });

        it('should be unaffected by monkeypatches to the console built-in', function() {
            Raven.debug = true;
            this.sinon.stub(console, level).throws("can't touch this");
            this.sinon.stub(originalConsoleMethods, level);
            Raven._logDebug(level, message);
            assert.isTrue(originalConsoleMethods[level].calledOnce);
            assert.isFalse(console[level].called);
            console[level].restore();
        });
    });

    describe('parseDSN', function() {
        it('should do what it advertises', function() {
            var pieces = Raven._parseDSN('http://abc@example.com:80/2');
            assert.strictEqual(pieces.protocol, 'http');
            assert.strictEqual(pieces.user, 'abc');
            assert.strictEqual(pieces.port, '80');
            assert.strictEqual(pieces.path, '/2');
            assert.strictEqual(pieces.host, 'example.com');
        });

        it('should parse protocol relative', function() {
            var pieces = Raven._parseDSN('//user@mattrobenolt.com/');
            assert.strictEqual(pieces.protocol, '');
            assert.strictEqual(pieces.user, 'user');
            assert.strictEqual(pieces.port, '');
            assert.strictEqual(pieces.path, '/');
            assert.strictEqual(pieces.host, 'mattrobenolt.com');
        });

        it('should parse domain with hyphen', function() {
            var pieces = Raven._parseDSN('http://user@matt-robenolt.com/1');
            assert.strictEqual(pieces.protocol, 'http');
            assert.strictEqual(pieces.user, 'user');
            assert.strictEqual(pieces.port, '');
            assert.strictEqual(pieces.path, '/1');
            assert.strictEqual(pieces.host, 'matt-robenolt.com');
        });

        it('should parse domain without user', function() {
            var pieces = Raven._parseDSN('http://matt-robenolt.com/1');
            assert.strictEqual(pieces.protocol, 'http');
            assert.strictEqual(pieces.user, '');
            assert.strictEqual(pieces.port, '');
            assert.strictEqual(pieces.path, '/1');
            assert.strictEqual(pieces.host, 'matt-robenolt.com');
        });

        it('should raise a RavenConfigError when setting a password', function() {
            try {
                Raven._parseDSN('http://user:pass@example.com/2');
            } catch(e) {
                return assert.equal(e.name, 'RavenConfigError');
            }
            // shouldn't hit this
            assert.isTrue(false);
        });

        it('should raise a RavenConfigError with an invalid DSN', function() {
            try {
                Raven._parseDSN('lol');
            } catch(e) {
                return assert.equal(e.name, 'RavenConfigError');
            }
            // shouldn't hit this
            assert.isTrue(false);
        });
    });

    describe('normalizeFrame', function() {
        it('should handle a normal frame', function() {
            var frame = {
                url: 'http://example.com/path/file.js',
                line: 10,
                column: 11,
                func: 'lol'
                // context: []    context is stubbed
            };

            assert.deepEqual(Raven._normalizeFrame(frame), {
                filename: 'http://example.com/path/file.js',
                lineno: 10,
                colno: 11,
                'function': 'lol',
                in_app: true
            });
        });

        it('should handle a frame without context', function() {
            var frame = {
                url: 'http://example.com/path/file.js',
                line: 10,
                column: 11,
                func: 'lol'
            };

            assert.deepEqual(Raven._normalizeFrame(frame), {
                filename: 'http://example.com/path/file.js',
                lineno: 10,
                colno: 11,
                'function': 'lol',
                in_app: true
            });
        });

        it('should not mark `in_app` if rules match', function() {
            var frame = {
                url: 'http://example.com/path/file.js',
                line: 10,
                column: 11,
                func: 'lol'
            };

            Raven._globalOptions.includePaths = /^http:\/\/example\.com/;

            assert.deepEqual(Raven._normalizeFrame(frame), {
                filename: 'http://example.com/path/file.js',
                lineno: 10,
                colno: 11,
                'function': 'lol',
                in_app: true
            });
        });

        it('should mark `in_app` if rules do not match', function() {
            var frame = {
                url: 'http://lol.com/path/file.js',
                line: 10,
                column: 11,
                func: 'lol'
            };

            Raven._globalOptions.includePaths = /^http:\/\/example\.com/;

            assert.deepEqual(Raven._normalizeFrame(frame), {
                filename: 'http://lol.com/path/file.js',
                lineno: 10,
                colno: 11,
                'function': 'lol',
                in_app: false
            });
        });

        it('should mark `in_app` for raven.js', function() {
            var frame = {
                url: 'http://lol.com/path/raven.js',
                line: 10,
                column: 11,
                func: 'lol'
            };

            assert.deepEqual(Raven._normalizeFrame(frame), {
                filename: 'http://lol.com/path/raven.js',
                lineno: 10,
                colno: 11,
                'function': 'lol',
                in_app: false
            });
        });

        it('should mark `in_app` for raven.min.js', function() {
            var frame = {
                url: 'http://lol.com/path/raven.min.js',
                line: 10,
                column: 11,
                func: 'lol'
            };

            assert.deepEqual(Raven._normalizeFrame(frame), {
                filename: 'http://lol.com/path/raven.min.js',
                lineno: 10,
                colno: 11,
                'function': 'lol',
                in_app: false
            });
        });

        it('should mark `in_app` for Raven', function() {
            var frame = {
                url: 'http://lol.com/path/file.js',
                line: 10,
                column: 11,
                func: 'Raven.wrap'
            };

            assert.deepEqual(Raven._normalizeFrame(frame), {
                filename: 'http://lol.com/path/file.js',
                lineno: 10,
                colno: 11,
                'function': 'Raven.wrap',
                in_app: false
            });
        });

        it('should mark `in_app` for TraceKit', function() {
            var frame = {
                url: 'http://lol.com/path/file.js',
                line: 10,
                column: 11,
                func: 'TraceKit.lol'
            };

            assert.deepEqual(Raven._normalizeFrame(frame), {
                filename: 'http://lol.com/path/file.js',
                lineno: 10,
                colno: 11,
                'function': 'TraceKit.lol',
                in_app: false
            });
        });

        it('should not blow up if includePaths is empty, regression for #377', function() {
            var frame = {
                url: 'http://lol.com/path/file.js',
                line: 10,
                column: 11,
                func: 'TraceKit.lol'
            };
            Raven._globalOptions.includePaths = [];
            Raven._normalizeFrame(frame);
        });
    });

    describe('processException', function() {
        it('should respect `ignoreErrors`', function() {
            this.sinon.stub(Raven, '_send');

            Raven._globalOptions.ignoreErrors = joinRegExp(['e1', 'e2']);
            Raven._processException('Error', 'e1', 'http://example.com', []);
            assert.isFalse(Raven._send.called);
            Raven._processException('Error', 'e2', 'http://example.com', []);
            assert.isFalse(Raven._send.called);
            Raven._processException('Error', 'error', 'http://example.com', []);
            assert.isTrue(Raven._send.calledOnce);
        });

        it('should handle empty `ignoreErrors`', function() {
            this.sinon.stub(Raven, '_send');

            Raven._globalOptions.ignoreErrors = [];
            Raven._processException('Error', 'e1', 'http://example.com', []);
            assert.isTrue(Raven._send.calledOnce);
        });

        it('should respect `ignoreUrls`', function() {
            this.sinon.stub(Raven, '_send');

            Raven._globalOptions.ignoreUrls = joinRegExp([/.+?host1.+/, /.+?host2.+/]);
            Raven._processException('Error', 'error', 'http://host1/', []);
            assert.isFalse(Raven._send.called);
            Raven._processException('Error', 'error', 'http://host2/', []);
            assert.isFalse(Raven._send.called);
            Raven._processException('Error', 'error', 'http://host3/', []);
            assert.isTrue(Raven._send.calledOnce);
        });

        it('should handle empty `ignoreUrls`', function() {
            this.sinon.stub(Raven, '_send');

            Raven._globalOptions.ignoreUrls = [];
            Raven._processException('Error', 'e1', 'http://example.com', []);
            assert.isTrue(Raven._send.calledOnce);
        });

        it('should respect `whitelistUrls`', function() {
            this.sinon.stub(Raven, '_send');

            Raven._globalOptions.whitelistUrls = joinRegExp([/.+?host1.+/, /.+?host2.+/]);
            Raven._processException('Error', 'error', 'http://host1/', []);
            assert.equal(Raven._send.callCount, 1);
            Raven._processException('Error', 'error', 'http://host2/', []);
            assert.equal(Raven._send.callCount, 2);
            Raven._processException('Error', 'error', 'http://host3/', []);
            assert.equal(Raven._send.callCount, 2);
        });

        it('should handle empty `whitelistUrls`', function() {
            this.sinon.stub(Raven, '_send');

            Raven._globalOptions.whitelistUrls = [];
            Raven._processException('Error', 'e1', 'http://example.com', []);
            assert.isTrue(Raven._send.calledOnce);
        });

        it('should send a proper payload with frames', function() {
            this.sinon.stub(Raven, '_send');

            var frames = [
                {
                    filename: 'http://example.com/file1.js'
                },
                {
                    filename: 'http://example.com/file2.js'
                }
            ], framesFlipped = frames.slice(0);

            framesFlipped.reverse();

            Raven._processException('Error', 'lol', 'http://example.com/override.js', 10, frames.slice(0), {});
            assert.deepEqual(Raven._send.lastCall.args, [{
                exception: {
                    values: [{
                        type: 'Error',
                        value: 'lol',
                        stacktrace: {
                            frames: framesFlipped
                        }
                    }]
                },
                culprit: 'http://example.com/file1.js'
            }]);

            Raven._processException('Error', 'lol', '', 10, frames.slice(0), {});
            assert.deepEqual(Raven._send.lastCall.args, [{
                exception: {
                    values: [{
                        type: 'Error',
                        value: 'lol',
                        stacktrace: {
                            frames: framesFlipped
                        }
                    }]
                },
                culprit: 'http://example.com/file1.js'
            }]);

            Raven._processException('Error', 'lol', '', 10, frames.slice(0), {extra: 'awesome'});
            assert.deepEqual(Raven._send.lastCall.args, [{
                exception: {
                    values: [{
                        type: 'Error',
                        value: 'lol',
                        stacktrace: {
                            frames: framesFlipped
                        }
                    }]
                },
                culprit: 'http://example.com/file1.js',
                extra: 'awesome'
            }]);
        });

        it('should send a proper payload without frames', function() {
            this.sinon.stub(Raven, '_send');

            Raven._processException('Error', 'lol', 'http://example.com/override.js', 10, [], {});
            assert.deepEqual(Raven._send.lastCall.args, [{
                exception: {
                    values: [{
                        type: 'Error',
                        value: 'lol',
                        stacktrace: {
                            frames: [{
                                filename: 'http://example.com/override.js',
                                lineno: 10,
                                in_app: true
                            }]
                        }
                    }]
                },
                culprit: 'http://example.com/override.js'
            }]);

            Raven._processException('Error', 'lol', 'http://example.com/override.js', 10, [], {});
            assert.deepEqual(Raven._send.lastCall.args, [{
                exception: {
                    values: [{
                        type: 'Error',
                        value: 'lol',
                        stacktrace: {
                            frames: [{
                                filename: 'http://example.com/override.js',
                                lineno: 10,
                                in_app: true
                            }]
                        }
                    }]
                },
                culprit: 'http://example.com/override.js'
            }]);

            Raven._processException('Error', 'lol', 'http://example.com/override.js', 10, [], {extra: 'awesome'});
            assert.deepEqual(Raven._send.lastCall.args, [{
                exception: {
                    values: [{
                        type: 'Error',
                        value: 'lol',
                        stacktrace: {
                            frames: [{
                                filename: 'http://example.com/override.js',
                                lineno: 10,
                                in_app: true
                            }]
                        }
                    }]
                },
                culprit: 'http://example.com/override.js',
                extra: 'awesome'
            }]);
        });

        it('should not blow up with `undefined` message', function() {
            this.sinon.stub(Raven, '_send');

            Raven._processException('TypeError', undefined, 'http://example.com', []);
            assert.isTrue(Raven._send.called);
        });
    });

    describe('send', function() {
        it('should build a good data payload', function() {
            this.sinon.stub(Raven, 'isSetup').returns(true);
            this.sinon.stub(Raven, '_makeRequest');
            this.sinon.stub(Raven, '_getHttpData').returns({
                url: 'http://localhost/?a=b',
                headers: {'User-Agent': 'lolbrowser'}
            });

            Raven._globalProject = '2';
            Raven._globalOptions = {
                logger: 'javascript',
                maxMessageLength: 100
            };
            Raven._breadcrumbs = [{type: 'request', timestamp: 0.1, data: {method: 'POST', url: 'http://example.org/api/0/auth/'}}];

            Raven._send({message: 'bar'});
            assert.deepEqual(Raven._makeRequest.lastCall.args[0].data, {
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
                message: 'bar',
                extra: {'session:duration': 100},
                breadcrumbs: {
                    values: [
                        { type: 'request', timestamp: 0.1, data: { method: 'POST', url: 'http://example.org/api/0/auth/' }}
                    ]
                }
            });
        });

        it('should create and append \'sentry\' breadcrumb', function () {
            this.sinon.stub(Raven, 'isSetup').returns(true);
            this.sinon.stub(Raven, '_makeRequest');
            this.sinon.stub(Raven, '_getHttpData').returns({
                url: 'http://localhost/?a=b',
                headers: {'User-Agent': 'lolbrowser'}
            });

            Raven._globalProject = '2';
            Raven._globalOptions = {
                logger: 'javascript',
                maxMessageLength: 100
            };
            Raven._breadcrumbs = [{type: 'http', timestamp: 0.1, data: {method: 'POST', url: 'http://example.org/api/0/auth/'}}];

            Raven._send({message: 'bar'});

            assert.deepEqual(Raven._breadcrumbs, [
                { type: 'http', timestamp: 0.1, data: { method: 'POST', url: 'http://example.org/api/0/auth/' }},
                { category: 'sentry', message: 'bar', timestamp: 0.1, /* 100ms */ event_id: 'abc123', level: 'error' }
            ]);

            Raven._send({message: 'foo', level: 'warning' });
            assert.deepEqual(Raven._breadcrumbs, [
                { type: 'http', timestamp: 0.1, data: { method: 'POST', url: 'http://example.org/api/0/auth/' }},
                { category: 'sentry', message: 'bar', timestamp: 0.1, /* 100ms */ event_id: 'abc123', level: 'error' },
                { category: 'sentry', message: 'foo', timestamp: 0.1, /* 100ms */ event_id: 'abc123', level: 'warning' }
            ]);

            Raven._send({
                exception: {
                    values: [{
                        type: 'ReferenceError',
                        value: 'foo is not defined'
                    }]
                }
            });
            assert.deepEqual(Raven._breadcrumbs, [
                { type: 'http', timestamp: 0.1, data: { method: 'POST', url: 'http://example.org/api/0/auth/' }},
                { category: 'sentry', message: 'bar', timestamp: 0.1, /* 100ms */ event_id: 'abc123', level: 'error' },
                { category: 'sentry', message: 'foo', timestamp: 0.1, /* 100ms */ event_id: 'abc123', level: 'warning' },
                { category: 'sentry', message: 'ReferenceError: foo is not defined', timestamp: 0.1, /* 100ms */ event_id: 'abc123', level: 'error' }
            ]);

        });

        it('should build a good data payload with a User', function() {
            this.sinon.stub(Raven, 'isSetup').returns(true);
            this.sinon.stub(Raven, '_makeRequest');
            this.sinon.stub(Raven, '_getHttpData').returns({
                url: 'http://localhost/?a=b',
                headers: {'User-Agent': 'lolbrowser'}
            });

            Raven._globalProject = '2';
            Raven._globalOptions = {
                logger: 'javascript',
                maxMessageLength: 100
            };
            Raven._globalContext = {user: {name: 'Matt'}};

            Raven._send({message: 'bar'});
            assert.deepEqual(Raven._makeRequest.lastCall.args[0].data, {
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
                message: 'bar',
                extra: {'session:duration': 100}
            });
        });

        it('should merge in global tags', function() {
            this.sinon.stub(Raven, 'isSetup').returns(true);
            this.sinon.stub(Raven, '_makeRequest');
            this.sinon.stub(Raven, '_getHttpData').returns({
                url: 'http://localhost/?a=b',
                headers: {'User-Agent': 'lolbrowser'}
            });

            Raven._globalProject = '2';
            Raven._globalOptions = {
                logger: 'javascript',
                maxMessageLength: 100
            };
            Raven._globalContext = {tags: {tag1: 'value1'}};

            Raven._send({message: 'bar', tags: {tag2: 'value2'}});
            assert.deepEqual(Raven._makeRequest.lastCall.args[0].data, {
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
                message: 'bar',
                tags: {tag1: 'value1', tag2: 'value2'},
                extra: {'session:duration': 100},
            });


            assert.deepEqual(Raven._globalOptions, {
                logger: 'javascript',
                maxMessageLength: 100
            });
            assert.deepEqual(Raven._globalContext, {
                tags: {tag1: 'value1'}
            });
        });

        it('should merge in global extra', function() {
            this.sinon.stub(Raven, 'isSetup').returns(true);
            this.sinon.stub(Raven, '_makeRequest');
            this.sinon.stub(Raven, '_getHttpData').returns({
                url: 'http://localhost/?a=b',
                headers: {'User-Agent': 'lolbrowser'}
            });

            Raven._globalProject = '2';
            Raven._globalOptions = {
                logger: 'javascript',
                maxMessageLength: 100
            };
            Raven._globalContext = {extra: {key1: 'value1'}};

            Raven._send({message: 'bar', extra: {key2: 'value2'}});
            assert.deepEqual(Raven._makeRequest.lastCall.args[0].data, {
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
                message: 'bar',
                extra: {key1: 'value1', key2: 'value2', 'session:duration': 100},
            });

            assert.deepEqual(Raven._globalOptions, {
                logger: 'javascript',
                maxMessageLength: 100
            });
            assert.deepEqual(Raven._globalContext, {
                extra: {key1: 'value1'}
            });
        });

        it('should let dataCallback override everything', function() {
            this.sinon.stub(Raven, 'isSetup').returns(true);
            this.sinon.stub(Raven, '_makeRequest');

            Raven._globalOptions = {
                projectId: 2,
                logger: 'javascript',
                maxMessageLength: 100,
                dataCallback: function() {
                    return {message: 'ibrokeit'};
                }
            };
            Raven._globalContext = {user: {name: 'Matt'}};

            Raven._send({message: 'bar'});
            assert.deepEqual(Raven._makeRequest.lastCall.args[0].data, {
                message: 'ibrokeit',
                event_id: 'abc123'
            });
        });

        it('should ignore dataCallback if it does not return anything', function() {
            this.sinon.stub(Raven, 'isSetup').returns(true);
            this.sinon.stub(Raven, '_makeRequest');
            this.sinon.stub(Raven, '_getHttpData').returns({
                url: 'http://localhost/?a=b',
                headers: {'User-Agent': 'lolbrowser'}
            });

            Raven._globalProject = '2';
            Raven._globalOptions = {
                logger: 'javascript',
                maxMessageLength: 100,
                dataCallback: function() {
                    return;
                }
            };

            Raven._send({message: 'bar'});
            assert.deepEqual(Raven._makeRequest.lastCall.args[0].data, {
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
                message: 'bar',
                extra: {'session:duration': 100}
            });
        });

        it('should strip empty tags', function() {
            this.sinon.stub(Raven, 'isSetup').returns(true);
            this.sinon.stub(Raven, '_makeRequest');
            this.sinon.stub(Raven, '_getHttpData').returns({
                url: 'http://localhost/?a=b',
                headers: {'User-Agent': 'lolbrowser'}
            });

            Raven._globalOptions = {
                projectId: 2,
                logger: 'javascript',
                maxMessageLength: 100,
                tags: {}
            };

            Raven._send({message: 'bar', tags: {}, extra: {}});
            assert.deepEqual(Raven._makeRequest.lastCall.args[0].data, {
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
                message: 'bar',
                extra: {'session:duration': 100}
            });
        });

        it('should attach environment if available', function() {
            this.sinon.stub(Raven, 'isSetup').returns(true);
            this.sinon.stub(Raven, '_makeRequest');
            this.sinon.stub(Raven, '_getHttpData').returns({
                url: 'http://localhost/?a=b',
                headers: {'User-Agent': 'lolbrowser'}
            });

            Raven._globalOptions = {
                projectId: 2,
                logger: 'javascript',
                maxMessageLength: 100,
                environment: 'abc123'
            };

            Raven._send({message: 'bar'});
            assert.deepEqual(Raven._makeRequest.lastCall.args[0].data, {
                project: '2',
                environment: 'abc123',
                logger: 'javascript',
                platform: 'javascript',
                request: {
                    url: 'http://localhost/?a=b',
                    headers: {
                        'User-Agent': 'lolbrowser'
                    }
                },
                event_id: 'abc123',
                message: 'bar',
                extra: {'session:duration': 100}
            });
        });

        it('should attach release if available', function() {
            this.sinon.stub(Raven, 'isSetup').returns(true);
            this.sinon.stub(Raven, '_makeRequest');
            this.sinon.stub(Raven, '_getHttpData').returns({
                url: 'http://localhost/?a=b',
                headers: {'User-Agent': 'lolbrowser'}
            });

            Raven._globalOptions = {
                projectId: 2,
                logger: 'javascript',
                maxMessageLength: 100,
                release: 'abc123'
            };

            Raven._send({message: 'bar'});
            assert.deepEqual(Raven._makeRequest.lastCall.args[0].data, {
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
                message: 'bar',
                extra: {'session:duration': 100}
            });
        });

        it('should attach server_name if available', function() {
            this.sinon.stub(Raven, 'isSetup').returns(true);
            this.sinon.stub(Raven, '_makeRequest');
            this.sinon.stub(Raven, '_getHttpData').returns({
                url: 'http://localhost/?a=b',
                headers: {'User-Agent': 'lolbrowser'}
            });

            Raven._globalOptions = {
                projectId: 2,
                logger: 'javascript',
                maxMessageLength: 100,
                serverName: 'abc123',
            };

            Raven._send({message: 'bar'});
            assert.deepEqual(Raven._makeRequest.lastCall.args[0].data, {
                project: '2',
                server_name: 'abc123',
                logger: 'javascript',
                platform: 'javascript',
                request: {
                    url: 'http://localhost/?a=b',
                    headers: {
                        'User-Agent': 'lolbrowser'
                    }
                },
                event_id: 'abc123',
                message: 'bar',
                extra: {'session:duration': 100}
            });
        });

        it('should pass correct opts to makeRequest', function() {
            this.sinon.stub(Raven, 'isSetup').returns(true);
            this.sinon.stub(Raven, '_makeRequest');
            this.sinon.stub(Raven, '_getHttpData').returns({
                url: 'http://localhost/?a=b',
                headers: {'User-Agent': 'lolbrowser'}
            });

            var globalOptions = {
                projectId: 2,
                logger: 'javascript',
                maxMessageLength: 100,
                release: 'abc123',
            };
            Raven._globalEndpoint = 'http://localhost/store/';
            Raven._globalOptions = globalOptions;

            Raven._send({message: 'bar'});
            var args = Raven._makeRequest.lastCall.args;
            assert.equal(args.length, 1);
            var opts = args[0];
            assert.equal(opts.url, 'http://localhost/store/');
            assert.deepEqual(opts.data, {
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
                message: 'bar',
                extra: {'session:duration': 100},
            });
            assert.deepEqual(opts.auth, {
                sentry_client: 'raven-js/3.8.1',
                sentry_key: 'abc',
                sentry_version: '7'
            });
            assert.deepEqual(opts.options, globalOptions);
            assert.isFunction(opts.onSuccess);
            assert.isFunction(opts.onError);
        });

        it('should pass sentry_secret as part of auth params if specified', function () {
            this.sinon.stub(Raven, 'isSetup').returns(true);
            this.sinon.stub(Raven, '_makeRequest');
            this.sinon.stub(Raven, '_getHttpData').returns({
                url: 'http://localhost/?a=b',
                headers: {'User-Agent': 'lolbrowser'}
            });

            Raven._globalEndpoint = 'http://localhost/store/';
            Raven._globalOptions = {
                projectId: 2,
                logger: 'javascript',
                maxMessageLength: 100,
                release: 'abc123'
            };;
            Raven._globalSecret = 'def'; // <-- secret

            Raven._send({message: 'bar'});
            var args = Raven._makeRequest.lastCall.args;
            assert.equal(args.length, 1);
            var opts = args[0];
            assert.equal(opts.url, 'http://localhost/store/');
            assert.deepEqual(opts.data, {
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
                message: 'bar',
                extra: {'session:duration': 100},
            });
            assert.deepEqual(opts.auth, {
                sentry_client: 'raven-js/3.8.1',
                sentry_key: 'abc',
                sentry_secret: 'def',
                sentry_version: '7'
            });
        });

        it('should call globalOptions.transport if specified', function() {
            this.sinon.stub(Raven, 'isSetup').returns(true);
            this.sinon.stub(Raven, '_getHttpData').returns({
                url: 'http://localhost/?a=b',
                headers: {'User-Agent': 'lolbrowser'}
            });

            var globalOptions = {
                logger: 'javascript',
                maxMessageLength: 100,
                transport: sinon.stub()
            };

            Raven._globalProject = '2';
            Raven._globalOptions = globalOptions;

            Raven._send({message: 'bar'});
            assert.deepEqual(globalOptions.transport.lastCall.args[0].data, {
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
                message: 'bar',
                extra: {'session:duration': 100}
            });
        });

        it('should check `Raven.isSetup`', function() {
            this.sinon.stub(Raven, 'isSetup').returns(false);
            this.sinon.stub(Raven, '_makeRequest');
            Raven._send({message: 'bar'});
            assert.isTrue(Raven.isSetup.called);
        });

        it('should not makeRequest if `Raven.isSetup` is false', function() {
            this.sinon.stub(Raven, 'isSetup').returns(false);
            this.sinon.stub(Raven, '_makeRequest');
            Raven._send({message: 'bar'});
            assert.isFalse(Raven._makeRequest.called);
        });

        it('should log to console', function() {
            this.sinon.stub(Raven, 'isSetup').returns(true);
            this.sinon.stub(Raven, '_logDebug');
            this.sinon.stub(Raven, '_makeRequest');
            Raven._send({message: 'bar'});
            assert.isTrue(Raven._logDebug.called);
        });

        it('should truncate messages to the specified length', function() {
            this.sinon.stub(Raven, 'isSetup').returns(true);
            this.sinon.stub(Raven, '_makeRequest');

            Raven._globalOptions.maxMessageLength = 150;

            var message = new Array(500).join('a');
            var shortMessage = new Array(151).join('a')+'\u2026';

            Raven._send({
                message: message,
                exception: {
                    values: [{
                        value: message
                    }]
                }
            });

            var args = Raven._makeRequest.lastCall.args;
            assert.equal(args.length, 1);
            var data = args[0].data;
            assert.equal(data.message, shortMessage);
            assert.equal(data.exception.values[0].value, shortMessage);
        });
    });

    describe('makeRequest', function() {
        beforeEach(function() {
            // NOTE: can't seem to call useFakeXMLHttpRequest via sandbox; must
            //       restore manually
            this.xhr = sinon.useFakeXMLHttpRequest();
            var requests = this.requests = [];

            this.xhr.onCreate = function (xhr) {
                requests.push(xhr);
            };
        });

        afterEach(function() {
            this.xhr.restore();
        });

        it('should create an XMLHttpRequest object with body as JSON payload', function() {
            XMLHttpRequest.prototype.withCredentials = true;

            Raven._makeRequest({
                url: 'http://localhost/',
                auth: {a: '1', b: '2'},
                data: {foo: 'bar'},
                options: Raven._globalOptions
            });

            var lastXhr = this.requests[this.requests.length - 1];
            assert.equal(lastXhr.requestBody, '{"foo":"bar"}');
            assert.equal(lastXhr.url, 'http://localhost/?a=1&b=2');
        });

        it('should no-op if CORS is not supported', function () {
            delete XMLHttpRequest.prototype.withCredentials;

            var oldXDR = window.XDomainRequest;
            window.XDomainRequest = undefined;

            Raven._makeRequest({
                url: 'http://localhost/',
                auth: {a: '1', b: '2'},
                data: {foo: 'bar'},
                options: Raven._globalOptions
            });

            assert.equal(this.requests.length, 1); // the "test" xhr
            assert.equal(this.requests[0].readyState, 0);

            window.XDomainRequest = oldXDR
        });
    });

    describe('handleOnErrorStackInfo', function () {
        it('should call handleStackInfo if ignoreOnError is falsy', function () {
            var frame = {url: 'http://example.com'};
            this.sinon.stub(Raven, '_handleStackInfo');

            var stackInfo = {
                name: 'Matt',
                message: 'hey',
                url: 'http://example.com',
                lineno: 10,
                stack: [
                    frame, frame
                ]
            };

            Raven._ignoreOnError = 1;
            Raven._handleOnErrorStackInfo(stackInfo, {foo: 'bar'});

            assert.equal(Raven._handleStackInfo.callCount, 0);

            Raven._ignoreOnError = 0;
            Raven._handleOnErrorStackInfo(stackInfo, {foo: 'bar'});

            assert.equal(Raven._handleStackInfo.callCount, 1);
            assert.deepEqual(Raven._handleStackInfo.lastCall.args, [
                stackInfo, {foo: 'bar'}
            ]);
        });
    });

    describe('handleStackInfo', function() {
        it('should work as advertised', function() {
            var frame = {url: 'http://example.com'};
            this.sinon.stub(Raven, '_normalizeFrame').returns(frame);
            this.sinon.stub(Raven, '_processException');

            var stackInfo = {
                name: 'Matt',
                message: 'hey',
                url: 'http://example.com',
                lineno: 10,
                stack: [
                    frame, frame
                ]
            };

            Raven._handleStackInfo(stackInfo, {foo: 'bar'});
            assert.deepEqual(Raven._processException.lastCall.args, [
                'Matt', 'hey', 'http://example.com', 10, [frame, frame], {foo: 'bar'}
            ]);
        });

        it('should work as advertised #integration', function() {
            this.sinon.stub(Raven, '_makeRequest');
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

            Raven._handleStackInfo(stackInfo, {foo: 'bar'});
            assert.isTrue(Raven._makeRequest.calledOnce);
            /* This is commented out because chai is broken.

            assert.deepEqual(Raven._makeRequest.lastCall.args, [{
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
                message: 'Error: crap',
                foo: 'bar'
            }]);
            */
        });

        it('should ignore frames that dont have a url', function() {
            this.sinon.stub(Raven, '_normalizeFrame').returns(undefined);
            this.sinon.stub(Raven, '_processException');

            var stackInfo = {
                name: 'Matt',
                message: 'hey',
                url: 'http://example.com',
                lineno: 10,
                stack: new Array(2)
            };

            Raven._handleStackInfo(stackInfo, {foo: 'bar'});
            assert.deepEqual(Raven._processException.lastCall.args, [
                'Matt', 'hey', 'http://example.com', 10, [], {foo: 'bar'}
            ]);
        });

        it('should not shit when there is no stack object from TK', function() {
            this.sinon.stub(Raven, '_normalizeFrame').returns(undefined);
            this.sinon.stub(Raven, '_processException');

            var stackInfo = {
                name: 'Matt',
                message: 'hey',
                url: 'http://example.com',
                lineno: 10
                // stack: new Array(2)
            };

            Raven._handleStackInfo(stackInfo);
            assert.isFalse(Raven._normalizeFrame.called);
            assert.deepEqual(Raven._processException.lastCall.args, [
                'Matt', 'hey', 'http://example.com', 10, [], undefined
            ]);
        });

        it('should detect 2-words patterns (angularjs frequent case)', function() {
            this.sinon.stub(Raven, '_normalizeFrame').returns(undefined);
            this.sinon.stub(Raven, '_processException');

            var stackInfo = {
                name: 'new <anonymous>',
                message: 'hey',
                url: 'http://example.com',
                lineno: 10
                // stack: new Array(2)
            };

            Raven._handleStackInfo(stackInfo);
            assert.isFalse(Raven._normalizeFrame.called);
            assert.deepEqual(Raven._processException.lastCall.args, [
                'new <anonymous>', 'hey', 'http://example.com', 10, [], undefined
            ]);
        });

        it('should trim number of frames based on stackTraceLimit', function() {
            var frame = {url: 'http://example.com'};
            this.sinon.stub(Raven, '_normalizeFrame').returns(frame);
            this.sinon.stub(Raven, '_processException');

            var stackInfo = {
                name: 'Matt',
                message: 'hey',
                url: 'http://example.com',
                lineno: 10,
                stack: [
                  frame, frame
                ]
            };

            Raven._globalOptions.stackTraceLimit = 1;

            Raven._handleStackInfo(stackInfo);
            assert.deepEqual(Raven._processException.lastCall.args, [
                'Matt', 'hey', 'http://example.com', 10, [frame], undefined
            ]);
        });
    });
});

describe('Raven (public API)', function() {

    beforeEach(function () {
        this.clock = sinon.useFakeTimers();
        this.clock.tick(0); // Raven initialized at time "0"
        Raven = new _Raven();

        this.clock.tick(100); // tick 100 ms
    });

    afterEach(function () {
        this.clock.restore();
    });

    describe('.VERSION', function() {
        it('should have a version', function() {
            assert.isString(Raven.VERSION);
        });
    });

    describe('ignore errors', function() {
        it('should install default ignore errors', function() {
            Raven.config('//abc@example.com/2');

            assert.isTrue(Raven._globalOptions.ignoreErrors.test('Script error'), 'it should install "Script error" by default');
            assert.isTrue(Raven._globalOptions.ignoreErrors.test('Script error.'), 'it should install "Script error." by default');
            assert.isTrue(Raven._globalOptions.ignoreErrors.test('Javascript error: Script error on line 0'), 'it should install "Javascript error: Script error on line 0" by default');
            assert.isTrue(Raven._globalOptions.ignoreErrors.test('Javascript error: Script error. on line 0'), 'it should install "Javascript error: Script error. on line 0" by default');
        });
    });

    describe('callback function', function() {
        it('should callback a function if it is global', function() {
            window.RavenConfig = {
                dsn: "http://random@some.other.server:80/2",
                config: {some: 'config'}
            };

            this.sinon.stub(Raven, 'isSetup').returns(false);
            this.sinon.stub(TraceKit.report, 'subscribe');

            Raven.afterLoad();

            assert.equal(Raven._globalKey, 'random');
            assert.equal(Raven._globalEndpoint, 'http://some.other.server:80/api/2/store/');

            assert.equal(Raven._globalOptions.some, 'config');
            assert.equal(Raven._globalProject, '2');

            assert.isTrue(Raven.isSetup.calledOnce);
            assert.isFalse(TraceKit.report.subscribe.calledOnce);

            delete window.RavenConfig;
        });
    });

    describe('.setDSN', function() {
        it('should work with a DSN after Raven has been configured', function() {
            Raven.config('//def@lol.com/3');
            Raven.setDSN(SENTRY_DSN)

            assert.equal(Raven._globalKey, 'abc');
            assert.equal(Raven._globalSecret, '');
            assert.equal(Raven._globalEndpoint, 'http://example.com:80/api/2/store/');
            assert.equal(Raven._globalProject, '2');
        });
    });

    describe('.config', function() {
        it('should work with a DSN', function() {
            assert.equal(Raven, Raven.config(SENTRY_DSN, {foo: 'bar'}), 'it should return Raven');

            assert.equal(Raven._globalKey, 'abc');
            assert.equal(Raven._globalSecret, '');
            assert.equal(Raven._globalEndpoint, 'http://example.com:80/api/2/store/');
            assert.equal(Raven._globalOptions.foo, 'bar');
            assert.equal(Raven._globalProject, '2');
            assert.isTrue(Raven.isSetup());
        });

        it('throw an Error if the DSN contains a private/secret key', function () {
            assert.throws(function () {
                Raven.config('http://abc:def@example.com:80/2');
            }, Error);
        });

        it('will NOT throw an Error if the DSN contains a private/secret key AND allowSecretKey is true', function () {
            assert.equal(
                Raven,
                Raven.config('http://abc:def@example.com:80/2', {allowSecretKey: true}),
                'it should return Raven'
            );

            assert.equal(Raven._globalKey, 'abc');
            assert.equal(Raven._globalSecret, 'def');
            assert.equal(Raven._globalEndpoint, 'http://example.com:80/api/2/store/');
            assert.equal(Raven._globalProject, '2');
            assert.isTrue(Raven.isSetup());
        });

        it('should work with a protocol relative DSN', function() {
            Raven.config('//abc@example.com/2');

            assert.equal(Raven._globalKey, 'abc');
            assert.equal(Raven._globalEndpoint, '//example.com/api/2/store/');
            assert.equal(Raven._globalProject, '2');
            assert.isTrue(Raven.isSetup());
        });

        it('should work should work at a non root path', function() {
            Raven.config('//abc@example.com/sentry/2');
            assert.equal(Raven._globalKey, 'abc');
            assert.equal(Raven._globalEndpoint, '//example.com/sentry/api/2/store/');
            assert.equal(Raven._globalProject, '2');
            assert.isTrue(Raven.isSetup());
        });

        it('should noop a falsey dsn', function() {
            Raven.config('');
            assert.isFalse(Raven.isSetup());
        });

        it('should return Raven for a falsey dsn', function() {
            assert.equal(Raven.config(''), Raven);
        });

        it('should not set global options more than once', function() {
            this.sinon.spy(Raven, '_parseDSN');
            this.sinon.stub(Raven, '_logDebug');
            setupRaven();
            setupRaven();
            assert.isTrue(Raven._parseDSN.calledOnce);
            assert.isTrue(Raven._logDebug.called);
        });

        describe('whitelistUrls', function() {
            it('should be false if none are passed', function() {
                Raven.config('//abc@example.com/2');
                assert.equal(Raven._globalOptions.whitelistUrls, false);
            });

            it('should join into a single RegExp', function() {
                Raven.config('//abc@example.com/2', {
                    whitelistUrls: [
                        /my.app/i,
                        /other.app/i
                    ]
                });

                assert.match(Raven._globalOptions.whitelistUrls, /my.app|other.app/i);
            });

            it('should handle strings as well', function() {
                Raven.config('//abc@example.com/2', {
                    whitelistUrls: [
                        /my.app/i,
                        "stringy.app"
                    ]
                });

                assert.match(Raven._globalOptions.whitelistUrls, /my.app|stringy.app/i);
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

        describe('maxBreadcrumbs', function () {
            it('should override the default', function () {
                Raven.config(SENTRY_DSN, { maxBreadcrumbs: 50 });
                assert.equal(Raven._globalOptions.maxBreadcrumbs, 50);
            });

            it('should not permit maxBreadcrumbs above 100', function () {
                Raven.config(SENTRY_DSN, { maxBreadcrumbs: 200 });
                assert.equal(Raven._globalOptions.maxBreadcrumbs, 100);
            });

            it('should not permit maxBreadcrumbs below 0', function () {
               Raven.config(SENTRY_DSN, { maxBreadcrumbs: -1 });
                assert.equal(Raven._globalOptions.maxBreadcrumbs, 0);
            });

            it('should set maxBreadcrumbs to the default if not provided', function () {
                Raven.config(SENTRY_DSN);
                assert.equal(Raven._globalOptions.maxBreadcrumbs, 100);
            });
        });

        describe('autoBreadcrumbs', function () {
            it('should convert `true` to a dictionary of enabled breadcrumb features', function () {
                Raven.config(SENTRY_DSN);
                assert.deepEqual(Raven._globalOptions.autoBreadcrumbs, {
                    xhr: true,
                    console: true,
                    dom: true,
                    location: true
                });
            });

            it('should leave false as-is', function () {
                Raven.config(SENTRY_DSN, {
                    autoBreadcrumbs: false
                });
                assert.equal(Raven._globalOptions.autoBreadcrumbs, false);
            });

            it('should merge objects with the default autoBreadcrumb settings', function () {
                Raven.config(SENTRY_DSN, {
                    autoBreadcrumbs: {
                        location: false
                    }
                });
                assert.deepEqual(Raven._globalOptions.autoBreadcrumbs, {
                    xhr: true,
                    console: true,
                    dom: true,
                    location: false /* ! */
                });
            });
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

        it('should copy prototype property when wrapping function', function() {
            var func = function() {};
            func.prototype.test = 'foo';
            var wrapped = Raven.wrap(func);
            assert.equal(new wrapped().test, 'foo');
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
            this.sinon.stub(Raven, 'captureException');
            assert.throws(function() {
                Raven.wrap(function() { throw error; })();
            }, error);
        });

        it('should return input funciton as-is if accessing __raven__ prop throws exception', function (){
            // see raven-js#495
            var fn = function () {};
            Object.defineProperty(fn, '__raven__', {
              get: function () {
                  throw new Error('Permission denied')
              }
            });
            assert.throw(function () { fn.__raven__; }, 'Permission denied');
            var wrapped = Raven.wrap(fn);
            assert.equal(fn, wrapped);
        });

        it('should return input funciton as-is if accessing __raven_wrapper__ prop throws exception', function (){
            // see raven-js#495
            var fn = function () {};
            Object.defineProperty(fn, '__raven_wrapper__', {
              get: function () {
                  throw new Error('Permission denied')
              }
            });
            assert.throw(function () { fn.__raven_wrapper__; }, 'Permission denied');
            var wrapped = Raven.wrap(fn);
            assert.equal(fn, wrapped);
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
            assert.deepEqual(Raven.captureException.lastCall.args, [error, {
                'foo': 'bar'
            }]);
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

    describe('.setUserContext', function() {
        it('should set the globalContext.user object', function() {
            Raven.setUserContext({name: 'Matt'});
            assert.deepEqual(Raven._globalContext.user, {name: 'Matt'});
        });

        it('should not merge globalContext.user object, but rewrite', function () {
            Raven._globalContext.user = {name: 'Matt'};
            Raven.setUserContext({age: 34, email: 'john@mail.com'});
            assert.deepEqual(Raven._globalContext.user, {age: 34, email: 'john@mail.com'});
        });

        it('should clear the globalContext.user with no arguments', function() {
            Raven._globalContext.user = {name: 'Matt'};
            Raven.setUserContext();
            assert.isUndefined(Raven._globalContext.user);
        });
    });

    describe('.setExtraContext', function() {
        it('should set the globalContext.extra object', function() {
            Raven.setExtraContext({name: 'Matt'});
            assert.deepEqual(Raven._globalContext.extra, {name: 'Matt'});
        });

        it('should clear globalContext.extra with no arguments', function() {
            Raven.extra = {name: 'Matt'};
            Raven.setExtraContext();
            assert.isUndefined(Raven._globalContext.extra);
        });

        it('should merge globalContext.extra with subsequent calls', function() {
            Raven.setExtraContext({a: 1});
            Raven.setExtraContext({b: 2});
            assert.deepEqual(Raven._globalContext.extra, {a: 1, b: 2});
        });
    });

    describe('.setTagsContext', function() {
        it('should set the globalContext.tags object', function() {
            Raven.setTagsContext({name: 'Matt'});
            assert.deepEqual(Raven._globalContext.tags, {name: 'Matt'});
        });

        it('should clear globalContext.tags with no arguments', function() {
            Raven._globalContext.tags = {name: 'Matt'};
            Raven.setTagsContext();
            assert.isUndefined(Raven._globalContext.tags);
        });

        it('should merge globalContext.tags with subsequent calls', function() {
            Raven.setTagsContext({a: 1});
            Raven.setTagsContext({b: 2});
            assert.deepEqual(Raven._globalContext.tags, {a: 1, b: 2});
        });
    });

    describe('.clearContext', function() {
        it('should clear the globalContext object', function() {
            Raven._globalState = {globalContext: {tags: {}, extra: {}, user: {}}};
            Raven.clearContext();
            assert.deepEqual(Raven._globalContext, {});
        });
    });

    describe('.getContext', function() {
        it('should retrieve a copy of the current context', function() {
            Raven._globalContext = {tags: {a: 1}};
            var context = Raven.getContext();
            var globalContext = Raven._globalContext;
            assert.deepEqual(globalContext, context);
            context.tags.a = 2;
            // It shouldn't have mutated the original
            assert.equal(globalContext.tags.a, 1);
        });
    });

    describe('.setEnvironment', function() {
        it('should set the globalOptions.environment attribute', function() {
            Raven.setEnvironment('abc123');
            assert.equal(Raven._globalOptions.environment, 'abc123');
        });

        it('should clear globalOptions.environment with no arguments', function() {
            Raven._globalOptions.environment = 'abc123';
            Raven.setEnvironment();
            assert.isUndefined(Raven._globalOptions.environment);
        });
    });

    describe('.setRelease', function() {
        it('should set the globalOptions.release attribute', function() {
            Raven.setRelease('abc123');
            assert.equal(Raven._globalOptions.release, 'abc123');
        });

        it('should clear globalOptions.release with no arguments', function() {
            Raven._globalOptions.release = 'abc123';
            Raven.setRelease();
            assert.isUndefined(Raven._globalOptions.release);
        });
    });

    describe('.setDataCallback', function() {
        it('should set the globalOptions.dataCallback attribute', function() {
            var foo = sinon.stub();
            Raven.setDataCallback(foo);

            // note that setDataCallback creates a callback/closure around
            // foo, so can't test for equality - just verify that calling the wrapper
            // also calls foo
            Raven._globalOptions.dataCallback();
            assert.isTrue(foo.calledOnce);
        });

        it('should clear globalOptions.dataCallback with no arguments', function() {
            var foo = function(){};
            Raven._globalOptions.dataCallback = foo;
            Raven.setDataCallback();
            assert.isUndefined(Raven._globalOptions.dataCallback);
        });

        it('should generate a wrapper that passes the prior callback as the 2nd argument', function () {
            var foo = sinon.stub();
            var bar = sinon.spy(function(data, orig) {
                assert.equal(orig, foo);
                foo();
            });
            Raven._globalOptions.dataCallback = foo;
            Raven.setDataCallback(bar);
            Raven._globalOptions.dataCallback({
                'a': 1 // "data"
            });
            assert.isTrue(bar.calledOnce);
            assert.isTrue(foo.calledOnce);
        });
    });

    describe('.setShouldSendCallback', function() {
        it('should set the globalOptions.shouldSendCallback attribute', function() {
            var foo = sinon.stub();
            Raven.setShouldSendCallback(foo);

            // note that setShouldSendCallback creates a callback/closure around
            // foo, so can't test for equality - just verify that calling the wrapper
            // also calls foo
            Raven._globalOptions.shouldSendCallback();
            assert.isTrue(foo.calledOnce);
        });

        it('should clear globalOptions.shouldSendCallback with no arguments', function() {
            var foo = function(){};
            Raven._globalOptions.shouldSendCallback = foo;
            Raven.setShouldSendCallback();
            assert.isUndefined(Raven._globalOptions.shouldSendCallback);
        });

        it('should generate a wrapper that passes the prior callback as the 2nd argument', function () {
            var foo = sinon.stub();
            var bar = sinon.spy(function(data, orig) {
                assert.equal(orig, foo);
                foo();
            });
            Raven._globalOptions.shouldSendCallback = foo;
            Raven.setShouldSendCallback(bar);
            Raven._globalOptions.shouldSendCallback({
                'a': 1 // "data"
            });
            assert.isTrue(bar.calledOnce);
            assert.isTrue(foo.calledOnce);
        });
    });

    describe('.captureMessage', function() {
        it('should work as advertised', function() {
            this.sinon.stub(Raven, 'isSetup').returns(true);
            this.sinon.stub(Raven, '_send');
            Raven.captureMessage('lol', {foo: 'bar'});
            assert.isTrue(Raven._send.called);
            assert.deepEqual(Raven._send.lastCall.args, [{
                message: 'lol',
                foo: 'bar'
            }]);
        });

        it('should coerce message to a string', function() {
            this.sinon.stub(Raven, 'isSetup').returns(true);
            this.sinon.stub(Raven, '_send');
            Raven.captureMessage({});
            assert.isTrue(Raven._send.called);
            assert.deepEqual(Raven._send.lastCall.args, [{
                message: '[object Object]'
            }]);
        });

        it('should work as advertised #integration', function() {
            var imageCache = [];

            this.sinon.stub(Raven, '_makeRequest');
            setupRaven();
            Raven.captureMessage('lol', {foo: 'bar'});
            assert.equal(Raven._makeRequest.callCount, 1);
            // It'd be hard to assert the actual payload being sent
            // since it includes the generated url, which is going to
            // vary between users running the tests
            // Unit tests should cover that the payload was constructed properly
        });

        it('should tag lastEventId #integration', function() {
            this.sinon.stub(Raven, '_makeRequest');
            setupRaven();
            Raven.captureMessage('lol');
            assert.equal(Raven.lastEventId(), 'abc123');
        });

        it('should respect `ignoreErrors`', function() {
            this.sinon.stub(Raven, 'isSetup').returns(true);
            this.sinon.stub(Raven, '_send');

            Raven._globalOptions.ignoreErrors = joinRegExp(['e1', 'e2']);
            Raven.captureMessage('e1');
            assert.isFalse(Raven._send.called);
            Raven.captureMessage('e2');
            assert.isFalse(Raven._send.called);
            Raven.captureMessage('Non-ignored error');
            assert.isTrue(Raven._send.calledOnce);
        });

        it('should not throw an error if not configured', function() {
            this.sinon.stub(Raven, 'isSetup').returns(false);
            this.sinon.stub(Raven, '_send');
            assert.doesNotThrow(function() {
                Raven.captureMessage('foo');
            });
        });

        it('should include a synthetic stacktrace if stacktrace:true is passed', function () {
            this.sinon.stub(Raven, 'isSetup').returns(true);
            this.sinon.stub(Raven, '_send');

            function foo() {
                Raven.captureMessage('foo', {
                    stacktrace: true
                });
            }

            foo();
            var frames = Raven._send.lastCall.args[0].stacktrace.frames;

            // Raven.captureMessage
            var last = frames[frames.length - 1];
            assert.isTrue(/(captureMessage|^\?)$/.test(last.function)); // loose equality check because differs per-browser
            assert.equal(last.in_app, false);

            // foo
            var secondLast = frames[frames.length - 2];
            assert.equal(secondLast.function, 'foo');
            assert.equal(secondLast.in_app, true);
        });
    });

    describe('.captureException', function() {
        it('should call handleStackInfo', function() {
            var error = new Error('crap');
            this.sinon.stub(Raven, 'isSetup').returns(true);
            this.sinon.stub(Raven, '_handleStackInfo');
            Raven.captureException(error, {foo: 'bar'});
            assert.isTrue(Raven._handleStackInfo.calledOnce);
        });

        it('should store the last exception', function() {
            var error = new Error('crap');
            this.sinon.stub(Raven, 'isSetup').returns(true);
            this.sinon.stub(Raven, '_handleStackInfo');
            Raven.captureException(error);
            assert.equal(Raven.lastException(), error);
        });

        it('shouldn\'t reraise the if error is the same error', function() {
            var error = new Error('crap');
            this.sinon.stub(Raven, 'isSetup').returns(true);
            this.sinon.stub(Raven, '_handleStackInfo').throws(error);
            // this would raise if the errors didn't match
            Raven.captureException(error, {foo: 'bar'});
            assert.isTrue(Raven._handleStackInfo.calledOnce);
        });

        it('should reraise a different error', function() {
            var error = new Error('crap1');
            this.sinon.stub(Raven, 'isSetup').returns(true);
            this.sinon.stub(Raven, '_handleStackInfo').throws(error);
            assert.throws(function() {
                Raven.captureException(new Error('crap2'));
            }, error);
        });

        it('should capture as a normal message if a non-Error is passed', function() {
            this.sinon.stub(Raven, 'isSetup').returns(true);
            this.sinon.stub(Raven, 'captureMessage');
            this.sinon.stub(Raven, '_handleStackInfo');
            Raven.captureException('derp');
            assert.isTrue(Raven.captureMessage.called);
            assert.equal(Raven.captureMessage.lastCall.args[0], 'derp');
            assert.isFalse(Raven._handleStackInfo.called);
            Raven.captureException(true);
            assert.isTrue(Raven.captureMessage.called);
            assert.equal(Raven.captureMessage.lastCall.args[0], true);
            assert.isFalse(Raven._handleStackInfo.called);
        });

        it('should capture custom errors that extend the Error prototype', function() {
            function NotImplementedError(message) {
                this.name = "NotImplementedError";
                this.message = message || "";
            }
            NotImplementedError.prototype = Error.prototype;

            this.sinon.stub(Raven, 'isSetup').returns(true);
            this.sinon.stub(Raven, '_handleStackInfo');

            Raven.captureException(new NotImplementedError('Bzzap'));

            assert.isTrue(Raven._handleStackInfo.calledOnce);
        });

        it('should not throw an error if not configured', function() {
            this.sinon.stub(Raven, 'isSetup').returns(false);
            this.sinon.stub(Raven, '_handleStackInfo');
            assert.doesNotThrow(function() {
                Raven.captureException(new Error('err'));
            });
        });
    });

    describe('.captureBreadcrumb', function () {
        it('should store the passed object in _breadcrumbs', function() {
            Raven.captureBreadcrumb({
                type: 'http',
                data: {
                    url: 'http://example.org/api/0/auth/',
                    status_code: 200
                }
            });

            assert.deepEqual(Raven._breadcrumbs[0], {
                type: 'http',
                timestamp: 0.1,
                data: {
                    url: 'http://example.org/api/0/auth/',
                    status_code: 200
                }
            });
        });

        it('should dequeue the oldest breadcrumb when over limit', function() {
            Raven._globalOptions.maxBreadcrumbs = 5;
            Raven._breadcrumbs = [
                { message: '1', timestamp: 0.1 },
                { message: '2', timestamp: 0.1 },
                { message: '3', timestamp: 0.1 },
                { message: '4', timestamp: 0.1 },
                { message: '5', timestamp: 0.1 }
            ];

            Raven.captureBreadcrumb({ message: 'lol' });
            assert.deepEqual(Raven._breadcrumbs, [
                { message: '2', timestamp: 0.1 },
                { message: '3', timestamp: 0.1 },
                { message: '4', timestamp: 0.1 },
                { message: '5', timestamp: 0.1 },
                { message: 'lol', timestamp: 0.1 }
            ]);
        });
    });

    describe('._captureUrlChange', function () {
        it('should create a new breadcrumb from its "from" and "to" arguments', function () {
            Raven._breadcrumbs = [];
            Raven._captureUrlChange('/foo', '/bar');
            assert.deepEqual(Raven._breadcrumbs, [
                { category: 'navigation', timestamp: 0.1, data: { from: '/foo', to: '/bar' }}
            ]);
        });

        it('should strip protocol/host if passed URLs share the same origin as location.href', function () {
            Raven._location = { href: 'http://example.com/foo' };
            Raven._breadcrumbs = [];

            Raven._captureUrlChange('http://example.com/foo', 'http://example.com/bar');
            assert.deepEqual(Raven._breadcrumbs, [
                { category: 'navigation', timestamp: 0.1, data: { from: '/foo', to: '/bar' }}
            ]);
        });
    });

    describe('.Raven.isSetup', function() {
        it('should work as advertised', function() {
            var isSetup = this.sinon.stub(Raven, 'isSetup');
            isSetup.returns(true);
            assert.isTrue(Raven.isSetup());
            isSetup.returns(false);
            assert.isFalse(Raven.isSetup());
        });
    });

    describe('.showReportDialog', function () {
        it('should throw a RavenConfigError if no eventId', function () {
            assert.throws(function () {
                Raven.showReportDialog({
                    dsn: SENTRY_DSN // dsn specified via options
                });
            }, 'Missing eventId');

            Raven.config(SENTRY_DSN);
            assert.throws(function () {
                Raven.showReportDialog(); // dsn specified via Raven.config
            }, 'Missing eventId');
        });

        it('should throw a RavenConfigError if no dsn', function () {
            assert.throws(function () {
                Raven.showReportDialog({
                    eventId: 'abc123'
                });
            }, 'Missing DSN');
        });

        describe('script tag insertion', function () {
            beforeEach(function () {
                this.appendChildStub = this.sinon.stub(document.head, 'appendChild');
            });

            it('should specify embed API endpoint and basic query string (DSN, eventId)', function () {
                this.sinon.stub(Raven, '_makeRequest');

                Raven.showReportDialog({
                    eventId: 'abc123',
                    dsn: SENTRY_DSN
                });

                var script = this.appendChildStub.getCall(0).args[0];
                assert.equal(script.src, 'http://example.com/api/embed/error-page/?eventId=abc123&dsn=http%3A%2F%2Fabc%40example.com%3A80%2F2');

                this.appendChildStub.reset();

                Raven
                    .config(SENTRY_DSN)
                    .captureException(new Error('foo')) // generates lastEventId
                    .showReportDialog();

                this.appendChildStub.getCall(0).args[0];
                assert.equal(script.src, 'http://example.com/api/embed/error-page/?eventId=abc123&dsn=http%3A%2F%2Fabc%40example.com%3A80%2F2');
            });

            it('should specify embed API endpoint and full query string (DSN, eventId, user)', function () {
                this.sinon.stub(Raven, '_makeRequest');

                Raven.showReportDialog({
                    eventId: 'abc123',
                    dsn: SENTRY_DSN,
                    user: {
                        name: 'Average Normalperson',
                        email: 'an@example.com'
                    }
                });

                var script = this.appendChildStub.getCall(0).args[0];
                assert.equal(script.src, 'http://example.com/api/embed/error-page/?eventId=abc123&dsn=http%3A%2F%2Fabc%40example.com%3A80%2F2&name=Average%20Normalperson&email=an%40example.com');

                this.appendChildStub.reset();
                Raven
                    .config(SENTRY_DSN)
                    .captureException(new Error('foo')) // generates lastEventId
                    .setUserContext({
                        name: 'Average Normalperson 2',
                        email: 'an2@example.com'
                    })
                    .showReportDialog();

                var script = this.appendChildStub.getCall(0).args[0];
                assert.equal(script.src, 'http://example.com/api/embed/error-page/?eventId=abc123&dsn=http%3A%2F%2Fabc%40example.com%3A80%2F2&name=Average%20Normalperson%202&email=an2%40example.com');
            });
        });
    });
});

// intentionally separate install/uninstall from other test methods, because
// the built-in wrapping doesn't play nice w/ Sinon's useFakeTimers() [won't
// restore setTimeout, setInterval, etc]
describe('install/uninstall', function () {
   beforeEach(function () {
      Raven = new _Raven();
   });

   describe('.install', function() {
        beforeEach(function () {
            this.sinon.stub(TraceKit.report, 'subscribe');
        });

        it('should check `Raven.isSetup`', function() {
            this.sinon.stub(Raven, 'isSetup').returns(false);
            Raven.install();
            assert.isTrue(Raven.isSetup.calledOnce);
            assert.isFalse(TraceKit.report.subscribe.calledOnce);
        });

        it('should register itself with TraceKit', function() {
            this.sinon.stub(Raven, 'isSetup').returns(true);
            this.sinon.stub(Raven, '_handleStackInfo');
            assert.equal(Raven, Raven.install());
            assert.isTrue(TraceKit.report.subscribe.calledOnce);

            // `install` subscribes to TraceKit w/ an anonymous function that
            // wraps _handleStackInfo to preserve `this`. Invoke the anonymous
            // function and verify that `_handleStackInfo` is called.
            TraceKit.report.subscribe.lastCall.args[0]();
            assert.isTrue(Raven._handleStackInfo.calledOnce);
        });

        it('should not register itself more than once', function() {
            this.sinon.stub(Raven, 'isSetup').returns(true);
            Raven.install();
            Raven.install();
            assert.isTrue(TraceKit.report.subscribe.calledOnce);
        });

        it('_instrumentBreadcrumbs should use attachEvent instead of addEventListener in IE8', function () {
            Raven._globalOptions.autoBreadcrumbs = {
                dom: true
            };

            // Maintain a ref to the old function so we can restore it later.
            var temp = document.addEventListener;

            // Test setup.
            this.sinon.stub(Raven, 'isSetup').returns(true);

            document.addEventListener = false;
            document.attachEvent = this.sinon.stub();

            // Invoke and assert.
            Raven.install();
            assert.isTrue(document.attachEvent.called);

            // Cleanup.
            document.addEventListener = temp;
        });

        it('should instrument breadcrumbs by default', function () {
            this.sinon.stub(Raven, '_instrumentBreadcrumbs');
            Raven.config(SENTRY_DSN).install();
            assert.isTrue(Raven._instrumentBreadcrumbs.calledOnce);
        });

        it('should instrument breadcrumbs if autoBreadcrumbs is an object', function () {
            this.sinon.stub(Raven, '_instrumentBreadcrumbs');
            Raven.config(SENTRY_DSN, {
                dom: true,
                location: false
            }).install();

            assert.isTrue(Raven._instrumentBreadcrumbs.calledOnce);
        });

        it('should not instrument breadcrumbs if autoBreadcrumbs is false', function () {
            this.sinon.stub(Raven, '_instrumentBreadcrumbs');
            Raven.config(SENTRY_DSN, {
                autoBreadcrumbs: false
            }).install();
            assert.isFalse(Raven._instrumentBreadcrumbs.called);
        });
    });

    describe('.uninstall', function() {
        it('should uninstall from TraceKit', function() {
            this.sinon.stub(TraceKit.report, 'uninstall');
            Raven.uninstall();
            assert.isTrue(TraceKit.report.uninstall.calledOnce);
        });

        it('should set isRavenInstalled flag to false', function() {
            Raven._isRavenInstalled = true;
            this.sinon.stub(TraceKit.report, 'uninstall');
            Raven.uninstall();
            assert.isFalse(Raven._isRavenInstalled);
        });
    });
});
