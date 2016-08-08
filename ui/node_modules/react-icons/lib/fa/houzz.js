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

var FaHouzz = function (_React$Component) {
    _inherits(FaHouzz, _React$Component);

    function FaHouzz() {
        _classCallCheck(this, FaHouzz);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaHouzz).apply(this, arguments));
    }

    _createClass(FaHouzz, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 26.585714285714285l11.42857142857143-6.585714285714285v13.191428571428574l-11.42857142857143 6.608571428571423v-13.214285714285715z m-11.428571428571429-6.585714285714285v13.191428571428574l11.428571428571429-6.607142857142858z m11.428571428571429-19.8v13.192857142857143l-11.428571428571429 6.607142857142858v-13.19142857142857z m0 13.192857142857143l11.42857142857143-6.585714285714286v13.192857142857143z' })
                )
            );
        }
    }]);

    return FaHouzz;
}(React.Component);

exports.default = FaHouzz;
module.exports = exports['default'];