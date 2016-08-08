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

var FaParagraph = function (_React$Component) {
    _inherits(FaParagraph, _React$Component);

    function FaParagraph() {
        _classCallCheck(this, FaParagraph);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaParagraph).apply(this, arguments));
    }

    _createClass(FaParagraph, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm34.24285714285714 4.218571428571429v1.628571428571429q0 0.6485714285714286-0.41428571428571104 1.362857142857143t-0.9485714285714266 0.7142857142857144q-1.1142857142857139 0-1.2057142857142864 0.02285714285714313-0.5799999999999983 0.13428571428571434-0.7142857142857153 0.6914285714285713-0.0671428571428585 0.24571428571428555-0.0671428571428585 1.4285714285714288v25.71428571428572q0 0.5571428571428569-0.3999999999999986 0.9600000000000009t-0.96142857142857 0.3999999999999986h-2.41q-0.5571428571428569 0-0.9600000000000009-0.3999999999999986t-0.3999999999999986-0.96142857142857v-27.18428571428572h-3.192857142857143v27.18571428571429q0 0.5571428571428569-0.39000000000000057 0.9600000000000009t-0.9714285714285715 0.3999999999999986h-2.4100000000000037q-0.5800000000000018 0-0.9714285714285715-0.3999999999999986t-0.39000000000000057-0.96142857142857v-11.07142857142857q-3.281428571428572-0.2671428571428578-5.46857142857143-1.3171428571428585-2.814285714285715-1.2942857142857136-4.285714285714286-3.9957142857142856-1.4285714285714288-2.611428571428572-1.4285714285714288-5.78142857142857 0-3.7042857142857137 1.9642857142857144-6.382857142857143 1.9642857142857153-2.6342857142857143 4.6657142857142855-3.5485714285714285 2.4771428571428604-0.8257142857142963 9.308571428571435-0.8257142857142963h10.691428571428574q0.5571428571428569 0 0.9600000000000009 0.3999999999999999t0.3999999999999986 0.9614285714285717z' })
                )
            );
        }
    }]);

    return FaParagraph;
}(React.Component);

exports.default = FaParagraph;
module.exports = exports['default'];