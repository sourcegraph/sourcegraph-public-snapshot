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

var TiMediaFastForwardOutline = function (_React$Component) {
    _inherits(TiMediaFastForwardOutline, _React$Component);

    function TiMediaFastForwardOutline() {
        _classCallCheck(this, TiMediaFastForwardOutline);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiMediaFastForwardOutline).apply(this, arguments));
    }

    _createClass(TiMediaFastForwardOutline, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm23.333333333333336 14.46l6.888333333333335 6.708333333333336-6.888333333333335 6.711666666666666v-13.416666666666668z m-0.33333333333333215-3.793333333333333c-1.6566666666666663 0-3 1.3450000000000006-3 3v15c0 1.658333333333335 1.3433333333333337 3 3 3 0.8133333333333326 0 1.5500000000000007-0.3249999999999993 2.0883333333333347-0.8533333333333317 3.9666666666666615-3.856666666666669 9.911666666666662-9.646666666666668 9.911666666666662-9.646666666666668s-5.943333333333335-5.790000000000001-9.906666666666666-9.646666666666668c-0.5450000000000017-0.5233333333333334-1.2800000000000011-0.8533333333333335-2.0933333333333337-0.8533333333333335z m-14.66666666666667 3.793333333333333l6.888333333333332 6.706666666666667-6.888333333333332 6.713333333333335v-13.416666666666668z m-0.3333333333333339-3.793333333333333c-1.6566666666666663 0-3 1.3450000000000006-3 3v15c0 1.658333333333335 1.3433333333333337 3 3 3 0.8133333333333326 0 1.5500000000000007-0.3249999999999993 2.088333333333333-0.8533333333333317 3.966666666666667-3.856666666666669 9.911666666666667-9.646666666666668 9.911666666666667-9.646666666666668l-9.906666666666666-9.645c-0.543333333333333-0.5250000000000021-1.2799999999999994-0.8550000000000004-2.0933333333333337-0.8550000000000004z' })
                )
            );
        }
    }]);

    return TiMediaFastForwardOutline;
}(React.Component);

exports.default = TiMediaFastForwardOutline;
module.exports = exports['default'];