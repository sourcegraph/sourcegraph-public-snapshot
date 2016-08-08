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

var FaMaxcdn = function (_React$Component) {
    _inherits(FaMaxcdn, _React$Component);

    function FaMaxcdn() {
        _classCallCheck(this, FaMaxcdn);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaMaxcdn).apply(this, arguments));
    }

    _createClass(FaMaxcdn, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm39.66571428571429 17.254285714285714l-3.6599999999999966 17.03142857142857h-7.4571428571428555l3.9742857142857133-18.571428571428573q0.28999999999999915-1.25-0.33428571428571274-1.9642857142857153-0.6028571428571432-0.7371428571428567-1.8528571428571432-0.7371428571428567h-3.7714285714285722l-4.555714285714291 21.272857142857145h-7.457142857142857l4.554285714285713-21.271428571428572h-6.385714285714286l-4.548571428571428 21.271428571428572h-7.457142857142856l4.5528571428571425-21.271428571428572-3.4142857142857137-7.299999999999998h28.481428571428573q2.254285714285711 0 4.228571428571431 0.9042857142857139t3.2942857142857136 2.532857142857143q1.3400000000000034 1.6285714285714281 1.808571428571426 3.7614285714285707t0 4.34z' })
                )
            );
        }
    }]);

    return FaMaxcdn;
}(React.Component);

exports.default = FaMaxcdn;
module.exports = exports['default'];