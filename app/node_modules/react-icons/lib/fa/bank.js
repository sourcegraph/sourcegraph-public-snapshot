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

var FaBank = function (_React$Component) {
    _inherits(FaBank, _React$Component);

    function FaBank() {
        _classCallCheck(this, FaBank);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaBank).apply(this, arguments));
    }

    _createClass(FaBank, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 1.3333333333333333l20 8v2.666666666666666h-2.6666666666666643q0 0.5413333333333341-0.4266666666666694 0.9373333333333331t-1.0106666666666655 0.395999999999999h-31.792q-0.5840000000000001 0-1.0106666666666664-0.3960000000000008t-0.4266666666666663-0.9373333333333314h-2.6666666666666665v-2.666666666666668z m-14.666666666666668 13.333333333333332h5.333333333333334v15.999999999999998h2.666666666666666v-15.999999999999998h5.333333333333332v15.999999999999998h2.666666666666668v-15.999999999999998h5.333333333333332v15.999999999999998h2.666666666666668v-15.999999999999998h5.333333333333332v15.999999999999998h1.2293333333333365q0.5840000000000032 0 1.0106666666666655 0.3960000000000008t0.4266666666666623 0.9373333333333349v1.3333333333333357h-34.666666666666664v-1.3333333333333357q2.220446049250313e-15-0.5413333333333341 0.42666666666666897-0.9373333333333349t1.010666666666666-0.3960000000000008h1.2293333333333338v-15.999999999999998z m33.22933333333333 20q0.5840000000000103 0 1.0106666666666726 0.3960000000000008t0.4266666666666694 0.9373333333333349v2.6666666666666643h-40v-2.6666666666666643q0-0.5413333333333341 0.42666666666666664-0.9373333333333349t1.0106666666666668-0.3960000000000008h37.12533333333333z' })
                )
            );
        }
    }]);

    return FaBank;
}(React.Component);

exports.default = FaBank;
module.exports = exports['default'];