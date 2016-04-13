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

var TiMediaFastForward = function (_React$Component) {
    _inherits(TiMediaFastForward, _React$Component);

    function TiMediaFastForward() {
        _classCallCheck(this, TiMediaFastForward);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiMediaFastForward).apply(this, arguments));
    }

    _createClass(TiMediaFastForward, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm25.088333333333335 11.52c-0.5399999999999991-0.5233333333333334-1.2749999999999986-0.8533333333333335-2.0883333333333347-0.8533333333333335-1.6566666666666663 0-3 1.3450000000000006-3 3v14.999999999999998c0 1.658333333333335 1.3433333333333337 3 3 3 0.8133333333333326 0 1.5500000000000007-0.3249999999999993 2.0883333333333347-0.8533333333333317 3.966666666666665-3.8566666666666656 9.911666666666665-9.646666666666665 9.911666666666665-9.646666666666665s-5.943333333333335-5.788333333333334-9.911666666666669-9.646666666666668z m-15.000000000000002 0c-0.5383333333333322-0.5233333333333334-1.2749999999999986-0.8533333333333317-2.088333333333333-0.8533333333333317-1.6566666666666663 0-3 1.3450000000000006-3 3v15c0 1.658333333333335 1.3433333333333337 3 3 3 0.8133333333333326 0 1.5500000000000007-0.3249999999999993 2.088333333333333-0.8533333333333317 3.966666666666667-3.856666666666669 9.911666666666667-9.646666666666668 9.911666666666667-9.646666666666668s-5.943333333333333-5.786666666666667-9.911666666666667-9.645z' })
                )
            );
        }
    }]);

    return TiMediaFastForward;
}(React.Component);

exports.default = TiMediaFastForward;
module.exports = exports['default'];