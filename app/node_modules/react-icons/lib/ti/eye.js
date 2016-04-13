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

var TiEye = function (_React$Component) {
    _inherits(TiEye, _React$Component);

    function TiEye() {
        _classCallCheck(this, TiEye);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiEye).apply(this, arguments));
    }

    _createClass(TiEye, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm36.36666666666667 20.71666666666667c-0.13666666666666316-0.1999999999999993-3.43333333333333-4.906666666666666-7.986666666666665-8.125-2.360000000000003-1.6716666666666686-5.336666666666666-2.5916666666666686-8.380000000000003-2.5916666666666686s-6.016666666666666 0.9199999999999999-8.383333333333335 2.591666666666667c-4.55 3.216666666666667-7.846666666666668 7.926666666666668-7.983333333333333 8.125000000000002-0.3999999999999999 0.5716666666666654-0.3999999999999999 1.3299999999999983 0 1.8999999999999986 0.13666666666666671 0.1999999999999993 3.4333333333333336 4.908333333333331 7.983333333333333 8.126666666666665 2.366666666666669 1.6700000000000017 5.34166666666667 2.5900000000000034 8.383333333333335 2.5900000000000034 3.043333333333333 0 6.02-0.9200000000000017 8.379999999999999-2.5916666666666686 4.550000000000001-3.2166666666666686 7.849999999999998-7.926666666666669 7.988333333333333-8.125 0.3999999999999986-0.5700000000000003 0.3999999999999986-1.3299999999999983 0-1.8999999999999986z m-16.366666666666667 6.783333333333331c-3.2233333333333327 0-5.833333333333334-2.616666666666667-5.833333333333334-5.833333333333336 0-3.2233333333333327 2.610000000000001-5.833333333333334 5.833333333333334-5.833333333333334 3.2166666666666686 0 5.833333333333336 2.610000000000001 5.833333333333336 5.833333333333334 0 3.2166666666666686-2.616666666666667 5.833333333333336-5.833333333333336 5.833333333333336z m3.3333333333333357-5.833333333333332c0 1.836666666666666-1.4966666666666661 3.333333333333332-3.333333333333332 3.333333333333332-1.841666666666665 0-3.333333333333332-1.4966666666666661-3.333333333333332-3.333333333333332 0-1.841666666666665 1.4916666666666671-3.333333333333332 3.333333333333332-3.333333333333332 1.836666666666666 0 3.333333333333332 1.4916666666666671 3.333333333333332 3.333333333333332z' })
                )
            );
        }
    }]);

    return TiEye;
}(React.Component);

exports.default = TiEye;
module.exports = exports['default'];