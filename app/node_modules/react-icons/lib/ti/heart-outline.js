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

var TiHeartOutline = function (_React$Component) {
    _inherits(TiHeartOutline, _React$Component);

    function TiHeartOutline() {
        _classCallCheck(this, TiHeartOutline);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiHeartOutline).apply(this, arguments));
    }

    _createClass(TiHeartOutline, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 33.333333333333336c-0.3249999999999993 0-0.6499999999999986-0.09499999999999886-0.9333333333333336-0.28666666666666885-0.37666666666666515-0.25-9.183333333333334-6.216666666666669-11.911666666666667-8.950000000000003-3.0500000000000007-3.0533333333333346-3.4050000000000002-6.296666666666667-3.4050000000000002-8.471666666666668 8.881784197001252e-16-4.941666666666663 4.016666666666668-8.958333333333329 8.958333333333336-8.958333333333329 3.003333333333332-8.881784197001252e-16 5.663333333333334 1.4833333333333325 7.291666666666664 3.759999999999998 1.6283333333333339-2.2766666666666655 4.288333333333334-3.759999999999999 7.291666666666668-3.759999999999999 4.940000000000001 0 8.958333333333332 4.016666666666667 8.958333333333332 8.958333333333332 0 2.1750000000000007-0.3533333333333317 5.416666666666668-3.405000000000001 8.469999999999999-2.7333333333333343 2.7333333333333343-11.538333333333334 8.700000000000003-11.91 8.950000000000003-0.283333333333335 0.19333333333333513-0.6083333333333343 0.288333333333334-0.9333333333333336 0.288333333333334z m-7.291666666666666-23.333333333333336c-3.1000000000000014 0-5.625 2.5233333333333334-5.625 5.625 0 1.8216666666666654 0.2883333333333331 3.9733333333333327 2.428333333333333 6.113333333333333 2.0216666666666665 2.020000000000003 8.138333333333335 6.291666666666671 10.488333333333333 7.911666666666665 2.3500000000000014-1.620000000000001 8.466666666666669-5.891666666666666 10.488333333333333-7.911666666666669 2.139999999999997-2.1400000000000006 2.428333333333331-4.291666666666668 2.428333333333331-6.113333333333333 7.105427357601002e-15-3.099999999999996-2.52333333333333-5.6249999999999964-5.6249999999999964-5.6249999999999964s-5.625 2.5233333333333334-5.625 5.625c0 0.9200000000000017-0.745000000000001 1.6666666666666679-1.6666666666666679 1.6666666666666679s-1.6666666666666679-0.7466666666666661-1.6666666666666679-1.666666666666666c3.552713678800501e-15-3.1000000000000014-2.52333333333333-5.625000000000002-5.624999999999998-5.625000000000002z' })
                )
            );
        }
    }]);

    return TiHeartOutline;
}(React.Component);

exports.default = TiHeartOutline;
module.exports = exports['default'];