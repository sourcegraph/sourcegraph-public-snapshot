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

var FaViacoin = function (_React$Component) {
    _inherits(FaViacoin, _React$Component);

    function FaViacoin() {
        _classCallCheck(this, FaViacoin);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaViacoin).apply(this, arguments));
    }

    _createClass(FaViacoin, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm37.142857142857146 0l-4.285714285714285 10h4.285714285714285v4.285714285714285h-6.114285714285714l-1.2285714285714278 2.8571428571428577h7.342857142857142v4.285714285714285h-9.17142857142857l-7.971428571428575 18.571428571428573-7.965714285714286-18.57142857142857h-9.177142857142856v-4.285714285714285h7.345714285714285l-1.2285714285714295-2.8571428571428577h-6.112857142857141v-4.285714285714288h4.285714285714285l-4.285714285714286-10h5.714285714285714l7.210000000000001 17.142857142857142h8.435714285714289l7.210000000000001-17.142857142857142h5.714285714285715z m-17.142857142857146 27.142857142857142l2.41-5.714285714285715h-4.821428571428573z' })
                )
            );
        }
    }]);

    return FaViacoin;
}(React.Component);

exports.default = FaViacoin;
module.exports = exports['default'];