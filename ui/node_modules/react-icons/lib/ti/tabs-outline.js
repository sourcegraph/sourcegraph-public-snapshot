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

var TiTabsOutline = function (_React$Component) {
    _inherits(TiTabsOutline, _React$Component);

    function TiTabsOutline() {
        _classCallCheck(this, TiTabsOutline);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiTabsOutline).apply(this, arguments));
    }

    _createClass(TiTabsOutline, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30 6.666666666666667h-16.666666666666664c-1.8400000000000016 0-3.3333333333333357 1.4933333333333332-3.3333333333333357 3.333333333333333v3.333333333333334h-1.666666666666666c-1.8399999999999999 0-3.333333333333334 1.493333333333334-3.333333333333334 3.333333333333334v15c0 1.8399999999999999 1.4933333333333332 3.333333333333332 3.333333333333334 3.333333333333332h15.000000000000002c1.8399999999999999 0 3.333333333333332-1.4933333333333323 3.333333333333332-3.333333333333332v-1.6666666666666679h3.333333333333332c1.8399999999999999 0 3.3333333333333357-1.4933333333333323 3.3333333333333357-3.333333333333332v-16.666666666666668c0-1.8399999999999999-1.4933333333333323-3.333333333333334-3.333333333333332-3.333333333333334z m-21.666666666666664 25v-15h14.166666666666664c0.45833333333333215 0 0.8333333333333321 0.375 0.8333333333333321 0.8333333333333321v14.166666666666668h-14.999999999999998z m21.666666666666664-5h-5v-9.166666666666668c0-1.3783333333333339-1.1216666666666661-2.5-2.5-2.5h-9.166666666666666v-5h16.666666666666664v16.666666666666668z' })
                )
            );
        }
    }]);

    return TiTabsOutline;
}(React.Component);

exports.default = TiTabsOutline;
module.exports = exports['default'];