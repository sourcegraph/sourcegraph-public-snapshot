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

var MdExitToApp = function (_React$Component) {
    _inherits(MdExitToApp, _React$Component);

    function MdExitToApp() {
        _classCallCheck(this, MdExitToApp);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdExitToApp).apply(this, arguments));
    }

    _createClass(MdExitToApp, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.640000000000004 5q1.3283333333333367 0 2.34333333333333 1.0166666666666666t1.0166666666666657 2.3400000000000007v23.28333333333333q0 1.326666666666668-1.0166666666666657 2.3416666666666686t-2.3433333333333337 1.0166666666666657h-23.28333333333333q-1.405000000000002 0-2.3833333333333346-1.0166666666666657t-0.9733333333333345-2.341666666666665v-6.640000000000004h3.3583333333333343v6.640000000000001h23.283333333333335v-23.28333333333333h-23.285000000000004v6.643333333333331h-3.3566666666666656v-6.643333333333333q0-1.3283333333333331 0.9749999999999996-2.3433333333333337t2.383333333333333-1.0133333333333336h23.28333333333333z m-14.843333333333337 21.016666666666666l4.296666666666667-4.376666666666665h-16.093333333333334v-3.2833333333333314h16.093333333333334l-4.296666666666667-4.373333333333335 2.3433333333333337-2.3433333333333337 8.36 8.36-8.36 8.36z' })
                )
            );
        }
    }]);

    return MdExitToApp;
}(React.Component);

exports.default = MdExitToApp;
module.exports = exports['default'];