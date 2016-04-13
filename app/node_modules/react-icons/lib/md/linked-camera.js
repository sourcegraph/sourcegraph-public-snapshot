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

var MdLinkedCamera = function (_React$Component) {
    _inherits(MdLinkedCamera, _React$Component);

    function MdLinkedCamera() {
        _classCallCheck(this, MdLinkedCamera);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdLinkedCamera).apply(this, arguments));
    }

    _createClass(MdLinkedCamera, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm26.64 10q1.4066666666666663 0 2.383333333333333 0.9766666666666666t0.9766666666666666 2.3833333333333346h2.1883333333333326q0-2.2666666666666675-1.6400000000000006-3.9066666666666663t-3.908333333333335-1.6400000000000006v2.1866666666666656z m-6.640000000000001 21.640000000000004q3.4383333333333326 0 5.899999999999999-2.421666666666667t2.460000000000001-5.858333333333334-2.461666666666666-5.900000000000002-5.898333333333333-2.460000000000001-5.9 2.460000000000001-2.459999999999999 5.899999999999999 2.460000000000001 5.856666666666669 5.899999999999999 2.424999999999997z m8.36-16.640000000000004h8.283333333333331v18.36q0 1.3283333333333331-0.9783333333333317 2.3049999999999997t-2.306666666666665 0.9750000000000014h-26.715q-1.330000000000001 0-2.3066666666666675-0.9766666666666666t-0.9766666666666666-2.306666666666665v-20q0-1.3283333333333331 0.9766666666666666-2.3433333333333337t2.3049999999999997-1.0133333333333354h5.3133333333333335l3.045-3.3566666666666665h10v5.000000000000001q1.408333333333335 0 2.383333333333333 0.9766666666666666t0.980000000000004 2.379999999999999z m-1.7199999999999989-9.453333333333333v-2.188333333333333q4.140000000000001 0 7.07 2.93t2.9299999999999997 7.07h-2.1899999999999977q0-3.203333333333333-2.3033333333333346-5.508333333333334t-5.508333333333333-2.3050000000000006z m-11.950000000000001 17.813333333333336q0-2.2666666666666657 1.5616666666666656-3.828333333333333t3.75-1.5666666666666664 3.75 1.5666666666666664 1.5666666666666664 3.826666666666668-1.5666666666666664 3.788333333333334-3.75 1.5216666666666683-3.75-1.5233333333333334-1.5616666666666674-3.789999999999999z' })
                )
            );
        }
    }]);

    return MdLinkedCamera;
}(React.Component);

exports.default = MdLinkedCamera;
module.exports = exports['default'];