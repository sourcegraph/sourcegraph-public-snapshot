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

var FaGenderless = function (_React$Component) {
    _inherits(FaGenderless, _React$Component);

    function FaGenderless() {
        _classCallCheck(this, FaGenderless);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaGenderless).apply(this, arguments));
    }

    _createClass(FaGenderless, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30 21.42857142857143q0-4.12857142857143-2.935714285714287-7.064285714285715t-7.064285714285713-2.935714285714287-7.064285714285715 2.935714285714285-2.935714285714285 7.064285714285717 2.935714285714287 7.064285714285717 7.064285714285713 2.9357142857142833 7.064285714285717-2.935714285714287 2.9357142857142833-7.064285714285713z m2.857142857142854 0q0 2.611428571428572-1.0142857142857125 4.988571428571429t-2.747142857142858 4.107142857142858-4.107142857142858 2.7457142857142856-4.988571428571426 1.0157142857142816-4.988571428571429-1.0142857142857125-4.107142857142858-2.747142857142858-2.7471428571428564-4.107142857142854-1.0142857142857133-4.988571428571429 1.0142857142857151-4.988571428571429 2.747142857142858-4.107142857142858 4.107142857142854-2.747142857142858 4.988571428571429-1.0142857142857142 4.988571428571429 1.0142857142857142 4.107142857142858 2.747142857142858 2.7457142857142856 4.107142857142858 1.0157142857142887 4.988571428571429z' })
                )
            );
        }
    }]);

    return FaGenderless;
}(React.Component);

exports.default = FaGenderless;
module.exports = exports['default'];