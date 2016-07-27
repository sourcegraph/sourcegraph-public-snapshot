"use strict";
var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
var arrify = require('arrify');
var extend = require('xtend');
var Promise = require('any-promise');
var throwback_1 = require('throwback');
var base_1 = require('./base');
var error_1 = require('./error');
var Request = (function (_super) {
    __extends(Request, _super);
    function Request(options) {
        var _this = this;
        _super.call(this, options);
        this.middleware = [];
        this.opened = false;
        this.aborted = false;
        this.uploaded = 0;
        this.downloaded = 0;
        this.uploadedBytes = null;
        this.downloadedBytes = null;
        this.uploadLength = null;
        this.downloadLength = null;
        this._progress = [];
        this.timeout = (options.timeout | 0);
        this.method = (options.method || 'GET').toUpperCase();
        this.body = options.body;
        var promised = new Promise(function (resolve, reject) {
            _this._resolve = resolve;
            _this._reject = reject;
        });
        this._promise = Promise.resolve()
            .then(function () {
            var run = throwback_1.compose(_this.middleware);
            var cb = function () {
                _this._handle();
                return promised;
            };
            return run(_this, cb);
        });
        this.transport = extend(options.transport);
        this.use(options.use || this.transport.use);
        this.progress(options.progress);
    }
    Request.prototype.error = function (message, code, original) {
        return new error_1.default(message, code, original, this);
    };
    Request.prototype.then = function (onFulfilled, onRejected) {
        return this._promise.then(onFulfilled, onRejected);
    };
    Request.prototype.catch = function (onRejected) {
        return this._promise.then(null, onRejected);
    };
    Request.prototype.exec = function (cb) {
        this.then(function (response) {
            cb(null, response);
        }, cb);
    };
    Request.prototype.toOptions = function () {
        return {
            url: this.url,
            method: this.method,
            body: this.body,
            transport: this.transport,
            timeout: this.timeout,
            rawHeaders: this.rawHeaders,
            use: this.middleware,
            progress: this._progress
        };
    };
    Request.prototype.toJSON = function () {
        return {
            url: this.url,
            headers: this.headers,
            body: this.body,
            timeout: this.timeout,
            method: this.method
        };
    };
    Request.prototype.clone = function () {
        return new Request(this.toOptions());
    };
    Request.prototype.use = function (fns) {
        for (var _i = 0, _a = arrify(fns); _i < _a.length; _i++) {
            var fn = _a[_i];
            this.middleware.push(fn);
        }
        return this;
    };
    Request.prototype.progress = function (fns) {
        for (var _i = 0, _a = arrify(fns); _i < _a.length; _i++) {
            var fn = _a[_i];
            this._progress.push(fn);
        }
        return this;
    };
    Request.prototype.abort = function () {
        if (this.completed === 1 || this.aborted) {
            return;
        }
        this.aborted = true;
        if (this.opened) {
            this._emit();
            if (this.transport.abort) {
                this.transport.abort(this);
            }
        }
        this._reject(this.error('Request aborted', 'EABORT'));
        return this;
    };
    Request.prototype._emit = function () {
        var fns = this._progress;
        try {
            for (var _i = 0, fns_1 = fns; _i < fns_1.length; _i++) {
                var fn = fns_1[_i];
                fn(this);
            }
        }
        catch (err) {
            this._reject(err);
            this.abort();
        }
    };
    Request.prototype._handle = function () {
        var _this = this;
        var _a = this, timeout = _a.timeout, url = _a.url;
        var timer;
        if (this.aborted) {
            return;
        }
        this.opened = true;
        if (/^https?\:\/*(?:[~#\\\?;\:]|$)/.test(url)) {
            this._reject(this.error("Refused to connect to invalid URL \"" + url + "\"", 'EINVALID'));
            return;
        }
        if (timeout > 0) {
            timer = setTimeout(function () {
                _this._reject(_this.error("Timeout of " + timeout + "ms exceeded", 'ETIMEOUT'));
                _this.abort();
            }, timeout);
        }
        return this.transport.open(this)
            .then(function (res) { return _this._resolve(res); }, function (err) { return _this._reject(err); });
    };
    Object.defineProperty(Request.prototype, "completed", {
        get: function () {
            return (this.uploaded + this.downloaded) / 2;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(Request.prototype, "completedBytes", {
        get: function () {
            return this.uploadedBytes + this.downloadedBytes;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(Request.prototype, "totalBytes", {
        get: function () {
            return this.uploadLength + this.downloadLength;
        },
        enumerable: true,
        configurable: true
    });
    Request.prototype._setUploadedBytes = function (bytes, uploaded) {
        if (bytes !== this.uploadedBytes) {
            this.uploaded = uploaded || bytes / this.uploadLength;
            this.uploadedBytes = bytes;
            this._emit();
        }
    };
    Request.prototype._setDownloadedBytes = function (bytes, downloaded) {
        if (bytes !== this.downloadedBytes) {
            this.downloaded = downloaded || bytes / this.downloadLength;
            this.downloadedBytes = bytes;
            this._emit();
        }
    };
    return Request;
}(base_1.default));
Object.defineProperty(exports, "__esModule", { value: true });
exports.default = Request;
//# sourceMappingURL=request.js.map