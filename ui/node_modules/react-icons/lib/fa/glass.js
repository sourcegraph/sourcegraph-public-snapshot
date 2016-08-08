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

var FaGlass = function (_React$Component) {
    _inherits(FaGlass, _React$Component);

    function FaGlass() {
        _classCallCheck(this, FaGlass);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaGlass).apply(this, arguments));
    }

    _createClass(FaGlass, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm37.924285714285716 4.151428571428572q0 0.7814285714285711-0.9600000000000009 1.7428571428571429l-14.107142857142858 14.105714285714285v17.14285714285714h7.142857142857142q0.5799999999999983 0 1.0042857142857144 0.42428571428571615t0.42428571428571615 1.0042857142857144-0.4242857142857126 1.0042857142857144-1.004285714285718 0.42428571428571615h-20q-0.5800000000000001 0-1.0042857142857144-0.42428571428571615t-0.4242857142857144-1.0042857142857144 0.4242857142857144-1.0042857142857144 1.0042857142857144-0.42428571428570905h7.142857142857142v-17.142857142857146l-14.107142857142858-14.107142857142858q-0.9599999999999991-0.9599999999999991-0.9599999999999991-1.742857142857142 0-0.5114285714285716 0.3999999999999999-0.8142857142857141t0.8500000000000001-0.3885714285714288 0.96-0.0900000000000003h31.42857142857143q0.5142857142857125 0 0.9600000000000009 0.08999999999999986t0.8485714285714252 0.3900000000000001 0.3999999999999986 0.8142857142857141z' })
                )
            );
        }
    }]);

    return FaGlass;
}(React.Component);

exports.default = FaGlass;
module.exports = exports['default'];