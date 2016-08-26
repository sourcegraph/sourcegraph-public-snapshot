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

var MdCall = function (_React$Component) {
    _inherits(MdCall, _React$Component);

    function MdCall() {
        _classCallCheck(this, MdCall);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdCall).apply(this, arguments));
    }

    _createClass(MdCall, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm11 18q3.8 7.2 11 11l3.7-3.7q0.8-0.8 1.7-0.4 2.8 1 6 1 0.7 0 1.1 0.4t0.5 1.2v5.9q0 0.7-0.5 1.1t-1.1 0.5q-11.8 0-20.1-8.3t-8.3-20.1q0-0.7 0.5-1.1t1.1-0.5h5.9q0.7 0 1.2 0.5t0.4 1.1q0 3.2 1 6 0.3 1-0.4 1.7z' })
                )
            );
        }
    }]);

    return MdCall;
}(React.Component);

exports.default = MdCall;
module.exports = exports['default'];