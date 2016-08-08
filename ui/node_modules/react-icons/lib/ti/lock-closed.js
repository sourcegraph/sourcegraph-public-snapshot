'use strict';

Object.defineProperty(exports, "__esModule", {
    value: true
});

var _extends = Object.assign || function (target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i]; for (var key in source) { if (Object.prototype.hasOwnProperty.call(source, key)) { target[key] = source[key]; } } } return target; };

var _createClass = function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; }();

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

function _possibleConstructorReturn(self, call) { if (!self) { throw new ReferenceError("this hasn't been initialised - super() hasn't been called"); } return call && (typeof call === "object" || typeof call === "function") ? call : self; }

function _inherits(subClass, superClass) { if (typeof superClass !== "function" && superClass !== null) { throw new TypeError("Super expression must either be null or a function, not " + typeof superClass); } subClass.prototype = Object.create(superClass && superClass.prototype, { constructor: { value: subClass, enumerable: false, writable: true, configurable: true } }); if (superClass) Object.setPrototypeOf ? Object.setPrototypeOf(subClass, superClass) : subClass.__proto__ = superClass; }

var React = require('react');
var IconBase = require('react-icon-base');

var TiLockClosed = function (_React$Component) {
    _inherits(TiLockClosed, _React$Component);

    function TiLockClosed() {
        _classCallCheck(this, TiLockClosed);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiLockClosed).apply(this, arguments));
    }

    _createClass(TiLockClosed, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.333333333333336 16.666666666666668h-1.6666666666666679v-3.333333333333334c0-3.6750000000000007-2.9899999999999984-6.666666666666667-6.666666666666668-6.666666666666667s-6.666666666666666 2.9916666666666663-6.666666666666666 6.666666666666667v3.333333333333334h-1.666666666666666c-1.838333333333333 0-3.333333333333334 1.4933333333333323-3.333333333333334 3.333333333333332v11.666666666666668c0 1.8399999999999999 1.4949999999999992 3.333333333333332 3.333333333333334 3.333333333333332h16.666666666666668c1.8383333333333347 0 3.333333333333332-1.4933333333333323 3.333333333333332-3.333333333333332v-11.666666666666668c0-1.8399999999999999-1.495000000000001-3.333333333333332-3.333333333333332-3.333333333333332z m-8.333333333333336 13.833333333333336c-1.1999999999999993 0-2.166666666666668-0.966666666666665-2.166666666666668-2.166666666666668s0.966666666666665-2.166666666666668 2.166666666666668-2.166666666666668 2.166666666666668 0.966666666666665 2.166666666666668 2.166666666666668-0.966666666666665 2.166666666666668-2.166666666666668 2.166666666666668z m3.333333333333332-12.166666666666668h-6.666666666666668v-5.000000000000002c0-1.8399999999999999 1.495000000000001-3.333333333333334 3.333333333333332-3.333333333333334s3.333333333333332 1.493333333333334 3.333333333333332 3.333333333333334v5.000000000000002z' })
                )
            );
        }
    }]);

    return TiLockClosed;
}(React.Component);

exports.default = TiLockClosed;
module.exports = exports['default'];