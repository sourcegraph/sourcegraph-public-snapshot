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

var TiArrowLeftThick = function (_React$Component) {
    _inherits(TiArrowLeftThick, _React$Component);

    function TiArrowLeftThick() {
        _classCallCheck(this, TiArrowLeftThick);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiArrowLeftThick).apply(this, arguments));
    }

    _createClass(TiArrowLeftThick, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30 18.333333333333336h-12.073333333333334l2.6433333333333344-2.6433333333333344c1.3000000000000007-1.3000000000000007 1.3000000000000007-3.416666666666666 0-4.713333333333333-1.3000000000000007-1.3000000000000007-3.411666666666669-1.3000000000000007-4.713333333333333 0l-10.690000000000001 10.69 10.69 10.690000000000001c0.6499999999999986 0.6499999999999986 1.5033333333333339 0.9766666666666666 2.3566666666666656 0.9766666666666666s1.7049999999999983-0.32500000000000284 2.3566666666666656-0.9766666666666666c1.3000000000000007-1.3000000000000007 1.3000000000000007-3.416666666666668 0-4.713333333333331l-2.6433333333333273-2.643333333333338h12.07333333333333c1.8399999999999999 0 3.3333333333333357-1.4933333333333323 3.3333333333333357-3.333333333333332 0-1.841666666666665-1.4933333333333323-3.333333333333332-3.333333333333332-3.333333333333332z' })
                )
            );
        }
    }]);

    return TiArrowLeftThick;
}(React.Component);

exports.default = TiArrowLeftThick;
module.exports = exports['default'];