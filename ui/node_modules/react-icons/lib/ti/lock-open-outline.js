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

var TiLockOpenOutline = function (_React$Component) {
    _inherits(TiLockOpenOutline, _React$Component);

    function TiLockOpenOutline() {
        _classCallCheck(this, TiLockOpenOutline);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiLockOpenOutline).apply(this, arguments));
    }

    _createClass(TiLockOpenOutline, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm22.166666666666668 28.333333333333336c0 1.1966666666666654-0.9699999999999989 2.166666666666668-2.166666666666668 2.166666666666668s-2.166666666666668-0.9699999999999989-2.166666666666668-2.166666666666668c0-1.1966666666666654 0.9699999999999989-2.166666666666668 2.166666666666668-2.166666666666668s2.166666666666668 0.9699999999999989 2.166666666666668 2.166666666666668z m7.833333333333332-21.666666666666668c-3.676666666666666-8.881784197001252e-16-6.666666666666668 2.9899999999999984-6.666666666666668 6.666666666666666v5.000000000000002h-6.666666666666668v-1.6666666666666679h-4.9999999999999964c-1.8399999999999999 0-3.333333333333334 1.4933333333333323-3.333333333333334 3.333333333333332v11.666666666666668c0 1.8399999999999999 1.493333333333334 3.333333333333332 3.333333333333334 3.333333333333332h16.666666666666668c1.8399999999999999 0 3.333333333333332-1.4933333333333323 3.333333333333332-3.333333333333332v-11.666666666666668c0-1.8399999999999999-1.4933333333333323-3.333333333333332-3.333333333333332-3.333333333333332h-1.6666666666666679v-3.333333333333334c0-1.8399999999999999 1.4933333333333323-3.333333333333334 3.333333333333332-3.333333333333334s3.3333333333333357 1.493333333333334 3.3333333333333357 3.333333333333334v5.000000000000002c0 0.9200000000000017 0.7466666666666697 1.6666666666666679 1.6666666666666643 1.6666666666666679s1.6666666666666643-0.7466666666666661 1.6666666666666643-1.6666666666666679v-5.000000000000002c0-3.6766666666666676-2.990000000000002-6.666666666666667-6.666666666666668-6.666666666666667z m-1.6666666666666679 25h-16.666666666666664v-11.666666666666668h16.671666666666667l-0.004999999999999005 11.666666666666668z' })
                )
            );
        }
    }]);

    return TiLockOpenOutline;
}(React.Component);

exports.default = TiLockOpenOutline;
module.exports = exports['default'];