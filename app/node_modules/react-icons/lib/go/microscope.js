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

var GoMicroscope = function (_React$Component) {
    _inherits(GoMicroscope, _React$Component);

    function GoMicroscope() {
        _classCallCheck(this, GoMicroscope);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(GoMicroscope).apply(this, arguments));
    }

    _createClass(GoMicroscope, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm29.1 35c3.3725000000000023-0.7325000000000017 5.899999999999999-3.90625 5.899999999999999-7.5 0-2.282499999999999-1.0399999999999991-4.300000000000001-2.6499999999999986-5.678750000000001 0.09125000000000227-0.5962500000000013 0.14999999999999858-1.1999999999999993 0.14999999999999858-1.8212499999999991 0-4.094999999999999-1.9499999999999993-7.73-5-10l2.5-2.5v-2.5l2.5-2.5-2.5-2.5-2.5 2.5h-2.5l-10 10-5 2.5v5l2.5 2.5h5l2.5-5 3.75-3.75c2.1675000000000004 1.3049999999999997 3.75 3.532499999999999 3.75 6.25-4.142499999999998 0-7.5 3.3575000000000017-7.5 7.5h-15v2.5h7.5c0.7475000000000005 0.5562499999999986 1.6425 0.8625000000000007 2.5 1.25v3.75h-5l-5 5h30l-5-5h-0.8999999999999986z m-4.100000000000001-7.5c0-1.3812500000000014 1.1187499999999986-2.5 2.5-2.5s2.5 1.1187499999999986 2.5 2.5c0 1.379999999999999-1.1187499999999986 2.5-2.5 2.5s-2.5-1.120000000000001-2.5-2.5z' })
                )
            );
        }
    }]);

    return GoMicroscope;
}(React.Component);

exports.default = GoMicroscope;
module.exports = exports['default'];