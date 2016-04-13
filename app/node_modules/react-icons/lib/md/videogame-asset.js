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

var MdVideogameAsset = function (_React$Component) {
    _inherits(MdVideogameAsset, _React$Component);

    function MdVideogameAsset() {
        _classCallCheck(this, MdVideogameAsset);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdVideogameAsset).apply(this, arguments));
    }

    _createClass(MdVideogameAsset, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm32.5 20q1.0933333333333337 0 1.7966666666666669-0.7033333333333331t0.7033333333333331-1.7966666666666669-0.7033333333333331-1.7966666666666669-1.7966666666666669-0.7033333333333331-1.7966666666666669 0.7033333333333331-0.7033333333333331 1.7966666666666669 0.7033333333333331 1.7966666666666669 1.7966666666666669 0.7033333333333331z m-6.640000000000001 5q1.0166666666666657 0 1.7583333333333329-0.7033333333333331t0.7433333333333323-1.7966666666666669-0.7416666666666671-1.7966666666666669-1.7583333333333293-0.7033333333333331-1.7583333333333329 0.7033333333333331-0.7399999999999984 1.7966666666666669 0.7416666666666671 1.7966666666666669 1.7600000000000016 0.7033333333333331z m-7.5-3.3599999999999994v-3.2833333333333314h-5v-5h-3.3599999999999994v5h-5v3.2833333333333314h5v5h3.3599999999999994v-5h5z m16.64-11.64q1.3283333333333331 0 2.3433333333333337 1.0166666666666657t1.0166666666666657 2.341666666666667v13.283333333333333q0 1.326666666666668-1.0166666666666657 2.3416666666666686t-2.3433333333333337 1.0166666666666657h-30q-1.3283333333333331 0-2.3433333333333333-1.0166666666666657t-1.0166666666666666-2.3433333333333337v-13.283333333333333q0-1.3266666666666662 1.0166666666666666-2.341666666666667t2.3433333333333333-1.0150000000000006h30z' })
                )
            );
        }
    }]);

    return MdVideogameAsset;
}(React.Component);

exports.default = MdVideogameAsset;
module.exports = exports['default'];