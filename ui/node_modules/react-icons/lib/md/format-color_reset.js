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

var MdFormatColorReset = function (_React$Component) {
    _inherits(MdFormatColorReset, _React$Component);

    function MdFormatColorReset() {
        _classCallCheck(this, MdFormatColorReset);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdFormatColorReset).apply(this, arguments));
    }

    _createClass(MdFormatColorReset, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm8.75 8.75l24.375 24.375-2.1099999999999994 2.1099999999999994-4.454999999999998-4.373333333333335q-2.8900000000000006 2.5-6.563333333333333 2.5-4.141666666666666 0-7.071666666666667-2.9299999999999997t-2.9250000000000025-7.071666666666665q0-2.6566666666666627 2.1833333333333336-6.876666666666665l-5.543333333333333-5.543333333333333z m21.25 14.61q0 1.1716666666666669-0.23333333333333428 2.1883333333333326l-14.299999999999999-14.376666666666665 4.533333333333333-5.858333333333333q1.0933333333333337 1.25 2.7733333333333334 3.3600000000000003t4.453333333333333 6.913333333333332 2.7733333333333334 7.773333333333333z' })
                )
            );
        }
    }]);

    return MdFormatColorReset;
}(React.Component);

exports.default = MdFormatColorReset;
module.exports = exports['default'];