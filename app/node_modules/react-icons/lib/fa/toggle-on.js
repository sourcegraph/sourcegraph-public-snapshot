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

var FaToggleOn = function (_React$Component) {
    _inherits(FaToggleOn, _React$Component);

    function FaToggleOn() {
        _classCallCheck(this, FaToggleOn);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaToggleOn).apply(this, arguments));
    }

    _createClass(FaToggleOn, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm0 20q0-2.5375000000000014 0.9962500000000001-4.85375t2.66625-3.9837500000000006 3.9837500000000006-2.66625 4.85375-0.9962499999999999h15q2.5375000000000014 0 4.853749999999998 0.9962499999999999t3.9837500000000006 2.66625 2.666249999999998 3.983749999999999 0.9962500000000034 4.853750000000002-0.9962500000000034 4.853749999999998-2.666249999999998 3.9837500000000006-3.9837500000000006 2.666249999999998-4.853749999999998 0.9962500000000034h-15q-2.5374999999999996 0-4.85375-0.9962499999999999t-3.98375-2.6662500000000016-2.66625-3.9837500000000006-0.9962500000000003-4.853749999999998z m27.5 10q2.03125 0 3.8775000000000013-0.7912500000000016t3.193749999999998-2.1374999999999993 2.137500000000003-3.1950000000000003 0.7924999999999969-3.875-0.7899999999999991-3.87875-2.137500000000003-3.1937499999999996-3.1950000000000003-2.137500000000001-3.8787499999999966-0.7949999999999982-3.8775000000000013 0.7912499999999998-3.1937500000000014 2.1400000000000006-2.1374999999999993 3.1937499999999996-0.7937499999999993 3.87875 0.7899999999999991 3.875 2.1374999999999993 3.1950000000000003 3.1950000000000003 2.1374999999999993 3.8800000000000026 0.7925000000000004z' })
                )
            );
        }
    }]);

    return FaToggleOn;
}(React.Component);

exports.default = FaToggleOn;
module.exports = exports['default'];