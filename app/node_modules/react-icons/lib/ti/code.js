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

var TiCode = function (_React$Component) {
    _inherits(TiCode, _React$Component);

    function TiCode() {
        _classCallCheck(this, TiCode);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiCode).apply(this, arguments));
    }

    _createClass(TiCode, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm13.616666666666667 30c-0.8499999999999996 0-1.7033333333333331-0.3249999999999993-2.3550000000000004-0.9766666666666666l-7.3566666666666665-7.356666666666666 7.3566666666666665-7.356666666666666c1.3000000000000007-1.3000000000000007 3.416666666666666-1.3000000000000007 4.713333333333333 0 1.299999999999999 1.3000000000000007 1.299999999999999 3.411666666666669 0 4.713333333333331l-2.6416666666666657 2.6433333333333344 2.6416666666666675 2.6433333333333344c1.3000000000000007 1.3000000000000007 1.3000000000000007 3.411666666666669 0 4.713333333333331-0.6500000000000004 0.6499999999999986-1.5033333333333339 0.9766666666666666-2.3566666666666656 0.9766666666666666z m12.766666666666666 0c-0.8550000000000004 0-1.7083333333333321-0.3249999999999993-2.3583333333333343-0.9766666666666666-1.3000000000000007-1.3000000000000007-1.3000000000000007-3.411666666666669 0-4.713333333333331l2.6416666666666693-2.6433333333333344-2.6416666666666657-2.6433333333333344c-1.3000000000000007-1.3000000000000007-1.3000000000000007-3.411666666666667 0-4.713333333333333 1.3000000000000007-1.3000000000000007 3.4116666666666653-1.3000000000000007 4.713333333333335 0l7.356666666666662 7.356666666666667-7.356666666666666 7.356666666666666c-0.6499999999999986 0.6499999999999986-1.5033333333333339 0.9766666666666666-2.3566666666666656 0.9766666666666666z' })
                )
            );
        }
    }]);

    return TiCode;
}(React.Component);

exports.default = TiCode;
module.exports = exports['default'];