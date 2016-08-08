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

var FaWindows = function (_React$Component) {
    _inherits(FaWindows, _React$Component);

    function FaWindows() {
        _classCallCheck(this, FaWindows);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaWindows).apply(this, arguments));
    }

    _createClass(FaWindows, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm16.65142857142857 22.45714285714286v14.528571428571428l-15.222857142857142-2.095714285714287v-12.432857142857141h15.222857142857142z m0-16.585714285714285v14.708571428571428h-15.222857142857142v-12.611428571428572z m21.919999999999998 16.585714285714285v17.54285714285714l-20.245714285714286-2.789999999999999v-14.752857142857142h20.245714285714286z m0-19.6v17.722857142857144h-20.245714285714286v-14.932857142857145z' })
                )
            );
        }
    }]);

    return FaWindows;
}(React.Component);

exports.default = FaWindows;
module.exports = exports['default'];