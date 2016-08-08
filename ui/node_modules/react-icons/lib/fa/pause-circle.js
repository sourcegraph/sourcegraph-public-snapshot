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

var FaPauseCircle = function (_React$Component) {
    _inherits(FaPauseCircle, _React$Component);

    function FaPauseCircle() {
        _classCallCheck(this, FaPauseCircle);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaPauseCircle).apply(this, arguments));
    }

    _createClass(FaPauseCircle, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm18.571428571428573 26.42857142857143v-12.857142857142858q0-0.31428571428571495-0.1999999999999993-0.5142857142857142t-0.514285714285716-0.20000000000000107h-5.714285714285715q-0.31428571428571495 0-0.5142857142857142 0.1999999999999993t-0.1999999999999993 0.5142857142857142v12.85714285714286q0 0.31428571428571317 0.1999999999999993 0.5142857142857125t0.5142857142857142 0.1999999999999993h5.714285714285715q0.31428571428571317 0 0.514285714285716-0.1999999999999993t0.1999999999999993-0.5142857142857125z m10 0v-12.857142857142858q0-0.31428571428571495-0.1999999999999993-0.5142857142857142t-0.514285714285716-0.20000000000000107h-5.714285714285715q-0.31428571428571317 0-0.514285714285716 0.1999999999999993t-0.19999999999999574 0.5142857142857142v12.85714285714286q0 0.31428571428571317 0.1999999999999993 0.5142857142857125t0.514285714285716 0.1999999999999993h5.714285714285715q0.31428571428571317 0 0.5142857142857125-0.1999999999999993t0.1999999999999993-0.5142857142857125z m8.571428571428573-6.428571428571431q0 4.665714285714287-2.299999999999997 8.604285714285716t-6.237142857142857 6.238571428571426-8.605714285714292 2.3000000000000043-8.6-2.3000000000000043-6.242857142857143-6.238571428571426-2.295714285714286-8.604285714285716 2.3000000000000003-8.604285714285714 6.234285714285714-6.238571428571428 8.604285714285714-2.3000000000000003 8.605714285714285 2.3000000000000003 6.238571428571426 6.238571428571428 2.298571428571435 8.604285714285714z' })
                )
            );
        }
    }]);

    return FaPauseCircle;
}(React.Component);

exports.default = FaPauseCircle;
module.exports = exports['default'];