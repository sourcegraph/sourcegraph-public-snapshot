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

var TiChartBar = function (_React$Component) {
    _inherits(TiChartBar, _React$Component);

    function TiChartBar() {
        _classCallCheck(this, TiChartBar);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiChartBar).apply(this, arguments));
    }

    _createClass(TiChartBar, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm23.333333333333336 6.666666666666667c0-1.8416666666666668-1.4933333333333323-3.3333333333333335-3.333333333333332-3.3333333333333335s-3.333333333333332 1.4916666666666667-3.333333333333332 3.3333333333333335v20h6.666666666666668v-20z m8.333333333333332 6.666666666666667c0-1.8416666666666668-1.4933333333333323-3.333333333333334-3.333333333333332-3.333333333333334s-3.333333333333332 1.4916666666666671-3.333333333333332 3.333333333333334v13.333333333333334h6.666666666666668v-13.333333333333334z m-16.666666666666668 5.000000000000002c0-1.8416666666666686-1.493333333333334-3.333333333333334-3.333333333333334-3.333333333333334s-3.333333333333334 1.4916666666666654-3.333333333333334 3.333333333333334v8.333333333333336h6.666666666666668v-8.333333333333336z m16.666666666666668 13.333333333333332h-23.333333333333336c-0.9216666666666651 0-1.6666666666666652 0.745000000000001-1.6666666666666652 1.6666666666666679s0.7450000000000001 1.6666666666666643 1.666666666666667 1.6666666666666643h23.333333333333336c0.9216666666666669 0 1.6666666666666643-0.7449999999999974 1.6666666666666643-1.6666666666666643s-0.7449999999999974-1.6666666666666679-1.6666666666666679-1.6666666666666679z' })
                )
            );
        }
    }]);

    return TiChartBar;
}(React.Component);

exports.default = TiChartBar;
module.exports = exports['default'];