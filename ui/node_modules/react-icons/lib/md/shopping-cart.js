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

var MdShoppingCart = function (_React$Component) {
    _inherits(MdShoppingCart, _React$Component);

    function MdShoppingCart() {
        _classCallCheck(this, MdShoppingCart);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdShoppingCart).apply(this, arguments));
    }

    _createClass(MdShoppingCart, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.36 30q1.3283333333333331 0 2.3049999999999997 1.0166666666666657t0.9750000000000014 2.3416666666666686-0.9766666666666666 2.3049999999999997-2.306666666666665 0.9766666666666666-2.3433333333333337-0.9766666666666666-1.0166666666666657-2.3049999999999997 1.0166666666666657-2.3433333333333337 2.3433333333333337-1.0166666666666657z m-26.72-26.64h5.466666666666667l1.5666666666666664 3.283333333333334h24.686666666666667q0.7033333333333331 0 1.1716666666666669 0.5066666666666668t0.46666666666666856 1.211666666666666q0 0.39000000000000057-0.23333333333333428 0.7833333333333332l-5.940000000000001 10.780000000000001q-0.9383333333333326 1.716666666666665-2.8900000000000006 1.716666666666665h-12.418333333333333l-1.4866666666666664 2.736666666666668-0.07833333333333314 0.23333333333333428q0 0.3916666666666657 0.3916666666666675 0.3916666666666657h19.296666666666667v3.3599999999999994h-20q-1.3283333333333331 0-2.3049999999999997-1.0166666666666657t-0.9766666666666666-2.341666666666665q0-0.783333333333335 0.39000000000000057-1.5633333333333326l2.2666666666666675-4.140000000000001-6.016666666666667-12.655000000000001h-3.361666666666669v-3.283333333333333z m10 26.64q1.3283333333333331 0 2.3433333333333337 1.0166666666666657t1.0166666666666657 2.3400000000000034-1.0166666666666657 2.3049999999999997-2.3433333333333337 0.9766666666666666-2.3049999999999997-0.9766666666666666-0.9749999999999996-2.3049999999999997 0.9766666666666666-2.3433333333333337 2.3049999999999997-1.0133333333333354z' })
                )
            );
        }
    }]);

    return MdShoppingCart;
}(React.Component);

exports.default = MdShoppingCart;
module.exports = exports['default'];