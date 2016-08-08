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

var GoIssueClosed = function (_React$Component) {
    _inherits(GoIssueClosed, _React$Component);

    function GoIssueClosed() {
        _classCallCheck(this, GoIssueClosed);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(GoIssueClosed).apply(this, arguments));
    }

    _createClass(GoIssueClosed, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm27.5 12.344999999999999l-3.75 3.75 6.25 6.405000000000001 10-10-3.75-3.75-6.199999999999999 6.202500000000001-2.5500000000000007-2.6075000000000017z m-7.5 20.155c-6.905000000000001 0-12.5-5.596250000000001-12.5-12.5s5.595000000000001-12.5 12.5-12.5c3.4525000000000006 0 6.577500000000001 1.4000000000000004 8.837499999999999 3.6624999999999996l3.5375000000000014-3.535c-3.166249999999998-3.1687499999999993-7.541249999999998-5.1274999999999995-12.375-5.1274999999999995-9.665 0-17.5 7.835000000000001-17.5 17.5s7.835000000000001 17.5 17.5 17.5 17.5-7.835000000000001 17.5-17.5l-7.822499999999998 7.822499999999998c0.3499999999999943-0.42999999999999616-2.9300000000000033 4.677500000000002-9.677500000000002 4.677500000000002z m2.5-22.5h-5v12.5h5v-12.5z m-5 20h5v-5h-5v5z' })
                )
            );
        }
    }]);

    return GoIssueClosed;
}(React.Component);

exports.default = GoIssueClosed;
module.exports = exports['default'];