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

var FaHtml5 = function (_React$Component) {
    _inherits(FaHtml5, _React$Component);

    function FaHtml5() {
        _classCallCheck(this, FaHtml5);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaHtml5).apply(this, arguments));
    }

    _createClass(FaHtml5, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm29.50857142857143 13.325714285714286l0.35714285714285765-3.9057142857142857h-19.732857142857142l1.0485714285714263 11.92h13.660000000000002l-0.49285714285714377 5.09-4.397142857142857 1.1857142857142868-4.375714285714286-1.1857142857142868-0.2900000000000009-3.1242857142857154h-3.902857142857142l0.4914285714285711 6.205714285714286 8.081428571428573 2.232857142857142h0.08999999999999986v-0.022857142857141355l8.014285714285716-2.210000000000001 1.1142857142857139-12.142857142857142h-14.378571428571432l-0.33428571428571097-4.039999999999999h15.042857142857141z m-25.222857142857144-10.468571428571428h31.42857142857143l-2.857142857142854 32.099999999999994-12.900000000000002 3.614285714285714-12.814285714285717-3.614285714285714z' })
                )
            );
        }
    }]);

    return FaHtml5;
}(React.Component);

exports.default = FaHtml5;
module.exports = exports['default'];