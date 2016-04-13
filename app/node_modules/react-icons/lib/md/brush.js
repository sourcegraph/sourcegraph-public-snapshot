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

var MdBrush = function (_React$Component) {
    _inherits(MdBrush, _React$Component);

    function MdBrush() {
        _classCallCheck(this, MdBrush);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdBrush).apply(this, arguments));
    }

    _createClass(MdBrush, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm34.53333333333333 7.733333333333333q0.46666666666666856 0.46999999999999975 0.46666666666666856 1.173333333333333t-0.46666666666666856 1.1716666666666669l-14.924999999999997 14.921666666666667-4.608333333333334-4.609999999999999 14.921666666666667-14.921666666666667q0.466666666666665-0.4666666666666668 1.1716666666666669-0.4666666666666668t1.1716666666666669 0.4666666666666668z m-22.894999999999996 15.628333333333334q2.0333333333333314 0 3.5166666666666657 1.4833333333333343t1.4833333333333343 3.5166666666666657q0 2.7333333333333343-1.9533333333333331 4.688333333333336t-4.685000000000002 1.9499999999999957q-4.066666666666666 0-6.641666666666667-3.3583333333333343 1.1716666666666669 0 2.2266666666666675-0.8999999999999986t1.0566666666666666-2.383333333333333q0-2.030000000000001 1.4833333333333325-3.5133333333333354t3.5166666666666657-1.4833333333333343z' })
                )
            );
        }
    }]);

    return MdBrush;
}(React.Component);

exports.default = MdBrush;
module.exports = exports['default'];