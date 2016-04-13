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

var TiDelete = function (_React$Component) {
    _inherits(TiDelete, _React$Component);

    function TiDelete() {
        _classCallCheck(this, TiDelete);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiDelete).apply(this, arguments));
    }

    _createClass(TiDelete, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 6.666666666666667c-7.366666666666667 0-13.333333333333334 5.97-13.333333333333334 13.333333333333332s5.966666666666667 13.333333333333336 13.333333333333334 13.333333333333336 13.333333333333336-5.969999999999999 13.333333333333336-13.333333333333336-5.966666666666669-13.333333333333334-13.333333333333336-13.333333333333334z m6.178333333333335 17.155c0.6499999999999986 0.6499999999999986 0.6499999999999986 1.7049999999999983 0 2.3566666666666656-0.3249999999999993 0.3249999999999993-0.75 0.4883333333333333-1.1783333333333346 0.4883333333333333s-0.8533333333333317-0.163333333333334-1.1783333333333346-0.4883333333333333l-3.8216666666666654-3.821666666666669-3.8216666666666654 3.821666666666669c-0.3249999999999993 0.3249999999999993-0.75 0.4883333333333333-1.1783333333333328 0.4883333333333333s-0.8533333333333335-0.163333333333334-1.1783333333333328-0.4883333333333333c-0.6500000000000004-0.6499999999999986-0.6500000000000004-1.7049999999999983 0-2.3566666666666656l3.8216666666666654-3.821666666666669-3.821666666666669-3.8216666666666654c-0.6500000000000004-0.6500000000000004-0.6500000000000004-1.705 0-2.3566666666666656s1.705-0.6500000000000004 2.3566666666666656 0l3.821666666666669 3.8216666666666654 3.8216666666666654-3.821666666666667c0.6499999999999986-0.6500000000000004 1.7049999999999983-0.6500000000000004 2.3566666666666656 0s0.6499999999999986 1.705 0 2.3566666666666674l-3.8216666666666654 3.8216666666666654 3.821666666666669 3.8216666666666654z' })
                )
            );
        }
    }]);

    return TiDelete;
}(React.Component);

exports.default = TiDelete;
module.exports = exports['default'];