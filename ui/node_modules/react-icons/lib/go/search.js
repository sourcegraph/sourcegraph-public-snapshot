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

var GoSearch = function (_React$Component) {
    _inherits(GoSearch, _React$Component);

    function GoSearch() {
        _classCallCheck(this, GoSearch);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(GoSearch).apply(this, arguments));
    }

    _createClass(GoSearch, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm38.75 32.5l-9.73125-9.73125c1.3874999999999993-2.2687500000000007 2.2312499999999993-4.912500000000001 2.2312499999999993-7.768750000000001 0-8.2825-6.71875-15-15-15s-15 6.717499999999999-15 15c0 8.28125 6.717499999999999 15 15 15 2.8575000000000017 0 5.5-0.8425000000000011 7.768750000000001-2.2250000000000014l9.73125 9.725000000000001c0.6837500000000034 0.6837500000000034 1.8162499999999966 0.6787500000000009 2.5 0l2.5-2.5c0.6837500000000034-0.6837500000000034 0.6837500000000034-1.8162499999999966 0-2.5z m-22.5-7.5c-5.522500000000001 0-10-4.477499999999999-10-10s4.477499999999999-10 10-10 10 4.477499999999999 10 10-4.477499999999999 10-10 10z' })
                )
            );
        }
    }]);

    return GoSearch;
}(React.Component);

exports.default = GoSearch;
module.exports = exports['default'];