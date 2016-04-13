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

var MdSettingsBluetooth = function (_React$Component) {
    _inherits(MdSettingsBluetooth, _React$Component);

    function MdSettingsBluetooth() {
        _classCallCheck(this, MdSettingsBluetooth);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdSettingsBluetooth).apply(this, arguments));
    }

    _createClass(MdSettingsBluetooth, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm24.766666666666666 23.828333333333337l-3.126666666666665-3.125v6.25z m-3.1283333333333303-17.42166666666667v6.25l3.1283333333333303-3.123333333333333z m7.894999999999996 3.126666666666667l-7.190000000000001 7.1083333333333325 7.188333333333333 7.188333333333333-9.533333333333335 9.533333333333331h-1.6400000000000006v-12.658333333333331l-7.656666666666666 7.656666666666666-2.3433333333333337-2.3433333333333337 9.296666666666667-9.373333333333335-9.296666666666667-9.296666666666667 2.3433333333333337-2.3433333333333337 7.656666666666666 7.656666666666666v-12.658333333333328h1.6416666666666728z m-4.533333333333331 30.46666666666667v-3.3599999999999994h3.3599999999999994v3.3599999999999994h-3.3599999999999994z m-13.360000000000001 0v-3.3599999999999994h3.360000000000001v3.3599999999999994h-3.3599999999999994z m6.720000000000001 0v-3.3599999999999994h3.2833333333333314v3.3599999999999994h-3.2833333333333314z' })
                )
            );
        }
    }]);

    return MdSettingsBluetooth;
}(React.Component);

exports.default = MdSettingsBluetooth;
module.exports = exports['default'];