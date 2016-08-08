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

var MdNotificationsActive = function (_React$Component) {
    _inherits(MdNotificationsActive, _React$Component);

    function MdNotificationsActive() {
        _classCallCheck(this, MdNotificationsActive);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdNotificationsActive).apply(this, arguments));
    }

    _createClass(MdNotificationsActive, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 36.64000000000001q-1.4066666666666663 0-2.383333333333333-0.9766666666666666t-0.9766666666666666-2.3049999999999997h6.640000000000001q0 1.4066666666666663-0.9383333333333326 2.3433333333333337t-2.3416666666666686 0.9399999999999906z m10-18.28v8.283333333333331l3.3599999999999994 3.356666666666662v1.6416666666666657h-26.716666666666665v-1.6416666666666657l3.3566666666666656-3.3566666666666656v-8.283333333333335q0-3.9049999999999994 1.995000000000001-6.793333333333331t5.508333333333333-3.75v-1.1716666666666669q0-1.0166666666666666 0.7049999999999983-1.7583333333333329t1.7916666666666679-0.7450000000000019 1.8000000000000007 0.7416666666666663 0.6999999999999993 1.7583333333333329v1.1716666666666669q3.5166666666666657 0.8599999999999994 5.510000000000002 3.75t1.9899999999999984 6.796666666666667z m3.2833333333333314-0.8599999999999994q-0.3916666666666657-6.716666666666669-5.861666666666665-10.703333333333333l2.3433333333333337-2.3433333333333337q6.483333333333334 5 6.873333333333335 13.046666666666667h-3.3599999999999994z m-20.62833333333333-10.70333333333334q-5.545000000000001 3.9066666666666663-5.9383333333333335 10.703333333333333h-3.3566666666666674q0.3899999999999997-8.046666666666667 6.873333333333332-13.046666666666667z' })
                )
            );
        }
    }]);

    return MdNotificationsActive;
}(React.Component);

exports.default = MdNotificationsActive;
module.exports = exports['default'];