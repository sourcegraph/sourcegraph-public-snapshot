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

var MdLabel = function (_React$Component) {
    _inherits(MdLabel, _React$Component);

    function MdLabel() {
        _classCallCheck(this, MdLabel);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdLabel).apply(this, arguments));
    }

    _createClass(MdLabel, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm29.375 9.766666666666667l7.266666666666666 10.233333333333333-7.266666666666666 10.233333333333334q-1.0166666666666657 1.408333333333335-2.7333333333333343 1.408333333333335h-18.284999999999997q-1.328333333333335 0-2.3433333333333355-0.9766666666666666t-1.0133333333333336-2.3066666666666684v-16.715000000000003q0-1.3299999999999983 1.0166666666666666-2.306666666666665t2.3416666666666677-0.9766666666666666h18.283333333333335q1.716666666666665 0 2.7333333333333343 1.4066666666666663z' })
                )
            );
        }
    }]);

    return MdLabel;
}(React.Component);

exports.default = MdLabel;
module.exports = exports['default'];