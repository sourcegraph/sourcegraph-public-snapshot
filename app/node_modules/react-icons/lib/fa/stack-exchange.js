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

var FaStackExchange = function (_React$Component) {
    _inherits(FaStackExchange, _React$Component);

    function FaStackExchange() {
        _classCallCheck(this, FaStackExchange);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaStackExchange).apply(this, arguments));
    }

    _createClass(FaStackExchange, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm33.81714285714286 27.96857142857143v1.4714285714285715q0 1.8999999999999986-1.2828571428571465 3.2285714285714313t-3.0914285714285725 1.3285714285714292h-1.2714285714285722l-5.805714285714284 6.002857142857138v-6.0042857142857144h-11.808571428571428q-1.807142857142857 0-3.09-1.3285714285714292t-1.281428571428573-3.2242857142857133v-1.4742857142857133h27.632857142857144z m0-7.277142857142856v5.69142857142857h-27.634285714285717v-5.69142857142857h27.634285714285717z m0-7.321428571428571v5.6899999999999995h-27.634285714285717v-5.688571428571432h27.634285714285717z m0-3.1257142857142863v1.4985714285714273h-27.634285714285717v-1.5q-1.7763568394002505e-15-1.8757142857142863 1.2828571428571411-3.202857142857143t3.0914285714285725-1.3257142857142865h18.885714285714286q1.807142857142857 0 3.0899999999999963 1.3285714285714283t1.2828571428571394 3.202857142857143z' })
                )
            );
        }
    }]);

    return FaStackExchange;
}(React.Component);

exports.default = FaStackExchange;
module.exports = exports['default'];