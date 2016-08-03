'use strict';

exports.__esModule = true;

var _createClass = (function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ('value' in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; })();

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError('Cannot call a class as a function'); } }

function _inherits(subClass, superClass) { if (typeof superClass !== 'function' && superClass !== null) { throw new TypeError('Super expression must either be null or a function, not ' + typeof superClass); } subClass.prototype = Object.create(superClass && superClass.prototype, { constructor: { value: subClass, enumerable: false, writable: true, configurable: true } }); if (superClass) Object.setPrototypeOf ? Object.setPrototypeOf(subClass, superClass) : subClass.__proto__ = superClass; }

var _supportsColor = require('supports-color');

var _supportsColor2 = _interopRequireDefault(_supportsColor);

var _warnOnce = require('./warn-once');

var _warnOnce2 = _interopRequireDefault(_warnOnce);

var CssSyntaxError = (function (_SyntaxError) {
    _inherits(CssSyntaxError, _SyntaxError);

    function CssSyntaxError(message, line, column, source, file, plugin) {
        _classCallCheck(this, CssSyntaxError);

        _SyntaxError.call(this, message);
        this.name = 'CssSyntaxError';
        this.reason = message;

        if (file) this.file = file;
        if (source) this.source = source;
        if (plugin) this.plugin = plugin;
        if (typeof line !== 'undefined' && typeof column !== 'undefined') {
            this.line = line;
            this.column = column;
        }

        this.setMessage();

        if (Error.captureStackTrace) {
            Error.captureStackTrace(this, CssSyntaxError);
        }
    }

    CssSyntaxError.prototype.setMessage = function setMessage() {
        this.message = this.plugin ? this.plugin + ': ' : '';
        this.message += this.file ? this.file : '<css input>';
        if (typeof this.line !== 'undefined') {
            this.message += ':' + this.line + ':' + this.column;
        }
        this.message += ': ' + this.reason;
    };

    CssSyntaxError.prototype.showSourceCode = function showSourceCode(color) {
        if (!this.source) return '';

        var num = this.line - 1;
        var lines = this.source.split('\n');

        var prev = num > 0 ? lines[num - 1] + '\n' : '';
        var broken = lines[num];
        var next = num < lines.length - 1 ? '\n' + lines[num + 1] : '';

        var mark = '\n';
        for (var i = 0; i < this.column - 1; i++) {
            mark += ' ';
        }

        if (typeof color === 'undefined') color = _supportsColor2['default'];
        if (color) {
            mark += '\x1B[1;31m^\x1B[0m';
        } else {
            mark += '^';
        }

        return '\n' + prev + broken + mark + next;
    };

    CssSyntaxError.prototype.setMozillaProps = function setMozillaProps() {
        var sample = Error.call(this, this.message);
        if (sample.columnNumber) this.columnNumber = this.column;
        if (sample.description) this.description = this.message;
        if (sample.lineNumber) this.lineNumber = this.line;
        if (sample.fileName) this.fileName = this.file;
    };

    CssSyntaxError.prototype.toString = function toString() {
        return this.name + ': ' + this.message + this.showSourceCode();
    };

    _createClass(CssSyntaxError, [{
        key: 'generated',
        get: function get() {
            _warnOnce2['default']('CssSyntaxError#generated is depreacted. Use input instead.');
            return this.input;
        }
    }]);

    return CssSyntaxError;
})(SyntaxError);

exports['default'] = CssSyntaxError;
module.exports = exports['default'];