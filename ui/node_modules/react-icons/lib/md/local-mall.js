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

var MdLocalMall = function (_React$Component) {
    _inherits(MdLocalMall, _React$Component);

    function MdLocalMall() {
        _classCallCheck(this, MdLocalMall);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdLocalMall).apply(this, arguments));
    }

    _createClass(MdLocalMall, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 21.64q3.4383333333333326 0 5.899999999999999-2.421666666666667t2.460000000000001-5.858333333333334h-3.3599999999999994q0 2.033333333333333-1.4833333333333343 3.5166666666666657t-3.5166666666666657 1.4833333333333343-3.5166666666666657-1.4833333333333343-1.4833333333333343-3.5199999999999996h-3.3599999999999994q0 3.4383333333333326 2.461666666666666 5.859999999999999t5.898333333333333 2.4250000000000007z m0-16.64q-2.0333333333333314 0-3.5166666666666657 1.4833333333333334t-1.4833333333333343 3.5166666666666666h10q0-2.033333333333333-1.4833333333333343-3.5166666666666666t-3.5166666666666657-1.4833333333333334z m11.64 5q1.3283333333333331 0 2.3433333333333337 1.0166666666666657t1.0166666666666657 2.34v20.000000000000004q0 1.3283333333333331-1.0166666666666657 2.3049999999999997t-2.3433333333333337 0.9766666666666666h-23.28333333333333q-1.3266666666666689 0-2.3416666666666686-0.9766666666666666t-1.0150000000000006-2.3033333333333346v-20q0-1.3283333333333331 1.0166666666666666-2.3433333333333337t2.3400000000000007-1.0150000000000006h3.283333333333333q0-3.4383333333333344 2.460000000000001-5.9t5.899999999999999-2.458333333333333 5.899999999999999 2.461666666666667 2.4583333333333357 5.8966666666666665h3.2833333333333314z' })
                )
            );
        }
    }]);

    return MdLocalMall;
}(React.Component);

exports.default = MdLocalMall;
module.exports = exports['default'];