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

var FaQrcode = function (_React$Component) {
    _inherits(FaQrcode, _React$Component);

    function FaQrcode() {
        _classCallCheck(this, FaQrcode);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaQrcode).apply(this, arguments));
    }

    _createClass(FaQrcode, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm12.857142857142858 25.714285714285715v2.8571428571428577h-2.8571428571428577v-2.8571428571428577h2.8571428571428577z m0-17.142857142857146v2.8571428571428594h-2.8571428571428577v-2.8571428571428577h2.8571428571428577z m17.142857142857142 1.7763568394002505e-15v2.8571428571428577h-2.8571428571428577v-2.8571428571428577h2.8571428571428577z m-22.857142857142858 22.83428571428572h8.571428571428573v-8.548571428571432h-8.571428571428573v8.548571428571428z m8.881784197001252e-16-17.120000000000005h8.571428571428573v-8.57142857142857h-8.571428571428573v8.57142857142857z m17.142857142857142 1.7763568394002505e-15h8.57142857142857v-8.571428571428573h-8.57142857142857v8.571428571428571z m-5.714285714285715 5.7142857142857135v14.285714285714285h-14.285714285714285v-14.285714285714285h14.285714285714288z m11.42857142857143 11.42857142857143v2.857142857142854h-2.8571428571428577v-2.8571428571428577h2.8571428571428577z m5.714285714285715 0v2.857142857142854h-2.857142857142854v-2.8571428571428577h2.857142857142854z m0-11.42857142857143v8.57142857142857h-8.57142857142857v-2.8571428571428577h-2.8571428571428577v8.571428571428573h-2.8571428571428577v-14.285714285714285h8.57142857142857v2.8571428571428577h2.857142857142854v-2.8571428571428577h2.857142857142854z m-17.142857142857142-17.142857142857142v14.285714285714285h-14.285714285714288v-14.285714285714285h14.285714285714288z m17.142857142857142-4.440892098500626e-16v14.285714285714285h-14.285714285714285v-14.285714285714285h14.285714285714285z' })
                )
            );
        }
    }]);

    return FaQrcode;
}(React.Component);

exports.default = FaQrcode;
module.exports = exports['default'];