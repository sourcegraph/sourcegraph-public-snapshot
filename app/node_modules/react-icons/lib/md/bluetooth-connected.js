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

var MdBluetoothConnected = function (_React$Component) {
    _inherits(MdBluetoothConnected, _React$Component);

    function MdBluetoothConnected() {
        _classCallCheck(this, MdBluetoothConnected);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdBluetoothConnected).apply(this, arguments));
    }

    _createClass(MdBluetoothConnected, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.640000000000004 16.64l3.359999999999996 3.3599999999999994-3.3599999999999994 3.3599999999999994-3.283333333333335-3.3599999999999994z m-6.875 10.55l-3.126666666666665-3.205000000000002v6.329999999999998z m-3.126666666666665-17.5v6.326666666666664l3.125-3.203333333333333z m7.891666666666669 3.123333333333333l-7.188333333333343 7.186666666666666 7.188333333333333 7.190000000000001-9.529999999999998 9.453333333333337h-1.6400000000000006v-12.656666666666666l-7.656666666666666 7.656666666666666-2.3433333333333337-2.3433333333333337 9.296666666666667-9.296666666666667-9.296666666666667-9.296666666666667 2.3433333333333337-2.3433333333333337 7.656666666666666 7.656666666666666v-12.65666666666667h1.6400000000000006z m-17.891666666666673 7.186666666666666l-3.283333333333335 3.3599999999999994-3.3550000000000004-3.3599999999999994 3.3583333333333343-3.3599999999999994z' })
                )
            );
        }
    }]);

    return MdBluetoothConnected;
}(React.Component);

exports.default = MdBluetoothConnected;
module.exports = exports['default'];