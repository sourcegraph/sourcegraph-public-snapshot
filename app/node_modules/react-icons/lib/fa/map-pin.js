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

var FaMapPin = function (_React$Component) {
    _inherits(FaMapPin, _React$Component);

    function FaMapPin() {
        _classCallCheck(this, FaMapPin);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaMapPin).apply(this, arguments));
    }

    _createClass(FaMapPin, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 24.285714285714285q1.4714285714285715 0 2.8571428571428577-0.33428571428571274v14.619999999999997q0 0.5799999999999983-0.4242857142857126 1.0042857142857144t-1.0042857142857144 0.42428571428571615h-2.8571428571428577q-0.5800000000000018 0-1.0042857142857144-0.42428571428571615t-0.42428571428571615-1.0042857142857144v-14.620000000000001q1.361428571428572 0.3342857142857163 2.8571428571428577 0.3342857142857163z m0-24.285714285714285q4.732857142857142 0 8.079999999999998 3.3485714285714283t3.3485714285714323 8.08-3.3485714285714288 8.08-8.080000000000002 3.3485714285714288-8.08-3.3485714285714288-3.3485714285714288-8.08 3.3485714285714288-8.08 8.08-3.3485714285714288z m0 5q0.31428571428571317 0 0.5142857142857125-0.20000000000000018t0.20000000000000284-0.5142857142857142-0.1999999999999993-0.5142857142857142-0.514285714285716-0.19999999999999973q-3.2571428571428562 0-5.557142857142857 2.3000000000000003t-2.3000000000000007 5.557142857142857q0 0.31428571428571495 0.1999999999999993 0.5142857142857142t0.514285714285716 0.1999999999999993 0.5142857142857142-0.1999999999999993 0.1999999999999993-0.5142857142857142q0-2.6571428571428566 1.8857142857142861-4.542857142857144t4.542857142857143-1.8857142857142852z' })
                )
            );
        }
    }]);

    return FaMapPin;
}(React.Component);

exports.default = FaMapPin;
module.exports = exports['default'];