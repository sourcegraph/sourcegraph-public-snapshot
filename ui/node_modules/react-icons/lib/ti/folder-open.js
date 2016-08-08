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

var TiFolderOpen = function (_React$Component) {
    _inherits(TiFolderOpen, _React$Component);

    function TiFolderOpen() {
        _classCallCheck(this, TiFolderOpen);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiFolderOpen).apply(this, arguments));
    }

    _createClass(TiFolderOpen, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm37.2 13.3h-4c-0.7000000000000028-2-2.5-3.3000000000000007-4.699999999999999-3.3000000000000007h-10c0-1.8000000000000007-1.5-3.3-3.3000000000000007-3.3h-6.900000000000002c-2.8 0-5 2.1000000000000005-5 4.999999999999999v16.6c0 2.8999999999999986 2.2 4.9999999999999964 5 4.9999999999999964h20c2.8999999999999986 0 5.699999999999999-2.1000000000000014 6.400000000000002-5l3.5999999999999943-13.299999999999997c0.20000000000000284-1-0.29999999999999716-1.6999999999999993-1.1000000000000014-1.6999999999999993z m-30.500000000000004 1.6999999999999993v-3.3000000000000007c8.881784197001252e-16-1 0.6000000000000005-1.6999999999999993 1.6000000000000014-1.6999999999999993h6.699999999999999c0 1.8000000000000007 1.5 3.3000000000000007 3.3000000000000007 3.3000000000000007h10c1 0 1.6999999999999993 0.6999999999999993 1.6999999999999993 1.6999999999999993h-18.5c-1 0-1.8000000000000007 0.6999999999999993-2.1999999999999993 1.6999999999999993l-2.6000000000000005 10.5v-12.2z m24.8 12.5c-0.3000000000000007 1.3000000000000007-1.8000000000000007 2.5-3.1999999999999993 2.5h-20s-0.5999999999999996-0.3000000000000007-0.3000000000000007-1.3000000000000007l3.1999999999999993-11.7c0-0.1999999999999993 0.3000000000000007-0.3000000000000007 0.5-0.3000000000000007h22.8l-3 10.8z' })
                )
            );
        }
    }]);

    return TiFolderOpen;
}(React.Component);

exports.default = TiFolderOpen;
module.exports = exports['default'];