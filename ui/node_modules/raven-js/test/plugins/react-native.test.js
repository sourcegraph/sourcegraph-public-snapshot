var Promise = require('bluebird');

var _Raven = require('../../src/raven');
var reactNativePlugin = require('../../plugins/react-native');

window.ErrorUtils = {};

var Raven;
describe('React Native plugin', function () {
    beforeEach(function () {
        Raven = new _Raven();
        Raven.config('http://abc@example.com:80/2');

        reactNativePlugin._persistPayload = self.sinon.stub().returns(Promise.resolve());
        reactNativePlugin._restorePayload = self.sinon.stub().returns(Promise.resolve());
        reactNativePlugin._clearPayload = self.sinon.stub().returns(Promise.resolve());
    });

    describe('_normalizeData()', function () {
        it('should normalize culprit and frame filenames/URLs from app', function () {
            var data = {
                project: '2',
                logger: 'javascript',
                platform: 'javascript',

                culprit: 'file:///var/mobile/Containers/Bundle/Application/ABC/123.app/app.js',
                message: 'Error: crap',
                exception: {
                    type: 'Error',
                    values: [{
                        stacktrace: {
                            frames: [{
                                filename: 'file:///var/containers/Bundle/Application/ABC/123.app/file1.js',
                                lineno: 10,
                                colno: 11,
                                'function': 'broken'

                            }, {
                                filename: 'file:///var/mobile/Containers/Bundle/Application/ABC/123.app/file2.js',
                                lineno: 12,
                                colno: 13,
                                'function': 'lol'
                            }]
                        }
                    }],
                }
            };
            reactNativePlugin._normalizeData(data);

            assert.equal(data.culprit, '/app.js');
            var frames = data.exception.values[0].stacktrace.frames;
            assert.equal(frames[0].filename, '/file1.js');
            assert.equal(frames[1].filename, '/file2.js');
        });

        it('should normalize culprit and frame filenames/URLs from CodePush', function () {
            var data = {
                project: '2',
                logger: 'javascript',
                platform: 'javascript',

                culprit: 'file:///var/mobile/Containers/Data/Application/ABC/Library/Application%20Support/CodePush/CDE/CodePush/app.js',
                message: 'Error: crap',
                exception: {
                    type: 'Error',
                    values: [{
                        stacktrace: {
                            frames: [{
                                filename: 'file:///var/mobile/Containers/Data/Application/ABC/Library/Application%20Support/CodePush/CDE/CodePush/file1.js',
                                lineno: 10,
                                colno: 11,
                                'function': 'broken'

                            }, {
                                filename: 'file:///var/mobile/Containers/Data/Application/ABC/Library/Application%20Support/CodePush/CDE/CodePush/file2.js',
                                lineno: 12,
                                colno: 13,
                                'function': 'lol'
                            }]
                        }
                    }],
                }
            };
            reactNativePlugin._normalizeData(data);

            assert.equal(data.culprit, '/app.js');
            var frames = data.exception.values[0].stacktrace.frames;
            assert.equal(frames[0].filename, '/file1.js');
            assert.equal(frames[1].filename, '/file2.js');
        });
    });

    describe('_transport()', function () {
        beforeEach(function () {
            this.xhr = sinon.useFakeXMLHttpRequest();
            var requests = this.requests = [];

            this.xhr.onCreate = function (xhr) {
              requests.push(xhr);
            };
        });

        afterEach(function () {
            this.xhr.restore();
        });

        it('should open and send a new XHR POST with urlencoded auth, fake origin', function () {
            reactNativePlugin._transport({
                url: 'https://example.org/1',
                auth: {
                    sentry_version: '7',
                    sentry_client: 'raven-js/2.2.0',
                    sentry_key: 'abc123'
                },
                data: {foo: 'bar'}
            });

            var lastXhr = this.requests.shift();
            lastXhr.respond(200);

            assert.equal(
                lastXhr.url,
                'https://example.org/1?sentry_version=7&sentry_client=raven-js%2F2.2.0&sentry_key=abc123'
            );
            assert.equal(lastXhr.method, 'POST');
            assert.equal(lastXhr.requestBody, '{"foo":"bar"}');
            assert.equal(lastXhr.requestHeaders['Content-type'], 'application/json');
            assert.equal(lastXhr.requestHeaders['Origin'], 'react-native://');
        });

        it('should call onError callback on failure', function () {
            var onError = this.sinon.stub();
            var onSuccess = this.sinon.stub();
            reactNativePlugin._transport({
                url: 'https://example.org/1',
                auth: {},
                data: {foo: 'bar'},
                onError: onError,
                onSuccess: onSuccess
            });

            var lastXhr = this.requests.shift();
            lastXhr.respond(401);

            assert.isTrue(onError.calledOnce);
            assert.isFalse(onSuccess.calledOnce);
        });

        it('should call onSuccess callback on success', function () {
            var onError = this.sinon.stub();
            var onSuccess = this.sinon.stub();
            reactNativePlugin._transport({
                url: 'https://example.org/1',
                auth: {},
                data: {foo: 'bar'},
                onError: onError,
                onSuccess: onSuccess
            });

            var lastXhr = this.requests.shift();
            lastXhr.respond(200);

            assert.isTrue(onSuccess.calledOnce);
            assert.isFalse(onError.calledOnce);
        });
    });

    describe('ErrorUtils global error handler', function () {
        beforeEach(function () {
            var self = this;
            ErrorUtils.setGlobalHandler = function(fn) {
                self.globalErrorHandler = fn;
            };
            self.defaultErrorHandler = self.sinon.stub();
            ErrorUtils.getGlobalHandler = function () {
                return self.defaultErrorHandler;
            }
        });

        it('checks for persisted errors when starting', function () {
            var onInit = self.sinon.stub();
            reactNativePlugin(Raven, {onInitialize: onInit});
            assert.isTrue(reactNativePlugin._restorePayload.calledOnce);

            return Promise.resolve().then(function () {
                assert.isTrue(onInit.calledOnce);
            });
        });

        it('reports persisted errors', function () {
            var payload = {abc: 123};
            self.sinon.stub(Raven, '_sendProcessedPayload');
            reactNativePlugin._restorePayload = self.sinon.stub().returns(Promise.resolve(payload));
            var onInit = self.sinon.stub();
            reactNativePlugin(Raven, {onInitialize: onInit});

            return Promise.resolve().then(function () {
                assert.isTrue(onInit.calledOnce);
                assert.equal(onInit.getCall(0).args[0], payload);
                assert.isTrue(Raven._sendProcessedPayload.calledOnce);
                assert.equal(Raven._sendProcessedPayload.getCall(0).args[0], payload);
            });
        });

        it('clears persisted errors after they are reported', function () {
            var payload = {abc: 123};
            var callback;
            self.sinon.stub(Raven, '_sendProcessedPayload', function(p, cb) { callback = cb; });
            reactNativePlugin._restorePayload = self.sinon.stub().returns(Promise.resolve(payload));

            reactNativePlugin(Raven);

            return Promise.resolve().then(function () {
                assert.isFalse(reactNativePlugin._clearPayload.called);
                callback();
                assert.isTrue(reactNativePlugin._clearPayload.called);
            });
        });

        it('does not clear persisted errors if there is an error reporting', function () {
            var payload = {abc: 123};
            var callback;
            self.sinon.stub(Raven, '_sendProcessedPayload', function(p, cb) { callback = cb; });
            reactNativePlugin._restorePayload = self.sinon.stub().returns(Promise.resolve(payload));

            reactNativePlugin(Raven);

            return Promise.resolve().then(function () {
                assert.isFalse(reactNativePlugin._clearPayload.called);
                callback(new Error('nope'));
                assert.isFalse(reactNativePlugin._clearPayload.called);
            });
        });

        describe('in development mode', function () {
            beforeEach(function () {
                global.__DEV__ = true;
            });

            it('should call the default React Native handler and Raven.captureException', function () {
                reactNativePlugin(Raven);
                var err = new Error();
                this.sinon.stub(Raven, 'captureException');

                this.globalErrorHandler(err, true);

                assert.isTrue(this.defaultErrorHandler.calledOnce);
                assert.isTrue(Raven.captureException.calledOnce);
                assert.equal(Raven.captureException.getCall(0).args[0], err);
            });
        });

        describe('in production mode', function () {
            beforeEach(function () {
                global.__DEV__ = false;
            });

            it('should call the default React Native handler after persisting the error', function () {
                reactNativePlugin(Raven);
                var err = new Error();
                this.globalErrorHandler(err, true);

                assert.isTrue(reactNativePlugin._persistPayload.calledOnce);

                var defaultErrorHandler = this.defaultErrorHandler;
                return Promise.resolve().then(function () {
                    assert.isTrue(defaultErrorHandler.calledOnce);
                });
            });
        });
    });
});
