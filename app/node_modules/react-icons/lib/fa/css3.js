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

var FaCss3 = function (_React$Component) {
    _inherits(FaCss3, _React$Component);

    function FaCss3() {
        _classCallCheck(this, FaCss3);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaCss3).apply(this, arguments));
    }

    _createClass(FaCss3, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm6.138571428571428 2.857142857142857h33.59428571428572l-5.9371428571428595 29.75428571428571-17.947142857142854 5.960000000000001-15.58-5.960000000000001 1.5857142857142859-7.9471428571428575h6.628571428571428l-0.6471428571428568 3.2814285714285703 9.419999999999998 3.5942857142857143 10.847142857142856-3.5942857142857143 1.5171428571428578-7.567142857142855h-26.96285714285714l1.2942857142857136-6.628571428571428h26.985714285714288l0.8500000000000014-4.264285714285714h-26.962857142857146z' })
                )
            );
        }
    }]);

    return FaCss3;
}(React.Component);

exports.default = FaCss3;
module.exports = exports['default'];