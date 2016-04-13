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

var TiCalendar = function (_React$Component) {
    _inherits(TiCalendar, _React$Component);

    function TiCalendar() {
        _classCallCheck(this, TiCalendar);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiCalendar).apply(this, arguments));
    }

    _createClass(TiCalendar, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.7 10.3v-0.3000000000000007c0-2.8-2.3000000000000007-5-5-5s-5 2.2-5 5h-3.3999999999999986c0-2.8-2.1999999999999993-5-5-5s-5 2.2-5 5v0.3000000000000007c-1.9000000000000004 0.6999999999999993-3.3000000000000007 2.5-3.3000000000000007 4.699999999999999v15c0 2.799999999999997 2.2 5 5 5h20c2.799999999999997 0 5-2.200000000000003 5-5v-15c0-2.1999999999999993-1.3999999999999986-4-3.3000000000000007-4.699999999999999z m-6.699999999999999-0.3000000000000007c0-0.9000000000000004 0.6999999999999993-1.6999999999999993 1.6999999999999993-1.6999999999999993s1.6000000000000014 0.8000000000000007 1.6000000000000014 1.6999999999999993v3.3000000000000007c0 1-0.6999999999999993 1.6999999999999993-1.6000000000000014 1.6999999999999993s-1.6999999999999993-0.6999999999999993-1.6999999999999993-1.6999999999999993v-3.3000000000000007z m-13.3 0c0-0.9000000000000004 0.6999999999999993-1.6999999999999993 1.5999999999999996-1.6999999999999993s1.700000000000001 0.7999999999999989 1.700000000000001 1.6999999999999993v3.3000000000000007c0 1-0.6999999999999993 1.6999999999999993-1.6999999999999993 1.6999999999999993s-1.5999999999999996-0.6999999999999993-1.5999999999999996-1.6999999999999993v-3.3000000000000007z m20 20c0 0.8999999999999986-0.8000000000000007 1.6999999999999993-1.6999999999999993 1.6999999999999993h-20c-0.9000000000000004 0-1.6999999999999993-0.8000000000000007-1.6999999999999993-1.6999999999999993v-10h23.4v10z' })
                )
            );
        }
    }]);

    return TiCalendar;
}(React.Component);

exports.default = TiCalendar;
module.exports = exports['default'];