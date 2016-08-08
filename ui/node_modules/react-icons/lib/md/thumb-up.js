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

var MdThumbUp = function (_React$Component) {
    _inherits(MdThumbUp, _React$Component);

    function MdThumbUp() {
        _classCallCheck(this, MdThumbUp);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdThumbUp).apply(this, arguments));
    }

    _createClass(MdThumbUp, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm38.36 16.64l-0.07833333333333314 0.1566666666666663h0.07666666666666799v3.203333333333333q0 0.625-0.23333333333333428 1.25l-5.079999999999998 11.716666666666669q-0.7816666666666663 2.0333333333333314-3.0450000000000017 2.0333333333333314h-15q-1.3283333333333331 0-2.3433333333333337-1.0166666666666657t-1.0166666666666657-2.3416666666666686v-16.641666666666666q0-1.3266666666666662 1.0166666666666657-2.341666666666667l10.938333333333333-11.016666666666667 1.7966666666666669 1.8000000000000003q0.7033333333333331 0.6999999999999997 0.7033333333333331 1.7166666666666672v0.5466666666666669l-1.6400000000000006 7.656666666666668h10.545000000000002q1.3299999999999983 0 2.344999999999999 0.9766666666666666t1.0166666666666657 2.3049999999999997z m-36.72 18.36v-20h6.716666666666667v20h-6.715000000000001z' })
                )
            );
        }
    }]);

    return MdThumbUp;
}(React.Component);

exports.default = MdThumbUp;
module.exports = exports['default'];