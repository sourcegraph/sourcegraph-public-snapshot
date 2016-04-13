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

var MdFormatUnderlined = function (_React$Component) {
    _inherits(MdFormatUnderlined, _React$Component);

    function MdFormatUnderlined() {
        _classCallCheck(this, MdFormatUnderlined);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdFormatUnderlined).apply(this, arguments));
    }

    _createClass(MdFormatUnderlined, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm8.360000000000001 31.640000000000004h23.28333333333334v3.359999999999996h-23.285000000000004v-3.3599999999999994z m11.639999999999999-3.2800000000000047q-4.140000000000001 0-7.07-2.9299999999999997t-2.9299999999999997-7.07v-13.36h4.140000000000001v13.36q0 2.421666666666667 1.7166666666666668 4.100000000000001t4.143333333333333 1.6849999999999987 4.138333333333335-1.6833333333333336 1.7166666666666686-4.100000000000001v-13.361666666666665h4.144999999999996v13.36q0 4.140000000000001-2.9299999999999997 7.07t-7.07 2.9299999999999997z' })
                )
            );
        }
    }]);

    return MdFormatUnderlined;
}(React.Component);

exports.default = MdFormatUnderlined;
module.exports = exports['default'];