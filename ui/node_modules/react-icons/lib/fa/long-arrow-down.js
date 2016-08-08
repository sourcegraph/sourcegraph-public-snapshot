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

var FaLongArrowDown = function (_React$Component) {
    _inherits(FaLongArrowDown, _React$Component);

    function FaLongArrowDown() {
        _classCallCheck(this, FaLongArrowDown);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaLongArrowDown).apply(this, arguments));
    }

    _createClass(FaLongArrowDown, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.504285714285714 28.995714285714286q0.17857142857142705 0.4242857142857126-0.1114285714285721 0.7814285714285703l-7.814285714285713 8.57142857142857q-0.22142857142857153 0.2228571428571442-0.5100000000000016 0.2228571428571442-0.31428571428571317 0-0.5357142857142847-0.2228571428571442l-7.924285714285714-8.57142857142857q-0.2900000000000009-0.35714285714285765-0.1114285714285721-0.7814285714285703 0.1985714285714284-0.4242857142857126 0.6457142857142859-0.4242857142857126h5v-27.857142857142858q0-0.3142857142857153 0.1999999999999993-0.5142857142857152t0.514285714285716-0.20000000000000007h4.285714285714285q0.31428571428571317 0 0.514285714285716 0.2t0.1999999999999993 0.5142857142857142v27.857142857142858h5q0.46857142857142975 0 0.6471428571428568 0.4242857142857126z' })
                )
            );
        }
    }]);

    return FaLongArrowDown;
}(React.Component);

exports.default = FaLongArrowDown;
module.exports = exports['default'];