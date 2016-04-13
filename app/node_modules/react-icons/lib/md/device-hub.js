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

var MdDeviceHub = function (_React$Component) {
    _inherits(MdDeviceHub, _React$Component);

    function MdDeviceHub() {
        _classCallCheck(this, MdDeviceHub);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdDeviceHub).apply(this, arguments));
    }

    _createClass(MdDeviceHub, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.36 26.64h6.640000000000001v8.36h-8.36v-5.078333333333333l-6.640000000000001-7.033333333333331-6.640000000000001 7.033333333333331v5.078333333333333h-8.36v-8.36h6.640000000000001l6.716666666666665-6.640000000000001v-5.313333333333334q-1.4833333333333343-0.5466666666666669-2.42-1.7966666666666669t-0.9366666666666656-2.889999999999999q0-2.033333333333333 1.4833333333333343-3.5166666666666666t3.5166666666666657-1.4833333333333334 3.5133333333333354 1.4833333333333334 1.4833333333333343 3.5166666666666666q0 1.6400000000000006-0.9366666666666674 2.8900000000000006t-2.421666666666667 1.7966666666666669v5.313333333333333z' })
                )
            );
        }
    }]);

    return MdDeviceHub;
}(React.Component);

exports.default = MdDeviceHub;
module.exports = exports['default'];