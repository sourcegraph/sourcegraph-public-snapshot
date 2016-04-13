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

var MdLocalSee = function (_React$Component) {
    _inherits(MdLocalSee, _React$Component);

    function MdLocalSee() {
        _classCallCheck(this, MdLocalSee);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdLocalSee).apply(this, arguments));
    }

    _createClass(MdLocalSee, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 28.36q3.4383333333333326 0 5.899999999999999-2.461666666666666t2.4583333333333357-5.898333333333333-2.4583333333333357-5.899999999999999-5.899999999999999-2.458333333333334-5.899999999999999 2.458333333333334-2.458333333333334 5.899999999999999 2.461666666666668 5.899999999999999 5.896666666666665 2.4583333333333357z m-5-25h10l3.046666666666667 3.283333333333333h5.313333333333333q1.3283333333333331 0 2.3049999999999997 1.0133333333333336t0.9750000000000085 2.3433333333333337v20q0 1.3283333333333331-0.9766666666666666 2.3433333333333337t-2.306666666666665 1.0166666666666657h-26.713333333333342q-1.330000000000001 0-2.3066666666666675-1.0166666666666657t-0.9766666666666666-2.3433333333333337v-20q0-1.3283333333333331 0.9766666666666666-2.3433333333333337t2.3050000000000006-1.0166666666666666h5.313333333333334z m-0.31166666666666565 16.64q0-2.1883333333333326 1.5633333333333344-3.75t3.7483333333333313-1.5633333333333344 3.75 1.5633333333333344 1.5666666666666664 3.75-1.5666666666666664 3.75-3.75 1.5633333333333326-3.75-1.5633333333333326-1.5599999999999987-3.75z' })
                )
            );
        }
    }]);

    return MdLocalSee;
}(React.Component);

exports.default = MdLocalSee;
module.exports = exports['default'];