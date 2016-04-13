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

var FaBluetoothB = function (_React$Component) {
    _inherits(FaBluetoothB, _React$Component);

    function FaBluetoothB() {
        _classCallCheck(this, FaBluetoothB);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaBluetoothB).apply(this, arguments));
    }

    _createClass(FaBluetoothB, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm21.874285714285715 31.762857142857147l3.861428571428572-3.84-3.861428571428572-3.84v7.678571428571427z m0-15.847142857142858l3.861428571428572-3.838571428571429-3.861428571428572-3.84v7.678571428571429z m0.7142857142857153 4.085714285714287l7.9471428571428575 7.9471428571428575-12.031428571428574 12.051428571428566v-15.87142857142857l-6.628571428571428 6.60857142857143-2.4128571428571437-2.4085714285714346 8.304285714285715-8.328571428571426-8.302857142857142-8.322857142857144 2.41-2.41 6.62857142857143 6.607142857142858v-15.87l12.032857142857143 12.052857142857142z' })
                )
            );
        }
    }]);

    return FaBluetoothB;
}(React.Component);

exports.default = FaBluetoothB;
module.exports = exports['default'];