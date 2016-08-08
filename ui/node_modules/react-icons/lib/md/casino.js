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

var MdCasino = function (_React$Component) {
    _inherits(MdCasino, _React$Component);

    function MdCasino() {
        _classCallCheck(this, MdCasino);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdCasino).apply(this, arguments));
    }

    _createClass(MdCasino, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm27.5 15q1.0933333333333337 0 1.7966666666666669-0.7033333333333331t0.7033333333333331-1.7966666666666669-0.7033333333333331-1.7966666666666669-1.7966666666666669-0.7033333333333331-1.7966666666666669 0.7033333333333331-0.7033333333333331 1.7966666666666669 0.7033333333333331 1.7966666666666669 1.7966666666666669 0.7033333333333331z m0 15q1.0933333333333337 0 1.7966666666666669-0.7033333333333331t0.7033333333333331-1.7966666666666669-0.7033333333333331-1.7966666666666669-1.7966666666666669-0.7033333333333331-1.7966666666666669 0.7033333333333331-0.7033333333333331 1.7966666666666669 0.7033333333333331 1.7966666666666669 1.7966666666666669 0.7033333333333331z m-7.5-7.5q1.0933333333333337 0 1.7966666666666669-0.7033333333333331t0.7033333333333331-1.7966666666666669-0.7033333333333331-1.7966666666666669-1.7966666666666669-0.7033333333333331-1.7966666666666669 0.7033333333333331-0.7033333333333331 1.7966666666666669 0.7033333333333331 1.7966666666666669 1.7966666666666669 0.7033333333333331z m-7.5-7.5q1.0933333333333337 0 1.7966666666666669-0.7033333333333331t0.7033333333333331-1.7966666666666669-0.7033333333333331-1.7966666666666669-1.7966666666666669-0.7033333333333331-1.7966666666666669 0.7033333333333331-0.7033333333333331 1.7966666666666669 0.7033333333333331 1.7966666666666669 1.7966666666666669 0.7033333333333331z m0 15q1.0933333333333337 0 1.7966666666666669-0.7033333333333331t0.7033333333333331-1.7966666666666669-0.7033333333333331-1.7966666666666669-1.7966666666666669-0.7033333333333331-1.7966666666666669 0.7033333333333331-0.7033333333333331 1.7966666666666669 0.7033333333333331 1.7966666666666669 1.7966666666666669 0.7033333333333331z m19.140000000000004-25q1.3283333333333367 0 2.34333333333333 1.0166666666666666t1.0166666666666657 2.3400000000000007v23.28333333333333q0 1.326666666666668-1.0166666666666657 2.3416666666666686t-2.3433333333333337 1.0166666666666657h-23.28333333333333q-1.3266666666666689 0-2.3416666666666686-1.0166666666666657t-1.0150000000000006-2.341666666666665v-23.28333333333334q0-1.3266666666666653 1.0166666666666666-2.341666666666665t2.3400000000000007-1.0150000000000006h23.28333333333333z' })
                )
            );
        }
    }]);

    return MdCasino;
}(React.Component);

exports.default = MdCasino;
module.exports = exports['default'];