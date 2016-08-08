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

var TiArrowUpThick = function (_React$Component) {
    _inherits(TiArrowUpThick, _React$Component);

    function TiArrowUpThick() {
        _classCallCheck(this, TiArrowUpThick);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiArrowUpThick).apply(this, arguments));
    }

    _createClass(TiArrowUpThick, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 5.286666666666667l-10.69 10.690000000000001c-1.3000000000000007 1.299999999999999-1.3000000000000007 3.411666666666667 0 4.713333333333333s3.411666666666667 1.3000000000000007 4.713333333333333 0l2.6433333333333344-2.6433333333333344v12.07c0 1.8399999999999999 1.4916666666666671 3.3333333333333357 3.333333333333332 3.3333333333333357s3.333333333333332-1.4933333333333323 3.333333333333332-3.333333333333332v-12.07l2.6433333333333344 2.6433333333333344c0.6499999999999986 0.6499999999999986 1.5033333333333339 0.9766666666666666 2.3566666666666656 0.9766666666666666s1.7049999999999983-0.3249999999999993 2.3566666666666656-0.9766666666666666c1.3000000000000007-1.3000000000000007 1.3000000000000007-3.411666666666669 0-4.713333333333333l-10.689999999999998-10.690000000000005z' })
                )
            );
        }
    }]);

    return TiArrowUpThick;
}(React.Component);

exports.default = TiArrowUpThick;
module.exports = exports['default'];