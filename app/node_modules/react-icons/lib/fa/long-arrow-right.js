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

var FaLongArrowRight = function (_React$Component) {
    _inherits(FaLongArrowRight, _React$Component);

    function FaLongArrowRight() {
        _classCallCheck(this, FaLongArrowRight);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaLongArrowRight).apply(this, arguments));
    }

    _createClass(FaLongArrowRight, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm39.285714285714285 19.93285714285714q0 0.31428571428571317-0.2228571428571442 0.5357142857142847l-8.57142857142857 7.900000000000002q-0.35714285714285765 0.31428571428571317-0.7814285714285703 0.13571428571428612-0.42428571428571615-0.20000000000000284-0.42428571428571615-0.6471428571428568v-5h-27.857142857142858q-0.3142857142857127 0-0.5142857142857127-0.1999999999999993t-0.20000000000000007-0.514285714285716v-4.285714285714285q0-0.31428571428571317 0.20000000000000007-0.514285714285716t0.5142857142857142-0.1999999999999993h27.857142857142858v-5q0-0.468571428571428 0.4242857142857126-0.6471428571428568t0.7814285714285703 0.1114285714285721l8.571428571428573 7.814285714285713q0.2228571428571442 0.22142857142857153 0.2228571428571442 0.5114285714285707z' })
                )
            );
        }
    }]);

    return FaLongArrowRight;
}(React.Component);

exports.default = FaLongArrowRight;
module.exports = exports['default'];