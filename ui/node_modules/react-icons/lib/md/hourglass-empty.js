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

var MdHourglassEmpty = function (_React$Component) {
    _inherits(MdHourglassEmpty, _React$Component);

    function MdHourglassEmpty() {
        _classCallCheck(this, MdHourglassEmpty);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdHourglassEmpty).apply(this, arguments));
    }

    _createClass(MdHourglassEmpty, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 19.14l6.640000000000001-6.640000000000001v-5.859999999999999h-13.283333333333333v5.859999999999999z m6.640000000000001 8.36l-6.640000000000001-6.640000000000001-6.639999999999999 6.640000000000001v5.859999999999999h13.283333333333333v-5.859999999999999z m-16.64-24.14h20v10l-6.640000000000001 6.640000000000001 6.640000000000001 6.640000000000001v10h-20v-10l6.640000000000001-6.640000000000001-6.640000000000001-6.639999999999999v-10z' })
                )
            );
        }
    }]);

    return MdHourglassEmpty;
}(React.Component);

exports.default = MdHourglassEmpty;
module.exports = exports['default'];