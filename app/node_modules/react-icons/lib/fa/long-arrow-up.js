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

var FaLongArrowUp = function (_React$Component) {
    _inherits(FaLongArrowUp, _React$Component);

    function FaLongArrowUp() {
        _classCallCheck(this, FaLongArrowUp);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaLongArrowUp).apply(this, arguments));
    }

    _createClass(FaLongArrowUp, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.504285714285714 11.004285714285714q-0.19999999999999574 0.4242857142857144-0.6471428571428568 0.4242857142857144h-5v27.857142857142854q0 0.3142857142857167-0.1999999999999993 0.5142857142857125t-0.514285714285716 0.20000000000000284h-4.285714285714285q-0.31428571428571317 0-0.514285714285716-0.20000000000000284t-0.1999999999999993-0.5142857142857125v-27.857142857142854h-5q-0.468571428571428-1.7763568394002505e-15-0.6471428571428568-0.42428571428571615t0.1114285714285721-0.781428571428572l7.814285714285713-8.571428571428571q0.22142857142857153-0.22285714285714264 0.5100000000000016-0.22285714285714264 0.31428571428571317 0 0.5357142857142847 0.22285714285714286l7.924285714285713 8.571428571428571q0.28999999999999915 0.35714285714285765 0.1114285714285721 0.781428571428572z' })
                )
            );
        }
    }]);

    return FaLongArrowUp;
}(React.Component);

exports.default = FaLongArrowUp;
module.exports = exports['default'];