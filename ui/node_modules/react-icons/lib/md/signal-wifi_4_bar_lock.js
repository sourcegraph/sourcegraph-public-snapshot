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

var MdSignalWifi4BarLock = function (_React$Component) {
    _inherits(MdSignalWifi4BarLock, _React$Component);

    function MdSignalWifi4BarLock() {
        _classCallCheck(this, MdSignalWifi4BarLock);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdSignalWifi4BarLock).apply(this, arguments));
    }

    _createClass(MdSignalWifi4BarLock, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm25.86 24.14v4.375l-5.859999999999999 7.341666666666669-19.296666666666667-24.215000000000003 0.466666666666667-0.3916666666666657q0.47-0.3133333333333326 0.9783333333333333-0.663333333333334t1.3666666666666667-0.9000000000000004 1.8333333333333335-1.0133333333333336 2.2283333333333335-1.0549999999999997 2.656666666666667-1.0133333333333336 2.9333333333333336-0.8216666666666645 3.280000000000001-0.5866666666666669 3.553333333333331-0.1966666666666672 3.556666666666665 0.19666666666666632 3.2833333333333314 0.5866666666666669 2.9283333333333346 0.8200000000000003 2.6583333333333314 1.0166666666666666 2.2266666666666666 1.0533333333333337 1.8383333333333312 1.0166666666666675 1.3666666666666671 0.8966666666666665 0.9766666666666666 0.663333333333334l0.46666666666666856 0.39000000000000057-3.4383333333333326 4.374999999999998q-0.46666666666666856-0.15833333333333321-1.7166666666666686-0.15833333333333321-3.5166666666666657 0-5.899999999999999 2.3833333333333346t-2.383333333333333 5.899999999999999z m10.780000000000001 2.5v-2.5q0-1.0166666666666657-0.7416666666666671-1.7583333333333329t-1.759999999999998-0.7433333333333323-1.7583333333333329 0.7416666666666671-0.7433333333333323 1.7566666666666677v2.5h5z m1.7199999999999989 0q0.625 0 1.1333333333333329 0.5466666666666669t0.5083333333333329 1.1716666666666669v6.641666666666666q0 0.6233333333333348-0.5066666666666677 1.1333333333333329t-1.1333333333333329 0.5066666666666677h-8.361666666666665q-0.625 0-1.1333333333333329-0.5083333333333329t-0.5083333333333329-1.1333333333333329v-6.641666666666666q0-0.625 0.5083333333333329-1.1716666666666669t1.1333333333333329-0.5466666666666669v-2.5q0-1.7966666666666669 1.1716666666666669-2.9666666666666686t2.9666666666666686-1.173333333333332 3.009999999999998 1.211666666666666 1.2100000000000009 2.9299999999999997v2.5z' })
                )
            );
        }
    }]);

    return MdSignalWifi4BarLock;
}(React.Component);

exports.default = MdSignalWifi4BarLock;
module.exports = exports['default'];