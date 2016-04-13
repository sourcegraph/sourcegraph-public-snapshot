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

var MdNoEncryption = function (_React$Component) {
    _inherits(MdNoEncryption, _React$Component);

    function MdNoEncryption() {
        _classCallCheck(this, MdNoEncryption);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdNoEncryption).apply(this, arguments));
    }

    _createClass(MdNoEncryption, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm14.843333333333335 10v2.033333333333333l-3.0466666666666686-3.0500000000000007q0.39000000000000057-3.123333333333333 2.6950000000000003-5.233333333333333t5.508333333333333-2.109999999999999q3.4383333333333326 0 5.899999999999999 2.4616666666666664t2.4583333333333357 5.898333333333333v3.3599999999999994h1.6416666666666657q1.3283333333333331 0 2.3433333333333337 0.9766666666666666t1.0166666666666657 2.3049999999999997v13.906666666666666l-17.268333333333334-17.189999999999998h9.063333333333333v-3.3583333333333343q0-2.1100000000000003-1.5233333333333334-3.6333333333333337t-3.631666666666664-1.5233333333333325-3.633333333333333 1.5233333333333334-1.5233333333333317 3.633333333333333z m20.156666666666666 26.328333333333333l-2.0333333333333314 2.0333333333333314-1.8733333333333348-1.8766666666666652q-0.625 0.15833333333333144-1.0933333333333337 0.15833333333333144h-20q-1.3283333333333331 0-2.3433333333333337-0.9766666666666666t-1.0166666666666666-2.306666666666665v-16.716666666666665q0-1.955 1.7999999999999998-2.8916666666666675l-3.4399999999999995-3.3583333333333343 2.033333333333333-2.033333333333333z' })
                )
            );
        }
    }]);

    return MdNoEncryption;
}(React.Component);

exports.default = MdNoEncryption;
module.exports = exports['default'];