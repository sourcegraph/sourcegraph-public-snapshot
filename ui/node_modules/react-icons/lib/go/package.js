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

var GoPackage = function (_React$Component) {
    _inherits(GoPackage, _React$Component);

    function GoPackage() {
        _classCallCheck(this, GoPackage);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(GoPackage).apply(this, arguments));
    }

    _createClass(GoPackage, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm18.75 2.5l-18.75 5v22.5l18.75 5 18.75-5v-22.5l-18.75-5z m-16.255 25.6625l-0.015000000000001013-16.9125 15.02 4.005000000000001v16.9075l-15.004999999999999-4z m-0.015000000000001013-19.4125l6.26-1.6687500000000002 16.259999999999998 4.34125-6.25 1.6662499999999998-16.27-4.338749999999999z m32.525 19.412499999999998l-15.004999999999995 4.0000000000000036v-16.9075l5-1.333750000000002v6.09375l5-1.3337499999999984v-6.093750000000002l5.019999999999996-1.3375000000000004-0.015000000000000568 16.91z m-5.0049999999999955-18.074999999999996v-0.0037500000000019185l-16.259999999999998-4.3375 5.009999999999998-1.3337500000000002 16.269999999999996 4.3375-5.019999999999996 1.3375000000000004z' })
                )
            );
        }
    }]);

    return GoPackage;
}(React.Component);

exports.default = GoPackage;
module.exports = exports['default'];