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

var TiArrowRightThick = function (_React$Component) {
    _inherits(TiArrowRightThick, _React$Component);

    function TiArrowRightThick() {
        _classCallCheck(this, TiArrowRightThick);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiArrowRightThick).apply(this, arguments));
    }

    _createClass(TiArrowRightThick, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm17.643333333333334 10.976666666666668c-1.3000000000000007 1.3000000000000007-1.3000000000000007 3.411666666666667 0 4.713333333333333l2.6433333333333344 2.6433333333333344h-12.073333333333336c-1.8399999999999999 0-3.333333333333334 1.4916666666666671-3.333333333333334 3.333333333333332s1.4933333333333332 3.333333333333332 3.333333333333334 3.333333333333332h12.073333333333336l-2.6433333333333344 2.6433333333333344c-1.3000000000000007 1.3000000000000007-1.3000000000000007 3.4116666666666653 0 4.713333333333335 0.6499999999999986 0.6499999999999986 1.5033333333333339 0.9766666666666666 2.3566666666666656 0.9766666666666666s1.7049999999999983-0.32500000000000284 2.3566666666666656-0.9766666666666666l10.690000000000001-10.690000000000001-10.689999999999998-10.69c-1.3000000000000007-1.3000000000000007-3.411666666666669-1.3000000000000007-4.713333333333331 0z' })
                )
            );
        }
    }]);

    return TiArrowRightThick;
}(React.Component);

exports.default = TiArrowRightThick;
module.exports = exports['default'];