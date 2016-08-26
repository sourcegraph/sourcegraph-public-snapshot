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

var MdPhoneMissed = function (_React$Component) {
    _inherits(MdPhoneMissed, _React$Component);

    function MdPhoneMissed() {
        _classCallCheck(this, MdPhoneMissed);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdPhoneMissed).apply(this, arguments));
    }

    _createClass(MdPhoneMissed, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm39.5 27.8c0.3 0.3 0.5 0.7 0.5 1.2s-0.2 0.8-0.5 1.2l-4.1 4.1c-0.3 0.3-0.7 0.5-1.2 0.5s-0.8-0.2-1.2-0.5c-1.3-1.3-2.8-2.3-4.4-3.1-0.6-0.3-0.9-0.9-0.9-1.5v-5.2c-2.5-0.8-5-1.1-7.7-1.1s-5.2 0.3-7.7 1.1v5.2c0 0.7-0.3 1.3-0.9 1.6-1.6 0.7-3.1 1.7-4.4 3-0.4 0.3-0.7 0.5-1.2 0.5s-0.9-0.2-1.2-0.5l-4.1-4.1c-0.3-0.4-0.5-0.7-0.5-1.2s0.2-0.9 0.5-1.2c5.1-4.8 11.9-7.8 19.5-7.8s14.5 3 19.5 7.8z m-28.6-18.7v5.9h-2.5v-10h10v2.5h-5.9l7.5 7.5 10-10 1.6 1.6-11.6 11.8z' })
                )
            );
        }
    }]);

    return MdPhoneMissed;
}(React.Component);

exports.default = MdPhoneMissed;
module.exports = exports['default'];