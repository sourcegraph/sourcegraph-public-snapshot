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

var TiVolumeDown = function (_React$Component) {
    _inherits(TiVolumeDown, _React$Component);

    function TiVolumeDown() {
        _classCallCheck(this, TiVolumeDown);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiVolumeDown).apply(this, arguments));
    }

    _createClass(TiVolumeDown, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm25.23 9.706666666666667c-0.75 0-1.5083333333333329 0.25333333333333385-2.2600000000000016 0.7550000000000008l-4.453333333333333 2.966666666666667c-1.254999999999999 0.841666666666665-3.676666666666666 1.5716666666666654-5.183333333333332 1.5716666666666654-2.756666666666666 0-5 2.2433333333333323-5 5v3.333333333333332c0 2.7566666666666677 2.243333333333334 5 5 5 1.506666666666666 0 3.928333333333333 0.7333333333333343 5.183333333333332 1.5666666666666664l4.449999999999999 2.969999999999999c0.7533333333333339 0.5 1.5133333333333319 0.7550000000000026 2.2616666666666667 0.7550000000000026 1.4966666666666661 0 3.1050000000000004-1.1333333333333329 3.1050000000000004-3.625v-16.666666666666664c0-2.491666666666669-1.6083333333333343-3.6266666666666687-3.1033333333333353-3.6266666666666687z m-11.896666666666667 15.293333333333333c-0.9199999999999999 0-1.666666666666666-0.7466666666666661-1.666666666666666-1.6666666666666679v-3.333333333333332c0-0.9200000000000017 0.7466666666666661-1.6666666666666679 1.666666666666666-1.6666666666666679 2.0166666666666675 0 4.845000000000001-0.8249999999999993 6.666666666666666-1.9100000000000001v10.488333333333333c-1.8216666666666654-1.086666666666666-4.649999999999999-1.9116666666666653-6.666666666666666-1.9116666666666653z m11.666666666666666 5c0 0.07666666666666799-0.0033333333333338544 0.14333333333333442-0.010000000000001563 0.1999999999999993-0.05000000000000071-0.026666666666667282-0.10833333333333428-0.05999999999999872-0.173333333333332-0.10333333333333172l-3.1499999999999986-2.1000000000000014v-12.659999999999998l3.1499999999999986-2.0999999999999996c0.06666666666666643-0.043333333333333 0.12333333333333485-0.07833333333333314 0.1750000000000007-0.10500000000000043 0.004999999999999005 0.05666666666666664 0.00833333333333286 0.12333333333333307 0.00833333333333286 0.1999999999999993v16.666666666666668z m5.486666666666672-12.84333333333333c-0.6499999999999986 0.6499999999999986-0.6499999999999986 1.7049999999999983 0.0033333333333338544 2.3566666666666656 0.5749999999999993 0.5749999999999993 0.8916666666666657 1.3383333333333347 0.8916666666666657 2.1499999999999986s-0.31666666666666643 1.5833333333333321-0.8949999999999996 2.1583333333333314c-0.6499999999999986 0.6499999999999986-0.6499999999999986 1.7049999999999983 0 2.3566666666666656 0.3249999999999993 0.3249999999999993 0.75 0.4883333333333333 1.1783333333333346 0.4883333333333333s0.8533333333333317-0.163333333333334 1.1783333333333346-0.4883333333333333c1.2066666666666634-1.2050000000000018 1.8699999999999974-2.8083333333333336 1.8699999999999974-4.513333333333332s-0.6666666666666643-3.3083333333333336-1.8733333333333348-4.513333333333335c-0.6499999999999986-0.6499999999999986-1.7049999999999983-0.6499999999999986-2.3566666666666656 0.0033333333333338544z' })
                )
            );
        }
    }]);

    return TiVolumeDown;
}(React.Component);

exports.default = TiVolumeDown;
module.exports = exports['default'];