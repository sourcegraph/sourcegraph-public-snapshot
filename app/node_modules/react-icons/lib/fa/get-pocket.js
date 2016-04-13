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

var FaGetPocket = function (_React$Component) {
    _inherits(FaGetPocket, _React$Component);

    function FaGetPocket() {
        _classCallCheck(this, FaGetPocket);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaGetPocket).apply(this, arguments));
    }

    _createClass(FaGetPocket, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm35.64714285714286 2.857142857142857q1.451428571428572 0 2.4571428571428555 1.0142857142857142t1.0028571428571453 2.4685714285714284v11.585714285714285q0 3.9285714285714306-1.5171428571428578 7.5t-4.074285714285715 6.137142857142859-6.114285714285714 4.074285714285711-7.468571428571433 1.5057142857142907q-3.928571428571427 0-7.488571428571429-1.5057142857142836t-6.125714285714286-4.074285714285715-4.085714285714285-6.138571428571428-1.5142857142857145-7.5v-11.585714285714285q0-1.4285714285714288 1.0257142857142858-2.454285714285714t2.455714285714288-1.027142857142863h31.450000000000006z m-15.714285714285715 23.75q1.048571428571428 0 1.8285714285714292-0.7371428571428567l9.018571428571427-8.66142857142857q0.8257142857142838-0.7814285714285703 0.8257142857142838-1.8971428571428568 0-1.0957142857142852-0.7714285714285722-1.8657142857142865t-1.8628571428571448-0.7714285714285722q-1.048571428571428 0-1.8285714285714292 0.7385714285714293l-7.21142857142857 6.921428571428571-7.209999999999997-6.920000000000002q-0.781428571428572-0.7357142857142858-1.8085714285714296-0.7357142857142858-1.0957142857142852 0-1.8657142857142865 0.7714285714285722t-0.7714285714285722 1.862857142857143q0 1.1371428571428588 0.8042857142857152 1.8957142857142841l9.04 8.66q0.7371428571428567 0.7357142857142875 1.8085714285714296 0.7357142857142875z' })
                )
            );
        }
    }]);

    return FaGetPocket;
}(React.Component);

exports.default = FaGetPocket;
module.exports = exports['default'];