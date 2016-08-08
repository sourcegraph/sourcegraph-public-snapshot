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

var MdDesktopMac = function (_React$Component) {
    _inherits(MdDesktopMac, _React$Component);

    function MdDesktopMac() {
        _classCallCheck(this, MdDesktopMac);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdDesktopMac).apply(this, arguments));
    }

    _createClass(MdDesktopMac, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm35 23.36v-16.716666666666665h-30v16.716666666666665h30z m0-20q1.3283333333333331 0 2.3433333333333337 0.9766666666666666t1.0166666666666657 2.3049999999999997v20q0 1.3283333333333331-1.0166666666666657 2.3433333333333337t-2.3433333333333337 1.0150000000000006h-11.64l3.2833333333333314 5v1.6400000000000006h-13.283333333333333v-1.6400000000000006l3.283333333333333-5h-11.64333333333333q-1.326666666666667 0-2.341666666666667-1.0166666666666657t-1.0166666666666666-2.3416666666666686v-20q0-1.3283333333333331 1.0166666666666666-2.3049999999999997t2.341666666666667-0.9766666666666657h30z' })
                )
            );
        }
    }]);

    return MdDesktopMac;
}(React.Component);

exports.default = MdDesktopMac;
module.exports = exports['default'];