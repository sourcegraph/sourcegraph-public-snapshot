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

var TiKey = function (_React$Component) {
    _inherits(TiKey, _React$Component);

    function TiKey() {
        _classCallCheck(this, TiKey);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiKey).apply(this, arguments));
    }

    _createClass(TiKey, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm14.166666666666668 18.333333333333336c0 1.2199999999999989 0.2766666666666673 2.373333333333335 0.75 3.416666666666668l-6.583333333333334 6.583333333333332v2.5s1.4933333333333323 2.5 3.333333333333334 2.5h3.333333333333334v-3.333333333333332h3.333333333333334v-3.333333333333332h4.166666666666668c4.603333333333335 0 8.333333333333332-3.7300000000000004 8.333333333333332-8.333333333333336s-3.7300000000000004-8.333333333333334-8.333333333333336-8.333333333333334-8.333333333333334 3.7300000000000004-8.333333333333334 8.333333333333334z m8.333333333333332 3.333333333333332c-1.8399999999999999 0-3.333333333333332-1.4933333333333323-3.333333333333332-3.333333333333332 0-1.8416666666666686 1.4933333333333323-3.336666666666666 3.333333333333332-3.336666666666666 1.8416666666666686 0 3.333333333333332 1.4933333333333323 3.333333333333332 3.336666666666666 0 1.8399999999999999-1.4916666666666671 3.333333333333332-3.333333333333332 3.333333333333332z' })
                )
            );
        }
    }]);

    return TiKey;
}(React.Component);

exports.default = TiKey;
module.exports = exports['default'];