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

var MdCropRotate = function (_React$Component) {
    _inherits(MdCropRotate, _React$Component);

    function MdCropRotate() {
        _classCallCheck(this, MdCropRotate);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdCropRotate).apply(this, arguments));
    }

    _createClass(MdCropRotate, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm13.360000000000001 26.64h20v3.3599999999999994h-3.3599999999999994v3.3599999999999994h-3.3599999999999994v-3.3599999999999994h-13.283333333333333q-1.4049999999999994 0-2.383333333333333-1.0166666666666657t-0.9733333333333345-2.34v-13.283333333333333h-3.3633333333333324v-3.360000000000001h3.3633333333333324v-3.3599999999999994h3.3583333333333343v20z m13.28-3.280000000000001v-10h-10v-3.3599999999999994h10q1.3283333333333331 0 2.3433333333333337 0.9766666666666666t1.0166666666666657 2.383333333333333v10h-3.361666666666668z m-6.561666666666664-23.36q7.813333333333333 0 13.555000000000003 5.3133333333333335t6.36666666666666 13.046666666666667h-2.5q-0.46666666666666856-4.688333333333333-3.125-8.438333333333333t-6.796666666666663-5.705l-2.2666666666666657 2.1900000000000004-6.325000000000003-6.328333333333334q0.23666666666666814 1.942890293094024e-16 0.5500000000000007-0.03833333333333314t0.5399999999999991-0.04z m-7.65666666666667 35.78333333333333l2.2666666666666675-2.1899999999999977 6.326666666666666 6.328333333333333q-0.23666666666666814 0-0.5500000000000007 0.038333333333333997t-0.5450000000000017 0.038333333333333997q-7.813333333333333 0-13.555-5.316666666666663t-6.364999999999998-13.041666666666671h2.5q0.4666666666666668 4.688333333333333 3.125 8.438333333333333t6.796666666666667 5.703333333333333z' })
                )
            );
        }
    }]);

    return MdCropRotate;
}(React.Component);

exports.default = MdCropRotate;
module.exports = exports['default'];