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

var MdSwapCalls = function (_React$Component) {
    _inherits(MdSwapCalls, _React$Component);

    function MdSwapCalls() {
        _classCallCheck(this, MdSwapCalls);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdSwapCalls).apply(this, arguments));
    }

    _createClass(MdSwapCalls, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30 6.640000000000001l6.640000000000001 6.716666666666669h-5v11.64333333333333q0 2.7333333333333343-1.9533333333333331 4.686666666666667t-4.686666666666667 1.9533333333333367-4.690000000000001-1.9533333333333331-1.9533333333333331-4.686666666666671v-11.643333333333334q0-1.3283333333333331-1.0166666666666657-2.3433333333333337t-2.34-1.0133333333333319-2.3450000000000006 1.0166666666666657-1.0166666666666675 2.341666666666667v11.641666666666667h5l-6.638333333333332 6.636666666666667-6.64-6.636666666666667h4.999999999999999v-11.643333333333334q0-2.7333333333333343 1.9533333333333331-4.726666666666667t4.686666666666667-1.9916666666666654 4.690000000000001 1.9916666666666671 1.9533333333333331 4.725v11.645q0 1.326666666666668 1.0166666666666657 2.3416666666666686t2.341666666666665 1.0166666666666657 2.344999999999999-1.0166666666666657 1.0166666666666657-2.3433333333333337v-11.64166666666667h-5z' })
                )
            );
        }
    }]);

    return MdSwapCalls;
}(React.Component);

exports.default = MdSwapCalls;
module.exports = exports['default'];