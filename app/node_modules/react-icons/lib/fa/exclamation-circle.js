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

var FaExclamationCircle = function (_React$Component) {
    _inherits(FaExclamationCircle, _React$Component);

    function FaExclamationCircle() {
        _classCallCheck(this, FaExclamationCircle);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaExclamationCircle).apply(this, arguments));
    }

    _createClass(FaExclamationCircle, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 2.857142857142857q4.665714285714287 0 8.604285714285716 2.3000000000000003t6.238571428571426 6.2371428571428575 2.3000000000000043 8.605714285714285-2.299999999999997 8.602857142857143-6.238571428571429 6.238571428571429-8.60428571428572 2.297142857142859-8.604285714285714-2.299999999999997-6.238571428571428-6.237142857142857-2.3000000000000003-8.601428571428578 2.3000000000000003-8.605714285714287 6.238571428571428-6.237142857142856 8.604285714285714-2.3000000000000003z m2.8571428571428577 27.834285714285716v-4.242857142857144q0-0.3114285714285714-0.1999999999999993-0.5228571428571414t-0.49428571428571644-0.21142857142857352h-4.285714285714285q-0.28999999999999915 0-0.5142857142857125 0.2228571428571442t-0.22142857142857153 0.5142857142857125v4.240000000000002q0 0.28999999999999915 0.2228571428571442 0.5142857142857125t0.514285714285716 0.22142857142857153h4.285714285714285q0.28857142857143003 0 0.48999999999999844-0.2142857142857153t0.1999999999999993-0.5228571428571414z m-0.04285714285714448-7.678571428571431l0.3999999999999986-13.862857142857145q0-0.2671428571428578-0.2228571428571442-0.40000000000000036-0.2228571428571371-0.17857142857142705-0.5342857142857085-0.17857142857142705h-4.914285714285718q-0.31428571428571317 0-0.5357142857142847 0.17857142857142883-0.2228571428571442 0.13428571428571345-0.2228571428571442 0.40000000000000036l0.38142857142857167 13.862857142857143q0 0.22142857142857153 0.2228571428571442 0.3885714285714279t0.5357142857142847 0.16714285714285637h4.12857142857143q0.31428571428571317 0 0.5257142857142867-0.16857142857142904t0.23428571428571487-0.39000000000000057z' })
                )
            );
        }
    }]);

    return FaExclamationCircle;
}(React.Component);

exports.default = FaExclamationCircle;
module.exports = exports['default'];