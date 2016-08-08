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

var MdSettingsInputSvideo = function (_React$Component) {
    _inherits(MdSettingsInputSvideo, _React$Component);

    function MdSettingsInputSvideo() {
        _classCallCheck(this, MdSettingsInputSvideo);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdSettingsInputSvideo).apply(this, arguments));
    }

    _createClass(MdSettingsInputSvideo, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm25.86 25q1.0166666666666657 0 1.7583333333333329 0.7033333333333331t0.7433333333333323 1.7966666666666669-0.7416666666666671 1.7966666666666669-1.7566666666666677 0.7033333333333331-1.7583333333333329-0.7033333333333331-0.7399999999999984-1.7966666666666669 0.7416666666666671-1.7966666666666669 1.7550000000000026-0.7033333333333331z m3.280000000000001-8.36q1.0166666666666657 0 1.7583333333333329 0.7416666666666671t0.7399999999999984 1.7566666666666677-0.7416666666666671 1.7583333333333329-1.7600000000000016 0.7399999999999984-1.7583333333333329-0.7416666666666671-0.7433333333333323-1.7600000000000016 0.7416666666666671-1.7583333333333329 1.7566666666666677-0.7433333333333323z m-9.14 18.36q6.25 0 10.625-4.375t4.375-10.625-4.375-10.625-10.625-4.375-10.625 4.375-4.375 10.625 4.375 10.625 10.625 4.375z m0-33.36q7.578333333333333-4.440892098500626e-16 12.966666666666669 5.390000000000001t5.391666666666666 12.966666666666669-5.390000000000001 12.969999999999999-12.968333333333334 5.390000000000001-12.966666666666667-5.390000000000001-5.393333333333333-12.966666666666669 5.393333333333333-12.971666666666668 12.966666666666667-5.388333333333332z m-5.859999999999999 23.36q1.0166666666666657 0 1.7583333333333329 0.7033333333333331t0.7399999999999984 1.7966666666666669-0.7416666666666671 1.7966666666666669-1.7583333333333293 0.7033333333333331-1.7583333333333329-0.7033333333333331-0.7433333333333341-1.7966666666666669 0.7416666666666671-1.7966666666666669 1.7599999999999998-0.7033333333333331z m10.86-14.139999999999999q0 1.0166666666666675-0.7033333333333331 1.7583333333333329t-1.7966666666666669 0.7433333333333341h-5q-1.0933333333333337 0-1.7966666666666669-0.7416666666666671t-0.7033333333333331-1.7583333333333329 0.7033333333333331-1.7583333333333329 1.7966666666666669-0.7400000000000002h5q1.0933333333333337 0 1.7966666666666669 0.7416666666666671t0.7033333333333331 1.756666666666666z m-11.64 8.28q0 1.0166666666666657-0.7416666666666671 1.7583333333333329t-1.7566666666666677 0.7399999999999984-1.7583333333333329-0.7416666666666671-0.7400000000000002-1.7600000000000016 0.7416666666666671-1.7583333333333329 1.7599999999999998-0.7433333333333323 1.7583333333333329 0.7416666666666671 0.7433333333333341 1.7566666666666677z' })
                )
            );
        }
    }]);

    return MdSettingsInputSvideo;
}(React.Component);

exports.default = MdSettingsInputSvideo;
module.exports = exports['default'];