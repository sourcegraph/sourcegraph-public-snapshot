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

var TiMediaPlayReverseOutline = function (_React$Component) {
    _inherits(TiMediaPlayReverseOutline, _React$Component);

    function TiMediaPlayReverseOutline() {
        _classCallCheck(this, TiMediaPlayReverseOutline);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiMediaPlayReverseOutline).apply(this, arguments));
    }

    _createClass(TiMediaPlayReverseOutline, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm23.3 11.7v16.6l-8.5-8.3 8.5-8.3m-2.3000000000000007-2.3999999999999986c-4.300000000000001 4.199999999999999-11 10.7-11 10.7s6.699999999999999 6.5 11 10.7c0.6999999999999993 0.6000000000000014 1.5 1 2.3000000000000007 1 1.8999999999999986 0 3.3999999999999986-1.5 3.3999999999999986-3.3999999999999986v-16.6c0-1.9000000000000004-1.5-3.4000000000000004-3.3999999999999986-3.4000000000000004-0.8000000000000007 0-1.6000000000000014 0.40000000000000036-2.3000000000000007 1z' })
                )
            );
        }
    }]);

    return TiMediaPlayReverseOutline;
}(React.Component);

exports.default = TiMediaPlayReverseOutline;
module.exports = exports['default'];