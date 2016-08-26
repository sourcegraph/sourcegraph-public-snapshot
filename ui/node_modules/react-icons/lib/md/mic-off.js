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

var MdMicOff = function (_React$Component) {
    _inherits(MdMicOff, _React$Component);

    function MdMicOff() {
        _classCallCheck(this, MdMicOff);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdMicOff).apply(this, arguments));
    }

    _createClass(MdMicOff, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm7.1 5l27.9 27.9-2.1 2.1-7-7c-1.2 0.8-2.8 1.3-4.3 1.5v5.5h-3.2v-5.5c-5.5-0.8-10-5.4-10-11.1h2.8c0 5 4.2 8.4 8.8 8.4 1.3 0 2.7-0.3 3.8-0.9l-2.7-2.7c-0.3 0.1-0.7 0.2-1.1 0.2-2.7 0-5-2.3-5-5v-1.3l-10-10z m17.9 13.6l-10-9.9v-0.3c0-2.8 2.3-5 5-5s5 2.2 5 5v10.2z m6.6-0.2c0 1.9-0.5 3.8-1.4 5.4l-2.1-2.1c0.5-1 0.7-2.1 0.7-3.3h2.8z' })
                )
            );
        }
    }]);

    return MdMicOff;
}(React.Component);

exports.default = MdMicOff;
module.exports = exports['default'];