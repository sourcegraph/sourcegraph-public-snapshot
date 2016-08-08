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

var MdLocalParking = function (_React$Component) {
    _inherits(MdLocalParking, _React$Component);

    function MdLocalParking() {
        _classCallCheck(this, MdLocalParking);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdLocalParking).apply(this, arguments));
    }

    _createClass(MdLocalParking, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm22.033333333333335 18.36q1.326666666666668 0 2.3033333333333346-1.0166666666666657t0.9766666666666666-2.341666666666667-0.9766666666666666-2.341666666666667-2.3049999999999997-1.0166666666666657h-5.391666666666669v6.720000000000002h5.390000000000001z m-0.3949999999999996-13.36q4.140000000000001 0 7.07 2.9299999999999997t2.9316666666666684 7.07-2.9299999999999997 7.07-7.070000000000004 2.9299999999999997h-5v10h-6.640000000000001v-30h11.64z' })
                )
            );
        }
    }]);

    return MdLocalParking;
}(React.Component);

exports.default = MdLocalParking;
module.exports = exports['default'];