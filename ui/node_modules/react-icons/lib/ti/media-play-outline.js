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

var TiMediaPlayOutline = function (_React$Component) {
    _inherits(TiMediaPlayOutline, _React$Component);

    function TiMediaPlayOutline() {
        _classCallCheck(this, TiMediaPlayOutline);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiMediaPlayOutline).apply(this, arguments));
    }

    _createClass(TiMediaPlayOutline, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm14.996666666666666 11.67c0.04499999999999993 0.043333333333333 0.09333333333333371 0.08833333333333293 0.1416666666666675 0.13000000000000078l8.416666666666666 8.2-8.491666666666667 8.273333333333333-0.06333333333333258 0.060000000000002274-0.0033333333333338544-16.663333333333334z m0.0033333333333338544-3.336666666666666c-1.8399999999999999 0-3.333333333333334 1.493333333333334-3.333333333333334 3.333333333333334v16.666666666666668c0 1.8399999999999999 1.493333333333334 3.333333333333332 3.333333333333334 3.333333333333332 0.9049999999999994 0 1.7216666666666676-0.3633333333333333 2.3216666666666654-0.9466666666666654l11.01166666666667-10.720000000000002-11.006666666666668-10.721666666666668c-0.6050000000000004-0.5833333333333321-1.4216666666666686-0.9449999999999985-2.326666666666668-0.9449999999999985z' })
                )
            );
        }
    }]);

    return TiMediaPlayOutline;
}(React.Component);

exports.default = TiMediaPlayOutline;
module.exports = exports['default'];