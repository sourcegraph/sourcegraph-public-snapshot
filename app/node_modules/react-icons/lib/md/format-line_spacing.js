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

var MdFormatLineSpacing = function (_React$Component) {
    _inherits(MdFormatLineSpacing, _React$Component);

    function MdFormatLineSpacing() {
        _classCallCheck(this, MdFormatLineSpacing);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdFormatLineSpacing).apply(this, arguments));
    }

    _createClass(MdFormatLineSpacing, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm16.64 21.64v-3.2833333333333314h20v3.2833333333333314h-20z m0 10v-3.2833333333333314h20v3.2833333333333314h-20z m0-23.28h20v3.283333333333333h-20v-3.283333333333333z m-6.640000000000001 3.280000000000001v16.716666666666665h4.140000000000001l-5.783333333333335 5.783333333333335-5.858333333333333-5.783333333333335h4.138333333333334v-16.714999999999996h-4.136666666666666l5.858333333333334-5.783333333333336 5.783333333333335 5.783333333333334h-4.141666666666669z' })
                )
            );
        }
    }]);

    return MdFormatLineSpacing;
}(React.Component);

exports.default = MdFormatLineSpacing;
module.exports = exports['default'];