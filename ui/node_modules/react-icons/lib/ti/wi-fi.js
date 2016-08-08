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

var TiWiFi = function (_React$Component) {
    _inherits(TiWiFi, _React$Component);

    function TiWiFi() {
        _classCallCheck(this, TiWiFi);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiWiFi).apply(this, arguments));
    }

    _createClass(TiWiFi, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm22.356666666666666 32.35333333333333c1.3049999999999997-1.3000000000000007 1.3049999999999997-3.411666666666669 0-4.710000000000001-1.3000000000000007-1.3083333333333336-3.416666666666668-1.3083333333333336-4.713333333333335-0.0033333333333338544-1.3049999999999997 1.3049999999999997-1.3049999999999997 3.416666666666668 0 4.716666666666665 1.3000000000000007 1.2999999999999972 3.416666666666668 1.3033333333333346 4.713333333333331-0.0033333333333303017z m11.785-13.161666666666662c-0.8533333333333317 0-1.7066666666666634-0.3249999999999993-2.3566666666666656-0.9766666666666666-6.496666666666666-6.5-17.070000000000004-6.5-23.57 0-1.3033333333333337 1.3000000000000007-3.413333333333334 1.3000000000000007-4.716666666666667 0-1.2999999999999998-1.3033333333333346-1.2999999999999998-3.413333333333334 0-4.716666666666667 9.1-9.096666666666668 23.903333333333336-9.096666666666668 33 0 1.2999999999999972 1.3000000000000007 1.2999999999999972 3.411666666666667 0 4.713333333333333-0.6499999999999986 0.6499999999999986-1.5 0.9783333333333317-2.354999999999997 0.9783333333333317z m-21.21333333333333 7.069999999999997c-0.8533333333333353 0-1.7066666666666688-0.3249999999999993-2.356666666666669-0.9766666666666666-1.3000000000000007-1.3000000000000007-1.3000000000000007-3.41 0-4.711666666666666 5.196666666666665-5.199999999999999 13.656666666666666-5.199999999999999 18.855 0 1.3000000000000007 1.3000000000000007 1.3000000000000007 3.4116666666666653 0 4.716666666666665s-3.4116666666666653 1.3000000000000007-4.716666666666665 0c-2.5966666666666676-2.6000000000000014-6.826666666666668-2.6000000000000014-9.426666666666666 0-0.6500000000000004 0.6499999999999986-1.5 0.9733333333333327-2.3550000000000004 0.9733333333333327z' })
                )
            );
        }
    }]);

    return TiWiFi;
}(React.Component);

exports.default = TiWiFi;
module.exports = exports['default'];