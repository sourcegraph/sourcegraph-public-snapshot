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

var MdBluetoothDisabled = function (_React$Component) {
    _inherits(MdBluetoothDisabled, _React$Component);

    function MdBluetoothDisabled() {
        _classCallCheck(this, MdBluetoothDisabled);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdBluetoothDisabled).apply(this, arguments));
    }

    _createClass(MdBluetoothDisabled, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm21.64 30.313333333333333l3.125-3.125-3.126666666666665-3.203333333333333v6.329999999999998z m-12.656666666666668-23.673333333333332l24.37833333333333 24.376666666666665-2.3433333333333337 2.3416666666666686-3.828333333333333-3.828333333333333-7.189999999999998 7.1100000000000065h-1.6400000000000006v-12.656666666666673l-7.656666666666666 7.656666666666666-2.3433333333333337-2.3433333333333337 9.296666666666667-9.296666666666667-11.016666666666667-11.016666666666667z m12.656666666666668 3.0500000000000007v5.388333333333334l-3.2833333333333314-3.360000000000001v-8.358333333333334h1.6433333333333309l9.530000000000001 9.453333333333333-5.078333333333333 5.078333333333333-2.3416666666666686-2.3433333333333337 2.6566666666666663-2.7333333333333343z' })
                )
            );
        }
    }]);

    return MdBluetoothDisabled;
}(React.Component);

exports.default = MdBluetoothDisabled;
module.exports = exports['default'];