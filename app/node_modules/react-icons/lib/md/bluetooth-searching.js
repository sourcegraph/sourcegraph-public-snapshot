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

var MdBluetoothSearching = function (_React$Component) {
    _inherits(MdBluetoothSearching, _React$Component);

    function MdBluetoothSearching() {
        _classCallCheck(this, MdBluetoothSearching);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdBluetoothSearching).apply(this, arguments));
    }

    _createClass(MdBluetoothSearching, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm21.483333333333334 27.188333333333333l-3.123333333333335-3.203333333333333v6.329999999999998z m-3.123333333333335-17.5v6.328333333333333l3.125-3.203333333333333z m7.811666666666667 3.125l-7.188333333333333 7.186666666666667 7.188333333333333 7.190000000000001-9.533333333333331 9.45333333333333h-1.6383333333333354v-12.656666666666666l-7.656666666666667 7.656666666666666-2.343333333333333-2.3433333333333337 9.296666666666667-9.296666666666667-9.296666666666667-9.296666666666663 2.3433333333333337-2.3433333333333337 7.656666666666666 7.656666666666666v-12.656666666666666h1.6400000000000006z m6.406666666666666-1.6416666666666657q2.421666666666667 3.906666666666668 2.421666666666667 8.671666666666667t-2.578333333333333 8.828333333333333l-1.9549999999999983-1.9549999999999983q1.6416666666666657-3.280000000000001 1.6416666666666657-6.716666666666669t-1.6400000000000006-6.716666666666669z m-8.828333333333333 8.828333333333333l3.828333333333333-3.828333333333333q0.783333333333335 1.9533333333333331 0.783333333333335 3.828333333333333 0 1.9533333333333331-0.783333333333335 3.9066666666666663z' })
                )
            );
        }
    }]);

    return MdBluetoothSearching;
}(React.Component);

exports.default = MdBluetoothSearching;
module.exports = exports['default'];