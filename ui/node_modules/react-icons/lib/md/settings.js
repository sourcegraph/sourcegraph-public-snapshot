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

var MdSettings = function (_React$Component) {
    _inherits(MdSettings, _React$Component);

    function MdSettings() {
        _classCallCheck(this, MdSettings);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdSettings).apply(this, arguments));
    }

    _createClass(MdSettings, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 25.9c3.2 0 5.9-2.7 5.9-5.9s-2.7-5.9-5.9-5.9-5.9 2.7-5.9 5.9 2.7 5.9 5.9 5.9z m12.4-4.3l3.5 2.8c0.4 0.2 0.4 0.7 0.2 1.1l-3.4 5.8c-0.2 0.3-0.6 0.4-1 0.3l-4.1-1.7c-0.9 0.7-1.8 1.3-2.8 1.7l-0.7 4.3c0 0.4-0.4 0.7-0.7 0.7h-6.8c-0.4 0-0.7-0.3-0.7-0.7l-0.7-4.3c-1-0.4-1.9-1-2.8-1.7l-4.1 1.7c-0.4 0.1-0.8 0-1-0.3l-3.4-5.8c-0.2-0.4-0.2-0.9 0.2-1.1l3.5-2.8c-0.1-0.5-0.1-1.1-0.1-1.6s0-1.1 0.1-1.6l-3.5-2.8c-0.4-0.2-0.4-0.7-0.2-1.1l3.4-5.7c0.2-0.4 0.6-0.5 1-0.4l4.1 1.7c0.9-0.6 1.8-1.3 2.8-1.7l0.7-4.3c0-0.4 0.3-0.7 0.7-0.7h6.8c0.3 0 0.7 0.3 0.7 0.7l0.7 4.3c1 0.4 1.9 1 2.8 1.7l4.1-1.7c0.4-0.1 0.8 0 1 0.4l3.4 5.7c0.2 0.4 0.2 0.9-0.2 1.1l-3.5 2.8c0.1 0.5 0.1 1.1 0.1 1.6s0 1.1-0.1 1.6z' })
                )
            );
        }
    }]);

    return MdSettings;
}(React.Component);

exports.default = MdSettings;
module.exports = exports['default'];