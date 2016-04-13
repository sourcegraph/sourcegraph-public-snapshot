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

var TiSpiral = function (_React$Component) {
    _inherits(TiSpiral, _React$Component);

    function TiSpiral() {
        _classCallCheck(this, TiSpiral);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiSpiral).apply(this, arguments));
    }

    _createClass(TiSpiral, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 19.7c1.6999999999999993-0.6999999999999993 1.1999999999999993-3 0-4-1.8000000000000007-1.5-4.5-0.6999999999999993-5.699999999999999 1.3000000000000007-2.5 4 1.5 8.3 5.699999999999999 8.2 4.5-0.3999999999999986 7.199999999999999-4.900000000000002 6.199999999999999-9-1.1999999999999993-5-6.5-7.5-11.2-6-4.300000000000001 1.3000000000000007-7 5.800000000000001-6.699999999999999 10.3 0.5 5 4.4 9 9.2 9.8 4.699999999999999 0.8999999999999986 9.5-1.3000000000000007 12-5.300000000000001 1.1999999999999993-1.8000000000000007 2-4 2-6.300000000000001 0-0.8999999999999986 0.7999999999999972-1.6999999999999993 1.7999999999999972-1.5 1.3999999999999986 0 1.7000000000000028 1.3000000000000007 1.5 2.3000000000000007-0.6000000000000014 7.800000000000001-7.5 14.299999999999997-15.5 14.299999999999997-9.799999999999997 0-17.499999999999996-10.299999999999997-13.299999999999997-19.599999999999998 4.199999999999999-9 17.2-10.9 22.2-2 2.5 4.099999999999998 2 9.599999999999998-1.5 13.099999999999998-3.3999999999999986 3.3999999999999986-8.900000000000002 4-12.9 1.1999999999999993-3.5999999999999996-2.6999999999999993-4.799999999999999-8.2-1.8000000000000007-12 2.799999999999999-3.8000000000000007 9.199999999999998-4 11.699999999999998 0.3000000000000007 1.8000000000000007 3.1999999999999993 0 8.7-4.199999999999999 8.2-2.6999999999999993 0-5-2.8000000000000007-3.5-5.300000000000001 1-1.5 3.1999999999999993-1 3.8000000000000007 0.10000000000000142 0.3999999999999986 1.3999999999999986 0.1999999999999993 1.8999999999999986 0.1999999999999993 1.8999999999999986z' })
                )
            );
        }
    }]);

    return TiSpiral;
}(React.Component);

exports.default = TiSpiral;
module.exports = exports['default'];