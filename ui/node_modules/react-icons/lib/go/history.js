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

var GoHistory = function (_React$Component) {
    _inherits(GoHistory, _React$Component);

    function GoHistory() {
        _classCallCheck(this, GoHistory);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(GoHistory).apply(this, arguments));
    }

    _createClass(GoHistory, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 2.5c-3.552500000000002 0-6.850000000000001 1.0750000000000002-9.6075 2.8925l-2.8925-2.8925v10h10l-3.4375-3.4375c1.7750000000000004-0.9624999999999986 3.7749999999999986-1.5625 5.9375-1.5625 6.903749999999999 0 12.5 5.594999999999999 12.5 12.5s-5.596250000000001 12.5-12.5 12.5c-6.904999999999999 0-12.5-5.596250000000001-12.5-12.5 0-1.7800000000000011 0.3825000000000003-3.467500000000001 1.0549999999999997-5h-3.5549999999999997v-3.8825000000000003c-1.5499999999999998 2.6125000000000007-2.5 5.625-2.5 8.8825 0 9.665 7.835000000000001 17.5 17.5 17.5s17.5-7.835000000000001 17.5-17.5-7.835000000000001-17.5-17.5-17.5z m-0.03750000000000142 29.9625l2.5375000000000014-2.4624999999999986v-7.5h5l2.5-2.5-2.5-2.5h-5l-2.5-2.5-5 5 2.5 2.5v7.5l2.4624999999999986 2.4624999999999986z' })
                )
            );
        }
    }]);

    return GoHistory;
}(React.Component);

exports.default = GoHistory;
module.exports = exports['default'];