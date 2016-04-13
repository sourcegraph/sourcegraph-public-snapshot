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

var MdLabelOutline = function (_React$Component) {
    _inherits(MdLabelOutline, _React$Component);

    function MdLabelOutline() {
        _classCallCheck(this, MdLabelOutline);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdLabelOutline).apply(this, arguments));
    }

    _createClass(MdLabelOutline, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm26.64 28.36l5.938333333333333-8.36-5.938333333333333-8.360000000000001h-18.28333333333333v16.71666666666666h18.28333333333333z m2.7333333333333343-18.595000000000002l7.266666666666666 10.23333333333333-7.266666666666666 10.233333333333334q-1.0133333333333319 1.408333333333335-2.7333333333333343 1.408333333333335h-18.28333333333333q-1.3266666666666689 0-2.3416666666666686-0.9766666666666666t-1.0150000000000006-2.304999999999996v-16.715000000000003q0-1.3299999999999983 1.0166666666666666-2.306666666666665t2.3416666666666677-0.9766666666666666h18.283333333333335q1.716666666666665 0 2.7333333333333343 1.4066666666666663z' })
                )
            );
        }
    }]);

    return MdLabelOutline;
}(React.Component);

exports.default = MdLabelOutline;
module.exports = exports['default'];