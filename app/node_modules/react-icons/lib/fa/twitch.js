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

var FaTwitch = function (_React$Component) {
    _inherits(FaTwitch, _React$Component);

    function FaTwitch() {
        _classCallCheck(this, FaTwitch);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaTwitch).apply(this, arguments));
    }

    _createClass(FaTwitch, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 9.685714285714287v9.688571428571429h-3.2371428571428567v-9.685714285714287h3.2371428571428567z m8.885714285714286 0v9.688571428571429h-3.2385714285714293v-9.685714285714287h3.2371428571428567z m0 16.965714285714284l5.645714285714284-5.671428571428571v-17.74142857142857h-26.65142857142857v23.414285714285715h7.277142857142859v4.842857142857142l4.842857142857142-4.842857142857142h8.885714285714286z m8.880000000000003-26.65142857142857v22.611428571428572l-9.685714285714287 9.685714285714287h-7.280000000000001l-4.84 4.845714285714287h-4.845714285714285v-4.842857142857142h-8.885714285714286v-25.851428571428574l2.435714285714285-6.44857142857143h33.10285714285715z' })
                )
            );
        }
    }]);

    return FaTwitch;
}(React.Component);

exports.default = FaTwitch;
module.exports = exports['default'];